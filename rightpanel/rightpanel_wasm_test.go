//go:build wasm

package rightpanel

import (
	"syscall/js"
	"testing"

	. "webtyp.com/dom"
	. "webtyp.com/html"
)

type stubComponent struct {
	text string
	id   string
}

func (s *stubComponent) GetID() string {
	if s.id == "" {
		s.id = "stub-" + s.text
	}
	return s.id
}
func (s *stubComponent) SetID(id string)       { s.id = id }
func (s *stubComponent) String() string        { return s.text }
func (s *stubComponent) Children() []Component { return nil }
func (s *stubComponent) Render() *Element      { return Div().Text(s.text) }

type doublePanelContainer struct {
	Element
	panelA *RightPanel
	panelB *RightPanel
}

func (c *doublePanelContainer) Render() *Element {
	return Div().Child(c.panelA, c.panelB)
}

func TestTwoRightPanels_DoNotCollide(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "test-rp-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() {
		doc.Get("body").Call("removeChild", root)
	})

	panelA := &RightPanel{
		Title: "Panel A",
		Aside: &stubComponent{text: "Aside A"},
	}
	panelB := &RightPanel{
		Title: "Panel B",
		Aside: &stubComponent{text: "Aside B"},
	}

	container := &doublePanelContainer{
		panelA: panelA,
		panelB: panelB,
	}

	if err := Render("test-rp-root", container); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Both wrappers exist in DOM as '.rp' elements
	rps := doc.Call("querySelectorAll", ".rp")
	if rps.Get("length").Int() != 2 {
		t.Fatalf("expected 2 .rp elements, got %d", rps.Get("length").Int())
	}

	elemA := rps.Call("item", 0)
	elemB := rps.Call("item", 1)

	idA := elemA.Get("id").String()
	idB := elemB.Get("id").String()

	if idA == "" || idB == "" {
		t.Errorf("expected dom-assigned IDs on rendered elements, got idA=%q idB=%q", idA, idB)
	}
	if idA == idB {
		t.Errorf("expected wrappers to have different IDs, got both %q", idA)
	}

	// Verify ShowAside on Panel A does not move Panel B's scroll position or affect B
	scrollLeftBInitial := elemB.Get("scrollLeft").Int()

	panelA.ShowAside()

	scrollLeftBAfter := elemB.Get("scrollLeft").Int()
	if scrollLeftBInitial != scrollLeftBAfter {
		t.Errorf("ShowAside on panel A moved panel B scroll position: initial %d, after %d", scrollLeftBInitial, scrollLeftBAfter)
	}
}
