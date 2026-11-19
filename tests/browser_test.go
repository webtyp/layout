//go:build wasm

package tests

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	"webtyp.com/dom"
	"webtyp.com/html"
	layoutapp "webtyp.com/layout/app"
	"webtyp.com/model"
	wtime "webtyp.com/time"
	"webtyp.com/view"
)

// screenBody is the leaf the menu offers: a component that shows a marker text.
type screenBody struct{ dom.Element }

func (s *screenBody) Render() *dom.Element { return html.Div().Text("screen body marker") }

// loginForm is the form of an entry screen a sign-in mode gives.
type loginForm struct{ dom.Element }

func (f *loginForm) Render() *dom.Element { return html.Div().Text("login form marker") }

// freshAppRoot puts an empty #app element in the page; the layout renders into it.
func freshAppRoot(t *testing.T) js.Value {
	t.Helper()
	doc := js.Global().Get("document")
	if old := doc.Call("getElementById", "app"); !old.IsNull() {
		old.Call("remove")
	}
	root := doc.Call("createElement", "div")
	root.Set("id", "app")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Call("remove") })
	return root
}

// waitForText polls the visible page text until it contains want, or fails the test after two seconds.
func waitForText(t *testing.T, want string) string {
	t.Helper()
	return waitForAnyText(t, want)
}

// waitForAnyText polls the visible page text until it contains one of wants and returns that text.
func waitForAnyText(t *testing.T, wants ...string) string {
	t.Helper()
	body := js.Global().Get("document").Get("body")
	deadline := time.Now().Add(2 * time.Second)
	for {
		text := body.Get("innerText").String()
		for _, want := range wants {
			if strings.Contains(text, want) {
				return text
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("page shows none of %q; visible text = %q", wants, text)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestLayout_IsAViewLayout(t *testing.T) {
	var _ view.Layout = layoutapp.New(layoutapp.Tools{})
}

func TestShow_DrawsTheSessionItIsGiven(t *testing.T) {
	freshAppRoot(t)
	zone, err := wtime.LoadZone("America/Santiago")
	if err != nil {
		t.Fatal(err)
	}
	s := view.Session{
		Name: "acme",
		User: view.User{ID: "u1", Name: "Ana", Grants: []model.Grant{{Resource: "demo", Actions: model.Read}}},
		Zone: zone,
		Screens: []view.Screen{{
			ID: "inventory", Label: "Inventario", Icon: "inventory", Help: "Shows the inventory.", Permission: "demo", Component: &screenBody{},
		}},
		Log: func(...any) {},
	}
	go layoutapp.New(layoutapp.Tools{}).Show(s)

	text := waitForText(t, "acme")
	if !strings.Contains(text, "Inventario") {
		t.Errorf("the screen the session gave is not in the menu; text = %q", text)
	}
}

// Nobody signed in: the layout draws the entry screens the Session brings (one per sign-in mode open at this
// door), never the chassis, and never asks the server itself.
func TestShow_NobodySignedInShowsTheEntryScreens(t *testing.T) {
	freshAppRoot(t)
	s := view.Session{
		Name:    "acme",
		Entries: []view.Entry{{Subtitle: "Sign in to continue", Form: &loginForm{}, Message: dom.NewString("")}},
		Log:     func(...any) {},
	}
	go layoutapp.New(layoutapp.Tools{}).Show(s)

	waitForText(t, "login form marker")
}

// The server did not answer who is signed in: the layout says so instead of showing the entry screens (D76).
func TestShow_ServerDownSaysSo(t *testing.T) {
	freshAppRoot(t)
	s := view.Session{
		Name:       "acme",
		ServerDown: true,
		Entries:    []view.Entry{{Subtitle: "Sign in to continue", Form: &loginForm{}, Message: dom.NewString("")}},
		Log:        func(...any) {},
	}
	go layoutapp.New(layoutapp.Tools{}).Show(s)

	text := waitForText(t, "The server did not answer")
	if strings.Contains(text, "login form marker") {
		t.Errorf("a server that did not answer must not show the entry screens; text = %q", text)
	}
}
