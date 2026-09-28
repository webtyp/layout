//go:build wasm

package crudview

import (
	"syscall/js"
	"testing"

	. "webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

var roomModel = model.Definition{
	Name: "room",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "floor_id", Type: input.Select(), NotNull: true},
	},
}

type Room struct{ Id, FloorId string }

func (r *Room) ModelName() string     { return "room" }
func (r *Room) Schema() []model.Field { return roomModel.Fields }
func (r *Room) Pointers() []any       { return []any{&r.Id, &r.FloorId} }
func (r *Room) IsNil() bool           { return r == nil }
func (r *Room) EncodeFields(w model.FieldWriter) {
	w.String("id", r.Id)
	w.String("floor_id", r.FloorId)
}
func (r *Room) DecodeFields(rd model.FieldReader) {
	r.Id, _ = rd.String("id")
	r.FloorId, _ = rd.String("floor_id")
}
func (r *Room) Item() view.Item { return view.Item{ID: r.Id, Label: r.Id} }

// TestSetOptions_FillsLiveSelect is the consumer's case: a record whose form has
// a select fed by another table (a room's floor). The choices arrive from a
// Caller after the view is on screen, and must reach the live <select>.
func TestSetOptions_FillsLiveSelect(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "cv-opts-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	v, err := New(Config{
		ParentID:  "cv-opts",
		Presenter: view.New(&conformance.FakeLister{}, &Room{}),
		IDs:       testIDs,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&mockCtxWasm{})
	if err := Render("cv-opts-root", v); err != nil {
		t.Fatalf("Render: %v", err)
	}

	v.SetOptions("floor_id",
		fmt.KeyValue{Key: "f1", Value: "Piso 1"},
		fmt.KeyValue{Key: "f2", Value: "Piso 2"},
	)

	sel := doc.Call("querySelector", "#cv-opts-root select[name='floor_id']")
	if sel.IsNull() || sel.IsUndefined() {
		t.Fatal("select floor_id not rendered")
	}
	if n := sel.Get("options").Get("length").Int(); n != 2 {
		t.Fatalf("live <select> has %d options, want 2", n)
	}
}
