//go:build !wasm

package rightpanel

import (
	"strings"
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/html"
)

// ruleBlock returns the declaration block of the LAST rule whose selector
// contains want, or "" when absent. The primitives layer is emitted before the
// widgets layer and groups edge cases (e.g. ".rp, .rp__header, .rp__title {
// margin: 0; border-radius: 0; }"), so the first match can be a legitimate
// collective primitive — the intent lives in the widgets-layer rule.
func ruleBlock(cssStr, want string) string {
	i := strings.LastIndex(cssStr, want)
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

// TestRightPanel_EdgeAsserts guards two things: the panel is welded to the
// application frame and must be square, and the title is inherited ink on that
// Primary panel — NOT a surface of its own. It carries no background and no
// radius: the panel behind it is the colour, and a fill here would re-origin a
// gradient Primary theme across the heading's own box (a detached rectangle).
// This supersedes the earlier "PLAN v0.2.0 item 5" contract that had the title
// keep an interior radius.
func TestRightPanel_EdgeAsserts(t *testing.T) {
	r := &RightPanel{
		Title:         "Test Title",
		Head:          html.Div(),
		HeadControls:  html.Div(),
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
	}

	cssStr := r.RenderCSS().String()

	// The mobile strip re-declares the root inside a media query that comes
	// after the desktop rules, so LastIndex would land there. The frame's
	// square corners are a desktop contract; assert on everything before the
	// first @media block.
	desktop := cssStr
	if i := strings.Index(cssStr, "@media"); i != -1 {
		desktop = cssStr[:i]
	}

	for _, sel := range []string{".rp {"} {
		if b := ruleBlock(desktop, sel); fmt.Contains(b, "border-radius") {
			t.Errorf("%s must be squared at the frame, block:\n%s", sel, b)
		}
	}
	if b := ruleBlock(desktop, ".rp__title {"); fmt.Contains(b, "border-radius") {
		t.Errorf(".rp__title must NOT carry a surface radius — it is inherited ink on Root's Primary panel, block:\n%s", b)
	}
	if b := ruleBlock(desktop, ".rp__title {"); fmt.Contains(b, "background") {
		t.Errorf(".rp__title must NOT carry a background — the panel behind it is the colour, block:\n%s", b)
	}
}

func TestRightPanel_StylesheetAsserts(t *testing.T) {
	r := &RightPanel{
		Title:         "Test Title",
		Head:          html.Div(),
		HeadControls:  html.Div(),
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
		AsideFooter:   html.Div(),
	}

	sheet := r.RenderSheet()
	cssStr := r.RenderCSS().String()

	// 1. Check "!important"
	if fmt.Contains(cssStr, "!important") {
		t.Error("stylesheet contains forbidden !important directive")
	}

	// 2. Check @layer order
	expectedLayers := "@layer tokens, primitives, widgets, states;"
	if !fmt.Contains(cssStr, expectedLayers) {
		t.Errorf("expected stylesheet to declare layers in order %q", expectedLayers)
	}

	// 3. Extract markup and match classes starting with "rp"
	allHTML := r.Render().String()

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
					if strings.HasPrefix(cls, "rp") {
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
		// We search for class selectors: e.g. .rp or .rp__something
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
				if strings.HasPrefix(cls, "rp") {
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
	for _, kv := range sheet.StateAttrs() {
		want := kv.Key + "='" + kv.Value + "'"
		if !fmt.Contains(allHTML, want) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", want)
		}
	}
}

// TestAsideKeepsGutterOnMobile guards the reversal of what used to be PLAN.md
// Stage B1. Flush (Pad(SpaceNone)) on mobile was tried there to make room for
// targetlist's row ⋮ menu in the sliver MasterDetail(Most) leaves visible —
// but zeroing the aside's OWN outer inset put crudview__list's Inset border
// exactly flush with the aside's own edge, with nothing left to frame against
// (the fields/article side, whose gutter was untouched, kept its border
// legible the whole time — the asymmetry is what made this visible). The
// menu's clearance is reclaimed instead from targetlist's own PartList
// padding, INSIDE the card, which does not touch this outer edge — see
// components/targetlist/css.go's mobile PartList rule.
func TestAsideKeepsGutterOnMobile(t *testing.T) {
	r := &RightPanel{
		Title:         "Test Title",
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
		AsideFooter:   html.Div(),
	}

	cssStr := r.RenderCSS().String()

	// The base/desktop widgets-layer rule carries the gutter padding.
	desktop := cssStr
	if i := strings.Index(cssStr, "@media"); i != -1 {
		desktop = cssStr[:i]
	}
	base := ruleBlock(desktop, ".rp__aside {")
	if base == "" {
		t.Fatal("expected a base rule for .rp__aside")
	}
	if !strings.Contains(base, "padding: var(--space-1") {
		t.Errorf("expected .rp__aside to carry its gutter padding, block:\n%s", base)
	}
	if !strings.Contains(base, "--gap:") {
		t.Errorf("expected .rp__aside to keep its Stack(gutter) gap, block:\n%s", base)
	}

	// No mobile override should zero it back out.
	mediaIdx := strings.Index(cssStr, "@media (max-width")
	if mediaIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	nextMediaIdx := strings.Index(cssStr[mediaIdx+1:], "@media")
	mobileRegion := cssStr[mediaIdx:]
	if nextMediaIdx != -1 {
		mobileRegion = cssStr[mediaIdx : mediaIdx+1+nextMediaIdx]
	}
	if mobile := ruleBlock(mobileRegion, ".rp__aside {"); strings.Contains(mobile, "padding: 0") {
		t.Errorf(".rp__aside must NOT be zeroed back to flush on mobile, block:\n%s", mobile)
	}
}

// TestFooterKeepsOffFrameOnMobile pins the floating-chrome criterion:
// the docked footer keeps Space3, which plus the aside's own Space1 gutter
// lands it exactly Space4 off the frame — the same Space4 the hamburger's
// msg-stack keeps. Flush (SpaceNone) strands tap targets at the edge.
func TestFooterKeepsOffFrameOnMobile(t *testing.T) {
	r := &RightPanel{
		Title:         "Test Title",
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
		AsideFooter:   html.Div(),
	}

	cssStr := r.RenderCSS().String()
	mediaIdx := strings.Index(cssStr, "@media (max-width")
	if mediaIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	mobileRegion := cssStr[mediaIdx:]
	if next := strings.Index(mobileRegion[1:], "@media"); next != -1 {
		mobileRegion = mobileRegion[:next+1]
	}
	mb := ruleBlock(mobileRegion, ".rp__aside-footer {")
	if mb == "" {
		t.Fatal("expected a mobile rule for .rp__aside-footer")
	}
	for _, want := range []string{"inset-block-end: var(--space-3", "inset-inline-end: var(--space-3"} {
		if !strings.Contains(mb, want) {
			t.Errorf(".rp__aside-footer should keep %q off the frame, block:\n%s", want, mb)
		}
	}
}

func TestRightPanel_FlowIsSplitRootStackedMain(t *testing.T) {
	r := &RightPanel{
		Title:         "Test Title",
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
	}

	cssStr := r.RenderCSS().String()

	// The root splits the two columns; main stacks the title band above the body.
	// These were swapped, and the swap was invisible because no demo module
	// passed an Aside. ruleBlock's LastIndex would land on the mobile strip's
	// re-flow of the root, so the desktop flow is asserted on the exact rule
	// the style DSL emits for the Split/Stack pair. Main's stack is part of the
	// primitives-layer compound selector it shares with the aside bands.
	if !fmt.Contains(cssStr, ".rp {\n  display: flex;\n  flex-wrap: wrap") {
		t.Errorf(".rp must carry the Split flow")
	}
	if !fmt.Contains(cssStr, ".rp__aside, .rp__aside-content, .rp__main {\n  display: flex;\n  flex-direction: column") {
		t.Errorf(".rp__main must stack, not split")
	}
}
