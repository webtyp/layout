//go:build wasm

package crudview

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

func TestCrudView_Wasm_Flow(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "1", Name: "Item One", Ip: "Desc One"},
			&Device{Id: "2", Name: "Item Two", Ip: "Desc Two"},
		},
	}
	p := view.New(fb, &Device{}, view.WithTitle("Wasm Test"))

	v := &CrudView{
		Title:     "Wasm Test",
		Presenter: p,
	}
	v.Init(&mockCtxWasm{})
	v.Reload(nil)

	if len(v.Presenter.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.Presenter.Items()))
	}

	// Test filtering
	v.search.Set("one")
	v.filter()
	if v.list.Count() != 1 {
		t.Errorf("expected 1 filtered item, got %d", v.list.Count())
	}

	// Test selection
	v.selectAction(view.Item{ID: "1"})
	if v.selected.Get() != "1" {
		t.Error("expected item 1 to be selected")
	}
	if !v.canDelete.Get() {
		t.Error("expected canDelete to be true")
	}

	v.selectAction(view.Item{ID: ""})
	if v.canDelete.Get() {
		t.Error("expected canDelete to be false")
	}
}

type mockCtxWasm struct{}

func (m *mockCtxWasm) OnCleanup(fn func()) {}
