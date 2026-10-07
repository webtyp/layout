//go:build wasm

package crudview

import (
	"syscall/js"
	"testing"

	. "webtyp.com/dom"

	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// mountActions renders a crudview over fl into a fresh root #id and returns
// the view and the document.
func mountActions(t *testing.T, id string, fl *conformance.FakeLister) (*CrudView, js.Value) {
	t.Helper()
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", id)
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	v, err := New(Config{ParentID: id + "-cv", Presenter: view.New(fl, &Device{}), IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&mockCtxWasm{})
	if err := Render(id, v); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return v, doc
}

// The live button: rendered with its label, disabled until the presenter has
// items, enabled after a reload — and it stays enabled when a filter term hides
// every row (the action runs over the presenter's items, not the filtered list).
func TestActions_LiveButtonFollowsPresenterItems(t *testing.T) {
	fl := fakeListBackend()
	fl.ActionList = []view.Action{{Op: "apply", Label: "Apply", Confirm: "Apply these changes?"}}
	v, doc := mountActions(t, "cv-actions-live", fl)

	btn := doc.Call("querySelector", "#cv-actions-live button[name='cv-action-apply']")
	if btn.IsNull() || btn.IsUndefined() {
		t.Fatal("action button not rendered")
	}
	if got := btn.Get("textContent").String(); got != "Apply" {
		t.Errorf("button text = %q, want Apply", got)
	}
	if !btn.Get("disabled").Bool() {
		t.Error("button enabled before any items were loaded")
	}

	v.Reload(nil)
	if btn.Get("disabled").Bool() {
		t.Error("button still disabled after items loaded")
	}

	v.search.Set("no-row-matches-this")
	v.filter()
	if btn.Get("disabled").Bool() {
		t.Error("a filter hiding every row disabled the action")
	}

	var reported int
	v.OnAction = func(op string, err error) { reported++ }
	btn.Call("click")
	if len(fl.Ran) != 0 {
		t.Fatalf("an action with Confirm ran on click: %v", fl.Ran)
	}
	v.confirmActionRun()
	if len(fl.Ran) != 1 || reported != 1 {
		t.Errorf("after confirm: Ran=%v reported=%d, want [apply] and 1", fl.Ran, reported)
	}
}

func TestActions_NoBandInDOMWithoutActions(t *testing.T) {
	_, doc := mountActions(t, "cv-actions-none", fakeListBackend())
	if el := doc.Call("querySelector", "#cv-actions-none ."+string(clsActions)); !el.IsNull() {
		t.Error("an actions band was rendered for a presenter without actions")
	}
}
