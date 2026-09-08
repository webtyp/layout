//go:build wasm

package crudview

import (
	"syscall/js"
	"testing"

	. "webtyp.com/dom"
	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// TestRowClick_WritesDataSelected is the observable the whole plan exists for.
// Before the fix, clicking a row wrote data-selected="" (HTML boolean form via
// BindAttrBool) while the sheet selects data-selected="true", so the highlight
// never applied — with no error anywhere. The value is asserted, never merely
// the attribute's presence: a presence test passes against the bug.
func TestRowClick_WritesDataSelected(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "cv-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "1", Name: "Item 1", Ip: "Desc 1"},
			&Device{Id: "2", Name: "Item 2", Ip: "Desc 2"},
		},
	}
	p := view.New(fb, &Device{})

	v := &CrudView{Title: "Wasm Test", Presenter: p}
	v.Init(&mockCtxWasm{})
	v.SetID("cv")
	_ = v.Reload()
	if err := Render("cv-root", v); err != nil {
		t.Fatalf("Render: %v", err)
	}

	row1 := doc.Call("querySelector", "[data-row='tl-1']")
	if row1.IsNull() || row1.IsUndefined() {
		t.Fatal("row [data-row='tl-1'] not mounted")
	}
	row2 := doc.Call("querySelector", "[data-row='tl-2']")
	if row2.IsNull() || row2.IsUndefined() {
		t.Fatal("row [data-row='tl-2'] not mounted")
	}

	if got := row1.Call("getAttribute", "data-selected").String(); got != "<null>" {
		t.Fatalf("initial: row 1 data-selected want absent, got %q", got)
	}

	row1.Call("click")

	if got := row1.Call("getAttribute", "data-selected").String(); got != "true" {
		t.Errorf("after click on row 1: data-selected want \"true\", got %q — the targetlist defect", got)
	}
	if got := row2.Call("getAttribute", "data-selected").String(); got != "<null>" {
		t.Errorf("after click on row 1: row 2 data-selected want absent, got %q", got)
	}

	row2.Call("click")

	if got := row2.Call("getAttribute", "data-selected").String(); got != "true" {
		t.Errorf("after click on row 2: data-selected want \"true\", got %q", got)
	}
	if got := row1.Call("getAttribute", "data-selected").String(); got != "<null>" {
		t.Errorf("after click on row 2: row 1 data-selected want absent, got %q", got)
	}
}
