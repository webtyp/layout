//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/layout/login"
	"webtyp.com/dom"
)

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

// Fixed UI text is English as written and translated on render; the login card
// is the one component that still carries a title of its own (a module's title
// is the platform header's <h1>).
func TestTitles_RenderEnglishAsIs(t *testing.T) {
	// Login
	v3 := &login.Login{Title: "Sign in", Message: dom.NewString("")}
	html3 := v3.Render().String()
	if !strings.Contains(html3, "Sign in") {
		t.Errorf("expected Login to render 'Sign in', got %s", html3)
	}
}
