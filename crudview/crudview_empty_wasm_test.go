//go:build wasm

package crudview

import (
	"strings"
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

func TestCrudView_Wasm_EmptyDisplays(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []view.Record{
			&Device{Id: "1", Name: "Item One", Ip: "Desc One"},
		},
	}

	p := view.New(fb, &Device{},
		view.WithTitle("Wasm Test"),
		view.WithDescription("No selection desc"),
		view.WithEmpty("No items msg"),
	)

	v := &CrudView{
		Presenter: p,
	}
	v.Init(&mockCtxWasm{})
	v.Reload(nil)

	rendered := v.Render().String()

	// The module's description is not drawn in the view: it opens from the
	// header's help (i), never as a fixed paragraph over the form.
	if strings.Contains(rendered, "No selection desc") {
		t.Errorf("Description must not be drawn inside the view. Got:\n%s", rendered)
	}
	// The quotes are single quotes around attributes in dom.Render
	if strings.Contains(rendered, `data-open='true'>No items msg</p>`) {
		t.Errorf("Expected empty message to NOT be open when list has items.")
	}
	if !strings.Contains(rendered, `>No items msg</p>`) {
		t.Errorf("Expected empty message element to exist.")
	}

	v.selectAction(model.Item{ID: "1"})
	rendered = v.Render().String()

	v.search.Set("does-not-exist")
	v.filter()
	rendered = v.Render().String()

	if !strings.Contains(rendered, `data-open='true'>No items msg</p>`) {
		t.Errorf("Expected empty message to be open when list is empty.")
	}
}

func TestCrudView_Wasm_EmptyDisplays_BothEmpty(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []view.Record{},
	}
	p := view.New(fb, &Device{}, view.WithTitle("Wasm Test"))

	v := &CrudView{
		Presenter: p,
	}
	v.Init(&mockCtxWasm{})
	v.Reload(nil)

	rendered := v.Render().String()

	if strings.Contains(rendered, "No selection desc") || strings.Contains(rendered, "No items msg") {
		t.Errorf("Expected neither element to be rendered when not provided.")
	}
}
