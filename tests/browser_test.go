//go:build wasm

package tests

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/json"
	layoutapp "webtyp.com/layout/app"
	"webtyp.com/model"
	"webtyp.com/router"
	wtime "webtyp.com/time"
	"webtyp.com/view"
)

// The operation the layout asks first: who the user is.
var opWhoAmI = router.Operation{Module: "authority", Entity: "me", Action: model.Read}

// whoAmICaller answers "who am I" with a JSON profile, or with an error when profile is empty (an anonymous visitor).
type whoAmICaller struct{ profile string }

func (c whoAmICaller) Call(op router.Operation, args model.Encodable, into model.Decodable, done func(error)) {
	if op != opWhoAmI {
		done(fmt.Err("unexpected operation: " + op.String()))
		return
	}
	if c.profile == "" {
		done(fmt.Err("not signed in"))
		return
	}
	done(json.Decode(c.profile, into))
}

func (c whoAmICaller) Dispatch(op router.Operation, args model.Encodable) {}

// screenBody is the leaf the menu offers: a component that shows a marker text.
type screenBody struct{ dom.Element }

func (s *screenBody) Render() *dom.Element { return html.Div().Text("screen body marker") }

// loginForm is the form of the entry screen a visitor sees when someone is already set up.
type loginForm struct{ dom.Element }

func (f *loginForm) Render() *dom.Element { return html.Div().Text("login form marker") }

// setupForm is the form of the first-administrator screen, shown when the server says there is no administrator yet.
type setupForm struct{ dom.Element }

func (f *setupForm) Render() *dom.Element { return html.Div().Text("setup form marker") }

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
		Brand:  view.Brand{Name: "acme"},
		Caller: whoAmICaller{profile: `{"id":"u1","name":"Ana","permissions":["demo:r"]}`},
		Zone:   zone,
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

// An anonymous visitor gets an entry screen, never the chassis. Which entry screen depends on the server's
// answer to "is there an administrator?": the layout asks it itself, so the test accepts either.
func TestShow_AnonymousSessionShowsTheEntry(t *testing.T) {
	freshAppRoot(t)
	s := view.Session{
		Brand:  view.Brand{Name: "acme"},
		Caller: whoAmICaller{},
		Entry: view.EntryScreens{
			Login: view.Entry{Subtitle: "Sign in to continue", Form: &loginForm{}, Message: dom.NewString("")},
			Setup: view.Entry{Subtitle: "Create the administrator", Form: &setupForm{}, Message: dom.NewString("")},
		},
		Log: func(...any) {},
	}
	go layoutapp.New(layoutapp.Tools{}).Show(s)

	waitForAnyText(t, "login form marker", "setup form marker")
}
