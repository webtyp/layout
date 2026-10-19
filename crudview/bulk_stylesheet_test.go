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

func TestActionGrowsInsteadOfFillingTheRow(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()
	b := ruleBlock(cssStr, ".crudview__action {")

	if fmt.Contains(b, "width: 100%") || fmt.Contains(b, "width:100%") {
		t.Errorf(".crudview__action must grow instead of filling the row (width: 100%%), block:\n%s", b)
	}
	if !fmt.Contains(cssStr, "flex-grow") {
		t.Errorf("css must carry flex-grow (from Grow())")
	}
}

func TestFooterButtonsShareTheControlHeight(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	for _, sel := range []string{".crudview__action {", ".crudview__action-delete {", ".crudview__action-edit {"} {
		b := ruleBlock(cssStr, sel)
		if b == "" {
			t.Fatalf("expected rule block for %s", sel)
		}
		if !fmt.Contains(b, "min-height: var(--control-height") {
			t.Errorf("footer button %s must be sized by --control-height, block:\n%s", sel, b)
		}
	}
}

// The delete button leans red with the rows it will act on, and only there:
// the rule selects Within the footer's Open state, which holds exactly while
// deleting — normal mode stays Primary blue. Solid Danger, not a wash: it is
// itself a destructive commit surface and carries the white glyph
// (--color-on-danger), the same fill the checked row marks wear.
func TestDeleteButtonTurnsRedOnlyWhileDeleting(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	sel := `.crudview__footer[data-open="true"] .crudview__action-delete {`
	i := strings.Index(cssStr, sel)
	if i == -1 {
		t.Fatalf("expected the footer-tone rule %s", sel)
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !fmt.Contains(body, "--color-danger") || fmt.Contains(body, "--color-danger-wash") {
		t.Errorf("the delete button must use the solid Danger surface while deleting, block:\n%s", body)
	}
	if !fmt.Contains(body, "--color-on-danger") {
		t.Errorf("the delete button must carry the white glyph (--color-on-danger) while deleting, block:\n%s", body)
	}
}

// The footer's delete/edit buttons carry both RevealedBy(Open) and an icon
// child, so their base rule ends `display:none` and the @layer states reveal
// must restore `display:flex` (not `block`) or CenterContent goes inert and
// the glyph strands at the leading edge — the toggle button, never hidden,
// stays centered and is the reference. Row(SpaceNone) is what makes the
// revealed display flex; see css.go.
func TestFooterButtonsCentreTheirGlyphWhenRevealed(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})
	v := &CrudView{
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	for _, sel := range []string{
		`.crudview__action-delete[data-open="true"] {`,
		`.crudview__action-edit[data-open="true"] {`,
	} {
		i := strings.Index(cssStr, sel)
		if i == -1 {
			t.Fatalf("missing reveal rule %s", sel)
		}
		body := cssStr[i:]
		if e := strings.Index(body, "}"); e != -1 {
			body = body[:e]
		}
		if !fmt.Contains(body, "display: flex") {
			t.Errorf("%s must reveal as display:flex (so CenterContent applies), got:\n%s", sel, body)
		}
		if fmt.Contains(body, "display: block") {
			t.Errorf("%s must not reveal as display:block (glyph strands left), got:\n%s", sel, body)
		}
	}

	for _, sel := range []string{".crudview__action-delete {", ".crudview__action-edit {"} {
		b := ruleBlock(cssStr, sel)
		if !fmt.Contains(b, "justify-content: center") {
			t.Errorf("%s lost its centring, got:\n%s", sel, b)
		}
	}
}
