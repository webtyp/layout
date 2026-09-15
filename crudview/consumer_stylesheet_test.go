//go:build !wasm

package crudview

import (
	"strings"
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// ruleBlock returns the declaration block of the first rule whose selector
// contains want, or "" when absent. The search is scoped to @layer widgets:
// the primitives layer groups shared flags (overflow, keepSize, flow) into
// rules whose selector list can also match want, and those blocks carry none
// of the per-part declarations the asserts below look for.
func ruleBlock(cssStr, want string) string {
	if i := strings.Index(cssStr, "@layer widgets {"); i != -1 {
		cssStr = cssStr[i:]
	}
	i := strings.Index(cssStr, want)
	if i == -1 {
		return ""
	}
	start := strings.Index(cssStr[i:], "{")
	if start == -1 {
		return ""
	}
	body := cssStr[i+start:]
	end := strings.Index(body, "}")
	if end == -1 {
		return ""
	}
	return body[:end]
}

// TestCrudView_ControlAndEdgeAsserts guards PLAN v0.2.0 items 5 and 6: the
// controls answer to --control-height by construction. The frame AND the form's
// card left this package — rightpanel owns the root, the columns, the strip and
// the article card (see rightpanel's EdgeAsserts and the article reconciliation
// in PLAN_STAGE_2). A rule reappearing here means someone re-hardcoded a
// skeleton into the controller.
func TestCrudView_ControlAndEdgeAsserts(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	// The action button answers to the control token.
	if b := ruleBlock(cssStr, ".crudview__action {"); !fmt.Contains(b, "min-height: var(--control-height") {
		t.Errorf(".crudview__action must be sized by --control-height, block:\n%s", b)
	}

	for _, gone := range []string{".crudview__search", ".crudview__detail",
		".crudview__aside", ".crudview__title", ".crudview__actions", ".crudview__article"} {
		if fmt.Contains(cssStr, gone) {
			t.Errorf("%s must not exist: rightpanel owns the frame and the article card", gone)
		}
	}
}

// TestListDeclaresFloatingChromeOnMobileOnly is the net for layout/docs/PLAN.md
// stage 1: the mobile action button is Docked over the list's own scrollable
// content, and the badge of a scrolled-to-end row used to land underneath it
// because nothing declared that the button occupies that strip. FloatingChrome
// crosses the boundary to targetlist's Scroll() without either package naming
// the other's class. Desktop, where the button sits in normal flow instead of
// floating, must NOT carry a --floating-bottom reservation — a scroller with
// nothing floating over it should not pad for one anyway.
func TestListDeclaresFloatingChromeOnMobileOnly(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	desktop := cssStr
	if i := strings.Index(cssStr, "@media"); i != -1 {
		desktop = cssStr[:i]
	}
	if b := ruleBlock(desktop, ".crudview__list {"); fmt.Contains(b, "--floating-bottom") {
		t.Errorf("desktop .crudview__list must not reserve floating-chrome space, block:\n%s", b)
	}

	mediaIdx := strings.Index(cssStr, "@media (max-width")
	if mediaIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	mobileRegion := cssStr[mediaIdx:]
	if next := strings.Index(mobileRegion[1:], "@media"); next != -1 {
		mobileRegion = mobileRegion[:next+1]
	}
	mb := ruleBlock(mobileRegion, ".crudview__list {")
	if mb == "" {
		t.Fatal("expected a mobile rule for .crudview__list")
	}
	// 1.5em (IconMd, matching every footer icon's own size) +
	// 2 * Space4 (matching the button's own Docked gap) — traceable to the
	// button's real construction, not an invented number.
	if !fmt.Contains(mb, "--floating-bottom: calc(1.5em + 2 * var(--space-4,1rem));") {
		t.Errorf(".crudview__list must declare its FloatingChrome reservation, block:\n%s", mb)
	}
}

// TestFooterButtonsAre50SquareOnMobile pins the single-source square on
// phones: FontSize(TextLg) + IconBox(IconLg) = 2.5em of 1.25rem — the same
// pair boxing the hamburger, the month arrows and every cap. The docked
// slot is shrink-to-fit, so without the IconBox the buttons would render
// content-sized (40px); desktop keeps the Grow bars (static parent).
func TestFooterButtonsAre50SquareOnMobile(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{Title: "CRUD", Presenter: p, Form: html.Div(), Filter: html.Div()}
	v.Init(&fakeCtx{})
	cssStr := v.RenderCSS().String()

	mediaIdx := strings.Index(cssStr, "@media (max-width")
	if mediaIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	mobileRegion := cssStr[mediaIdx:]
	if next := strings.Index(mobileRegion[1:], "@media"); next != -1 {
		mobileRegion = mobileRegion[:next+1]
	}
	for _, sel := range []string{
		".crudview__action {",
		".crudview__action-delete {",
		".crudview__action-edit {",
	} {
		mb := ruleBlock(mobileRegion, sel)
		if mb == "" {
			t.Fatalf("expected a mobile rule for %s", sel)
		}
		for _, want := range []string{"width: 2.5em", "height: 2.5em", "font-size: var(--text-lg"} {
			if !strings.Contains(mb, want) {
				t.Errorf("mobile %s should contain %q, block:\n%s", sel, want, mb)
			}
		}
	}
}

// TestNoManualFloatingReservation guards against reintroducing the exact
// antipattern layout/docs/PLAN.md replaced: a host reaching into a child's own
// padding to compensate for chrome the child cannot see, instead of declaring
// FloatingChrome and letting Scroll() reserve it through the inherited
// var(--floating-bottom, 0px) seam. Scroll()'s OWN reservation emits that
// exact var() reference; a hand-rolled PadEdge(EdgeBottom, someSpaceToken)
// would instead emit a literal "var(--space-N" — that pattern, not the
// property name itself (Scroll() legitimately emits padding-block-end every
// time it is used), is the thing to catch.
func TestNoManualFloatingReservation(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()
	if fmt.Contains(cssStr, "padding-block-end: var(--space-") {
		t.Error("found a hand-rolled bottom-edge padding reservation; declare FloatingChrome on the host instead")
	}
}

// TestListCardMatchesFormIndentOnMobile is the net for the indent budget:
// fields and list carry the same cardInset on mobile as on desktop, so the
// distance from the panel border to the content is identical in both columns
// (rightpanel 4 + cardInset 4 + targetlist's own 8 = 16px, the same 16px as
// the form column). The old mobile flush was a scoped exception buying room
// for the ⋮ menu in the master-detail sliver; the sliver no longer has to fit
// the ⋮ (see targetlist/css.go's PartList comment), so the exception is
// retired with the premise that justified it.
func TestListCardMatchesFormIndentOnMobile(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	mediaIdx := strings.Index(cssStr, "@media (max-width")
	if mediaIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	mobileRegion := cssStr[mediaIdx:]

	// The list declares its own cardInset on mobile (it used to flush to 0);
	// fields relies on its base rule's cardInset, which applies everywhere.
	if b := ruleBlock(mobileRegion, ".crudview__list {"); !fmt.Contains(b, "padding: var(--space-1") {
		t.Errorf(".crudview__list must carry cardInset (padding: var(--space-1...)) on mobile, block:\n%s", b)
	}
	// The base rules (every breakpoint) must also still carry the pad, so the
	// mobile list override cannot have silenced it. fields keeps cardInset on
	// its INLINE axis (the indent budget this test guards); its block inset is
	// the larger formGap, for the legend chips the list column does not have.
	if b := ruleBlock(cssStr, ".crudview__fields {"); !fmt.Contains(b, "padding-inline: var(--space-1") {
		t.Errorf(".crudview__fields base rule must carry cardInset on its inline axis, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".crudview__list {"); !fmt.Contains(b, "padding: var(--space-1") {
		t.Errorf(".crudview__list base rule must carry cardInset, block:\n%s", b)
	}

	// And both panels drop the card look on a phone: Bare background, no border.
	for _, sel := range []string{".crudview__list", ".crudview__fields"} {
		if b := ruleBlock(mobileRegion, sel+" {"); !fmt.Contains(b, "background-color: transparent;") {
			t.Errorf("%s must be Bare (background-color: transparent) on mobile, block:\n%s", sel, b)
		}
	}
}

func TestConsumer_StylesheetAsserts(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
		Filter:    html.Div(),
	}
	v.Init(&fakeCtx{})
	v.Reload(nil)
	v.composing.Set(true) // Ensure data-open renders as true

	cssStr := v.RenderCSS().String()

	// 1. Check "!important"
	if fmt.Contains(cssStr, "!important") {
		t.Error("stylesheet contains forbidden !important directive")
	}

	// 2. Check @layer order
	expectedLayers := "@layer tokens, primitives, widgets, states;"
	if !fmt.Contains(cssStr, expectedLayers) {
		t.Errorf("expected stylesheet to declare layers in order %q", expectedLayers)
	}

	// 3. Extract markup and match classes starting with "crudview"
	html1 := v.Render().String()

	pFull := view.New(&conformance.FakeLister{}, &Device{})
	vFull := &CrudView{Title: "Full", Form: html.Div(), Presenter: pFull}
	vFull.Init(&fakeCtx{})
	html2 := vFull.Render().String()

	// Manually render delete confirmation content because ModalDialog doesn't render it in hidden/initial state
	confirmHTML := v.renderDeleteConfirm().String()

	allHTML := html1 + " " + html2 + " " + confirmHTML

	// Extract classes from HTML
	importMatches := func() map[string]bool {
		classes := make(map[string]bool)
		// We look for class='...' attributes
		// HTML strings use single quotes for attributes in webtyp/dom.
		for i := 0; i < len(allHTML); i++ {
			if strings.HasPrefix(allHTML[i:], "class='") {
				start := i + len("class='")
				end := start
				for end < len(allHTML) && allHTML[end] != '\'' {
					end++
				}
				clsGroup := allHTML[start:end]
				// split by space in case of multiple classes
				for _, cls := range strings.Fields(clsGroup) {
					if strings.HasPrefix(cls, "crudview") {
						classes[cls] = true
					}
				}
			}
		}
		return classes
	}()

	// Extract classes from CSS
	cssClasses := func() map[string]bool {
		classes := make(map[string]bool)
		// We search for class selectors: e.g. .crudview or .crudview__something
		for i := 0; i < len(cssStr); i++ {
			if cssStr[i] == '.' {
				start := i + 1
				end := start
				for end < len(cssStr) && ((cssStr[end] >= 'a' && cssStr[end] <= 'z') ||
					(cssStr[end] >= 'A' && cssStr[end] <= 'Z') ||
					(cssStr[end] >= '0' && cssStr[end] <= '9') ||
					cssStr[end] == '-' || cssStr[end] == '_') {
					end++
				}
				cls := cssStr[start:end]
				if strings.HasPrefix(cls, "crudview") {
					classes[cls] = true
				}
			}
		}
		return classes
	}()

	// Verify all markup classes exist in CSS
	for c := range importMatches {
		if !cssClasses[c] {
			t.Errorf("class %q present in markup but missing in stylesheet", c)
		}
	}

	// Verify all CSS classes exist in markup
	for c := range cssClasses {
		if !importMatches[c] {
			t.Errorf("class %q present in stylesheet but missing in markup", c)
		}
	}

	// 4. State-attribute parity
	sheet := v.RenderSheet()
	for _, kv := range sheet.StateAttrs() {
		want := kv.Key + "='" + kv.Value + "'"
		if !fmt.Contains(allHTML, want) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", want)
		}
	}
}
