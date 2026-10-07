//go:build !wasm

package crudview

import (
	"testing"

	"webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

func TestCrudView_Render_Basic(t *testing.T) {
	v := &CrudView{
		Title: "Test CRUD",
		Form:  Div().Text("The Form"),
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "Test CRUD") {
		t.Error("expected title")
	}
	if !Contains(html, "The Form") {
		t.Error("expected form")
	}
	// No presenter → no source → rightpanel omits the aside band entirely.
	if Contains(html, "rp__aside") {
		t.Error("did not expect aside without source")
	}
}

func TestCrudView_Render_WithSource(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	v := &CrudView{
		Title:     "CRUD with List",
		Presenter: p,
		OnNew:     func() {},
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "class='rp'") {
		t.Error("expected the rightpanel skeleton")
	}
	if !Contains(html, "rp__aside") {
		t.Error("expected the aside band with a source")
	}
	if !Contains(html, "name='cv-crudtoggle'") {
		t.Error("expected the single crud toggle button (+/↺)")
	}
	// Delete/Edit no longer render as buttons here — they live in each
	// targetlist row's ⋮ menu (Editar/Eliminar).
	if Contains(html, "name='btn_cruddel'") {
		t.Error("did not expect a separate delete button")
	}
	// There is no back button: on a phone a sliver of the list stays visible
	// at the trailing edge of the swipe strip, and cancelling returns there
	// on its own (undoAction).
	if Contains(html, "name='cv-back'") {
		t.Error("did not expect a back-to-list button")
	}
	// The aside element is present.
	if !Contains(html, "class='rp__aside'") {
		t.Error("expected the aside element to be present")
	}
}

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

type stubList struct {
	items    []view.Item
	selected *dom.SignalString
}

func (s *stubList) GetID() string                  { return "" }
func (s *stubList) SetID(id string)                {}
func (s *stubList) String() string                 { return "" }
func (s *stubList) Render() *dom.Element           { return nil }
func (s *stubList) Children() []dom.Component      { return nil }
func (s *stubList) SetItems(items []view.Item)     { s.items = items }
func (s *stubList) Items() []view.Item             { return s.items }
func (s *stubList) Count() int                     { return len(s.items) }
func (s *stubList) SetSelectMode(on bool)          {}
func (s *stubList) SetDanger(on bool)              {}
func (s *stubList) CheckedIDs() []string           { return nil }
func (s *stubList) OnCheckedChange(fn func(n int)) {}

func TestOnAfterReload_FiresWithList(t *testing.T) {
	fb := fakeListBackend()
	p := view.New(fb, &Device{})

	var capturedList ListView
	sList := &stubList{}

	v := &CrudView{
		Title:     "OnAfterReload Test",
		Presenter: p,
		List: func(selected *dom.SignalString, onSelect func(view.Item)) ListView {
			sList.selected = selected
			return sList
		},
		OnAfterReload: func(list ListView) {
			capturedList = list
		},
	}
	v.Init(&mockCtx{})
	v.Activate()

	if capturedList != ListView(sList) {
		t.Errorf("expected capturedList to be sList, got %v", capturedList)
	}
}

func TestOnAfterReload_AfterSetItems(t *testing.T) {
	fb := fakeListBackend()
	p := view.New(fb, &Device{})

	var countInHook int
	sList := &stubList{}

	v := &CrudView{
		Title:     "OnAfterReload Test",
		Presenter: p,
		List: func(selected *dom.SignalString, onSelect func(view.Item)) ListView {
			sList.selected = selected
			return sList
		},
		OnAfterReload: func(list ListView) {
			countInHook = len(list.Items())
		},
	}
	v.Init(&mockCtx{})
	v.Activate()

	if countInHook != len(fb.Rows) {
		t.Errorf("expected countInHook == %d, got %d", len(fb.Rows), countInHook)
	}
}

func TestOnAfterReload_NilNoop(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	v := &CrudView{
		Title:     "OnAfterReload Nil Test",
		Presenter: p,
	}
	v.Init(&mockCtx{})

	var rerr error
	v.Reload(func(err error) { rerr = err })
	if rerr != nil {
		t.Fatalf("unexpected error on Reload: %v", rerr)
	}
}
