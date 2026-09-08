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

// The SSR bulk tests cannot mark a row: a row is checked by a DOM click and
// there is no click to dispatch without a browser. So the two assertions the
// whole plan exists for — "the batch ships in ONE call" and "only the fields
// the user touched are written" — live here, where the click is real and the
// path runs end to end through crudview's own methods.

func mountBulk(t *testing.T, withUpdate bool) (*CrudView, view.Lister, js.Value) {
	t.Helper()
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "cv-bulk-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	rows := []model.Model{
		&Device{Id: "id-1", Name: "Device id-1", Ip: "10.0.0.id-1"},
		&Device{Id: "id-2", Name: "Device id-2", Ip: "10.0.0.id-2"},
		&Device{Id: "id-3", Name: "Device id-3", Ip: "10.0.0.id-3"},
	}

	var backend view.Lister
	if withUpdate {
		backend = &conformance.FakeLister{Rows: rows}
	} else {
		backend = &listSaveDeleteBackend{rows: rows}
	}
	p := view.New(backend, &Device{})

	// Through New(), not a bare struct literal: New is what builds the real
	// form from the record's schema, and bulk edit is nothing without it.
	v, err := New(Config{ParentID: "cv-bulk-root", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&mockCtxWasm{})
	v.SetID("cvb")
	_ = v.Reload()
	if err := Render("cv-bulk-root", v); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if v.form == nil {
		t.Fatal("expected the real form from form.New")
	}
	return v, backend, doc
}

// findRow locates a targetlist row by the key it was built with. Rows carry
// data-row for exactly this: since components v0.6.19 the row's id is assigned
// by dom (Key + Ref), so a test cannot know it — data-row is the stable handle.
func findRow(t *testing.T, doc js.Value, key string) js.Value {
	t.Helper()
	row := doc.Call("querySelector", "[data-row='"+key+"']")
	if row.IsNull() || row.IsUndefined() {
		t.Fatalf("row %s not mounted", key)
	}
	return row
}

func clickRow(t *testing.T, doc js.Value, key string) {
	t.Helper()
	findRow(t, doc, key).Call("click")
}

func TestBulkDelete_ShipsEveryCheckedIDInOneCall(t *testing.T) {
	v, backend, doc := mountBulk(t, false)
	lsd := backend.(*listSaveDeleteBackend)

	v.setMode(modeDeleting)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-3")

	if got := v.list.CheckedIDs(); len(got) != 2 {
		t.Fatalf("expected 2 rows checked after two clicks, got %v", got)
	}

	v.confirmDeleteAction()

	if len(lsd.deleted) != 2 {
		t.Fatalf("expected 2 deleted IDs, got %v", lsd.deleted)
	}
	if lsd.deleted[0] != "id-1" || lsd.deleted[1] != "id-3" {
		t.Errorf("both checked ids must ship in the one call, got %v", lsd.deleted)
	}
}

func TestBulkDelete_ChecksFollowRenderOrder(t *testing.T) {
	v, _, doc := mountBulk(t, false)

	v.setMode(modeDeleting)
	// Tap the LAST row first: the order that reaches the confirmation message
	// must be the order on screen, not the order of tapping.
	clickRow(t, doc, "tl-id-3")
	clickRow(t, doc, "tl-id-1")

	got := v.list.CheckedIDs()
	if len(got) != 2 || got[0] != "id-1" || got[1] != "id-3" {
		t.Errorf("CheckedIDs must follow render order [id-1 id-3], got %v", got)
	}
}

func TestBulkEdit_WritesOnlyTheFieldsTheUserTouched(t *testing.T) {
	v, backend, doc := mountBulk(t, true)
	fb := backend.(*conformance.FakeLister)

	v.setMode(modeEditing)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-2")

	// One field out of the record's several.
	v.form.SetValues("name", "Renamed")
	v.autoSaveAction() // the form's own commit hook; must NOT save in this mode

	if n := len(fb.SavedRecords); n != 0 {
		t.Errorf("auto-save must stay suspended in bulk edit, got %d save calls", n)
	}

	v.bulkEditAction()

	if len(fb.UpdatedIDs) != 2 {
		t.Fatalf("expected 2 updated IDs, got %d (%v)", len(fb.UpdatedIDs), fb.UpdatedIDs)
	}
	if fb.UpdatedIDs[0] != "id-1" || fb.UpdatedIDs[1] != "id-2" {
		t.Errorf("both checked ids must ship, got %v", fb.UpdatedIDs)
	}
	hasField := false
	for _, f := range fb.UpdatedFields {
		if f == "name" {
			hasField = true
		}
		if f == "ip" {
			t.Errorf("an untouched field must never be written, got %v", fb.UpdatedFields)
		}
	}
	if !hasField {
		t.Errorf("the touched field must be named, got %v", fb.UpdatedFields)
	}
}

func TestBulkDelete_OnDeletedReceivesEveryID(t *testing.T) {
	v, _, doc := mountBulk(t, false)

	v.setMode(modeDeleting)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-2")

	var gotIDs []string
	var gotErr error
	called := false
	v.OnDeleted = func(ids []string, err error) {
		called = true
		gotIDs = ids
		gotErr = err
	}

	v.confirmDeleteAction()

	if !called {
		t.Fatal("expected OnDeleted to fire for the bulk delete")
	}
	if gotErr != nil {
		t.Fatalf("unexpected error: %v", gotErr)
	}
	if len(gotIDs) != 2 || gotIDs[0] != "id-1" || gotIDs[1] != "id-2" {
		t.Errorf("expected OnDeleted ids [id-1 id-2], got %v", gotIDs)
	}
}

func TestBulkEdit_ApplyIsDeadUntilSomethingIsEdited(t *testing.T) {
	v, _, doc := mountBulk(t, true)

	v.setMode(modeEditing)
	clickRow(t, doc, "tl-id-1")

	if v.hasEdits.Get() {
		t.Error("entering bulk edit with a blank form must leave nothing to apply")
	}

	v.form.SetValues("name", "Renamed")
	v.autoSaveAction()

	if !v.hasEdits.Get() {
		t.Error("touching a field must arm the apply button")
	}
}

func clickFooterButton(t *testing.T, doc js.Value, name string) {
	t.Helper()
	btn := doc.Call("querySelector", "[name='"+name+"']")
	if btn.IsNull() || btn.IsUndefined() {
		t.Fatalf("footer button %s not mounted", name)
	}
	btn.Call("click")
}

func rowAttr(t *testing.T, doc js.Value, key, attr string) string {
	t.Helper()
	v := findRow(t, doc, key).Call("getAttribute", attr)
	if v.IsNull() {
		return ""
	}
	return v.String()
}

// The same click that opens selection mode with nothing loaded confirms the
// loaded record directly: no mode change, then one single delete call.
func TestSingleDelete_ClickConfirmsLoadedRecord(t *testing.T) {
	v, backend, doc := mountBulk(t, false)
	lsd := backend.(*listSaveDeleteBackend)

	clickRow(t, doc, "tl-id-2") // normal mode: loads the record into the form
	if v.selected.Get() != "id-2" {
		t.Fatalf("expected id-2 loaded, got %q", v.selected.Get())
	}

	clickFooterButton(t, doc, "cv-cruddelete")

	if v.mode.Get() != string(modeNormal) {
		t.Fatalf("a single delete must not enter selection mode, got %q", v.mode.Get())
	}
	if v.deleteID.Get() != "id-2" {
		t.Fatalf("the confirmation must name the loaded record, got %q", v.deleteID.Get())
	}

	v.confirmDeleteAction()

	if len(lsd.deleted) != 1 || lsd.deleted[0] != "id-2" {
		t.Errorf("the single delete must ship id-2, got %v", lsd.deleted)
	}
	if v.selected.Get() != "" {
		t.Errorf("confirming must clear the selection, got %q", v.selected.Get())
	}
}

// The danger tone, live: a tapped row in delete mode carries data-invalid
// (red wash) and never data-selected; untapping drops it again.
func TestDeleteMode_TappedRowTurnsInvalidRed(t *testing.T) {
	_, _, doc := mountBulk(t, false)

	clickFooterButton(t, doc, "cv-cruddelete") // nothing loaded: enters selection

	clickRow(t, doc, "tl-id-1")
	if got := rowAttr(t, doc, "tl-id-1", "data-invalid"); got != "true" {
		t.Errorf("a tapped row in delete mode must carry data-invalid, got %q", got)
	}
	if got := rowAttr(t, doc, "tl-id-1", "data-selected"); got != "" {
		t.Errorf("a danger-marked row must not carry data-selected, got %q", got)
	}

	clickRow(t, doc, "tl-id-1")
	if got := rowAttr(t, doc, "tl-id-1", "data-invalid"); got != "" {
		t.Errorf("untapping must drop data-invalid, got %q", got)
	}
}

// Edit mode keeps the plain blue marks: red would lie about the action.
func TestEditMode_TappedRowStaysSelectedBlue(t *testing.T) {
	_, _, doc := mountBulk(t, true)

	clickFooterButton(t, doc, "cv-crudedit") // nothing loaded: enters edit mode

	clickRow(t, doc, "tl-id-1")
	if got := rowAttr(t, doc, "tl-id-1", "data-selected"); got != "true" {
		t.Errorf("a tapped row in edit mode must carry data-selected, got %q", got)
	}
	if got := rowAttr(t, doc, "tl-id-1", "data-invalid"); got != "" {
		t.Errorf("an edit-marked row must not carry data-invalid, got %q", got)
	}
}
