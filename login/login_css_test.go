//go:build !wasm

package login

import (
	"strings"
	"testing"
)

// RenderCSS vive en css.go, que es //go:build !wasm: este test tiene que
// llevar la misma etiqueta o el build de wasm no encuentra el metodo.
//
// The check is scoped to the ROOT rule (`.login {`), not the whole stylesheet:
// a part may legitimately use the brand colour (PartSubtitle is Glyph(Primary),
// brand-tinted text), and what this guards is only that the full-bleed backdrop
// stays the neutral Page surface — not a wash of --color-primary.
func TestLogin_RootUsesPageNotPrimary(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	rootBlocks := loginRuleBlocks(sheet, ".login {")
	if len(rootBlocks) == 0 {
		t.Fatalf("expected a `.login {` root rule, got:\n%s", sheet)
	}
	joined := strings.Join(rootBlocks, "\n---\n")
	if strings.Contains(joined, "--color-primary") {
		t.Errorf("login root must not paint the brand color as its backdrop, root rule:\n%s", joined)
	}
	if !strings.Contains(joined, "--color-background") {
		t.Errorf("login root must use the neutral page background, root rule:\n%s", joined)
	}
}

// The card used to carry Backdrop(Parent), which emits `position: absolute;
// inset: 0;` — that takes it out of Root's flex flow, so Root's own
// CenterContent() (display:flex; align-items:center; justify-content:center)
// had nothing left to center: the Width(Compact) card collapsed to the
// top-left corner instead of sitting in the middle of the screen. Guards
// against that regression coming back via a future Part(PartCard, ...) edit.
func TestLogin_CardStaysInFlowForRootToCenter(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	cardBlocks := loginRuleBlocks(sheet, ".login__card {")
	if len(cardBlocks) == 0 {
		t.Fatalf("expected a `.login__card {` rule, got:\n%s", sheet)
	}
	joined := strings.Join(cardBlocks, "\n---\n")
	if strings.Contains(joined, "position: absolute") {
		t.Errorf("login__card must stay a normal flex child of .login (no position: absolute) so Root's CenterContent() can center it, card rule:\n%s", joined)
	}
}

// loginRuleBlocks returns the declaration body of every rule whose selector
// line is exactly `sel` (standalone, not grouped or a __part), across layers.
func loginRuleBlocks(cssStr, sel string) []string {
	want := "\n" + sel
	var out []string
	for i := 0; ; {
		j := strings.Index(cssStr[i:], want)
		if j == -1 {
			return out
		}
		start := i + j + len(want)
		end := strings.Index(cssStr[start:], "}")
		if end == -1 {
			return out
		}
		out = append(out, cssStr[start:start+end])
		i = start + end
	}
}
