//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/layout/crudview"
	"webtyp.com/layout/rightpanel"
	"webtyp.com/layout/login"
	"webtyp.com/dom"
)

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

func TestTitles_RenderEnglishAsIs(t *testing.T) {
	// CrudView
	v1 := &crudview.CrudView{Title: "Devices"}
	v1.Init(&mockCtx{})
	html1 := v1.Render().String()
	if !strings.Contains(html1, "Devices") {
		t.Errorf("expected CrudView to render 'Devices', got %s", html1)
	}

	// RightPanel
	v2 := &rightpanel.RightPanel{Title: "Devices"}
	html2 := v2.Render().String()
	if !strings.Contains(html2, "Devices") {
		t.Errorf("expected RightPanel to render 'Devices', got %s", html2)
	}

	// Login
	v3 := &login.Login{Title: "Sign in", Message: dom.NewString("")}
	html3 := v3.Render().String()
	if !strings.Contains(html3, "Sign in") {
		t.Errorf("expected Login to render 'Sign in', got %s", html3)
	}
}
