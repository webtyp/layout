//go:build !wasm

package login

import (
	"strings"
	"testing"
)

// RenderCSS vive en css.go, que es //go:build !wasm: este test tiene que
// llevar la misma etiqueta o el build de wasm no encuentra el metodo.
//
// The login root uses the neutral Panel surface as its ambient backdrop so
// Primary stays reserved for the primary action button inside the form
// (ui-consistency: "Primary = brand and actions").
func TestLogin_RootUsesNeutralPanelBackdrop(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	rootBlocks := loginRuleBlocks(sheet, ".login {")
	if len(rootBlocks) == 0 {
		t.Fatalf("expected a `.login {` root rule, got:\n%s", sheet)
	}
	joined := strings.Join(rootBlocks, "\n---\n")
	if !strings.Contains(joined, "--color-surface") {
		t.Errorf("login root must paint the neutral Panel surface (--color-surface) as its backdrop, root rule:\n%s", joined)
	}
	if strings.Contains(joined, "--color-primary") {
		t.Errorf("login root must not paint a full-bleed Primary gradient behind the card, root rule:\n%s", joined)
	}
}

// The card sits on the white Field/Page surface (--color-background) with an
// outline and Popover elevation so it reads as an elevated card over the
// Panel backdrop.
func TestLogin_CardStaysNeutralOverBackdrop(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	cardBlocks := loginRuleBlocks(sheet, ".login__card {")
	if len(cardBlocks) == 0 {
		t.Fatalf("expected a `.login__card {` rule, got:\n%s", sheet)
	}
	joined := strings.Join(cardBlocks, "\n---\n")
	if strings.Contains(joined, "--color-primary") {
		t.Errorf("login__card must stay a neutral surface (As(Field)), not repaint itself with the brand color, card rule:\n%s", joined)
	}
	if !strings.Contains(joined, "--color-background") {
		t.Errorf("login__card must use --color-background (As(Field)), card rule:\n%s", joined)
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

func TestLogin_MessageRuleExistsInStylesheet(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	msgBlocks := loginRuleBlocks(sheet, ".login__message {")
	if len(msgBlocks) == 0 {
		t.Fatalf("expected a `.login__message {` rule in stylesheet, got:\n%s", sheet)
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
