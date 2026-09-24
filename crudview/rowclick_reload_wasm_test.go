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

// TestRowClick_SurvivesReloadWithGrowth is the consumer-shaped proof for the
// duplicate-row report: select a row, grow the list through a Reload (a
// create landing in the list, the exact trigger), and every row must keep
// its identity end to end — labels in order, highlight following the id,
// clicks loading their own record. Before the dom/components fix the growth
// render kept a stale node and dropped a fresh one, so selecting one record
// highlighted two and a third became unreachable.
func TestRowClick_SurvivesReloadWithGrowth(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "cv-reload-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "1", Name: "Item 1", Ip: "Desc 1"},
			&Device{Id: "2", Name: "Item 2", Ip: "Desc 2"},
		},
	}
	p := view.New(fb, &Device{})

	// Through New(), not a bare struct literal: only New builds the real
	// form from the record's schema, and the click→form assertions need it.
	v, err := New(Config{ParentID: "cv-reload-root", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&mockCtxWasm{})
	v.SetID("cv-reload")
	v.Reload(nil)
	if err := Render("cv-reload-root", v); err != nil {
		t.Fatalf("Render: %v", err)
	}

	find := func(key string) js.Value {
		t.Helper()
		row := doc.Call("querySelector", "#cv-reload-root [data-row='"+key+"']")
		if row.IsNull() || row.IsUndefined() {
			t.Fatalf("row %s not mounted", key)
		}
		return row
	}
	formName := func() string {
		t.Helper()
		inp := v.form.Input("name")
		if inp == nil {
			t.Fatal("form has no name input")
		}
		vals := inp.GetValues()
		if len(vals) == 0 {
			t.Fatal("name input holds no value")
		}
		return vals[0]
	}
	selected := func(key string) bool {
		t.Helper()
		return find(key).Call("getAttribute", "data-selected").String() == "true"
	}

	// Select row 1 before the growth.
	find("tl-1").Call("click")
	if got := formName(); got != "Item 1" {
		t.Fatalf("after click: form name = %q, want Item 1", got)
	}

	// A create lands: the backend gains a row, the view reloads.
	fb.Rows = append(fb.Rows, &Device{Id: "3", Name: "Item 3", Ip: "Desc 3"})
	v.Reload(nil)

	// All three rows mounted, labelled in order — no stale duplicate, none missing.
	ul := doc.Call("querySelector", "#cv-reload-root .targetlist__list")
	kids := ul.Get("children")
	if kids.Get("length").Int() != 3 {
		t.Fatalf("3 rows after growth, got %d", kids.Get("length").Int())
	}
	for i, want := range []struct{ key, label string }{
		{"tl-1", "Item 1"}, {"tl-2", "Item 2"}, {"tl-3", "Item 3"},
	} {
		row := kids.Call("item", i)
		if got := row.Call("getAttribute", "data-row").String(); got != want.key {
			t.Fatalf("position %d: row %q, want %q", i, got, want.key)
		}
		label := row.Call("querySelector", ".targetlist__label").Get("textContent").String()
		if label != want.label {
			t.Fatalf("position %d: label %q, want %q", i, label, want.label)
		}
	}

	// The selection followed its id through the reload: row 1 still the only highlight.
	if !selected("tl-1") {
		t.Error("row 1 lost its highlight across the reload")
	}
	if selected("tl-2") || selected("tl-3") {
		t.Error("a row that was never selected carries the highlight")
	}

	// Every click loads its own record — including the grown row and row 1 after it.
	find("tl-3").Call("click")
	if got := formName(); got != "Item 3" {
		t.Errorf("click on grown row: form name = %q, want Item 3", got)
	}
	find("tl-1").Call("click")
	if got := formName(); got != "Item 1" {
		t.Errorf("click on row 1 after growth: form name = %q, want Item 1", got)
	}
}
