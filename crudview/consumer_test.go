package crudview

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// fakeCtx is a simple implementation of Ctx for testing
type fakeCtx struct{}

func (f *fakeCtx) OnCleanup(fn func()) {}

// testIDGenerator is the composition-root double for Config.IDs — crudview
// never constructs its own generator, so every test injects one.
type testIDGenerator struct{ n int }

func (g *testIDGenerator) NewID() string {
	g.n++
	return "test-id-" + fmt.Convert(g.n).String()
}

var testIDs = &testIDGenerator{}

// deviceModel is a model with real widgets (input.Text() is a model.Kind)
var deviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip", Type: input.Text()},
	},
}

type Device struct {
	Id, Name, Ip string
}

func (d *Device) ModelName() string     { return "device" }
func (d *Device) Schema() []model.Field { return deviceModel.Fields }
func (d *Device) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Ip} }
func (d *Device) IsNil() bool           { return d == nil }

func (d *Device) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
	w.String("ip", d.Ip)
}

func (d *Device) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
	d.Ip, _ = r.String("ip")
}

func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

var _ model.Model = (*Device)(nil) // Verify compile-time implementation
var _ view.Itemizer = (*Device)(nil)

// listSaveDeleteBackend implements view.Lister + view.Saver +
// view.Deleter, but NOT view.Updater: the double for every test that
// asserts update UI stays absent.
type listSaveDeleteBackend struct {
	rows    []model.Model
	saved   []model.Model
	deleted []string
}

func (b *listSaveDeleteBackend) List(done func([]model.Model, error)) {
	out := make([]model.Model, len(b.rows))
	copy(out, b.rows)
	done(out, nil)
}

func (b *listSaveDeleteBackend) Save(recs []model.Model, done func(error)) {
	b.saved = append(b.saved, recs...)
	done(nil)
}

func (b *listSaveDeleteBackend) Delete(ids []string, done func(error)) {
	b.deleted = append(b.deleted, ids...)
	done(nil)
}

// deviceNoWidgetsModel is a model without widgets
var deviceNoWidgetsModel = model.Definition{
	Name: "device_no_widgets",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: model.Text(), NotNull: true},
	},
}

type DeviceNoWidgets struct {
	Id, Name string
}

func (d *DeviceNoWidgets) ModelName() string     { return "device_no_widgets" }
func (d *DeviceNoWidgets) Schema() []model.Field { return deviceNoWidgetsModel.Fields }
func (d *DeviceNoWidgets) Pointers() []any       { return []any{&d.Id, &d.Name} }
func (d *DeviceNoWidgets) IsNil() bool           { return d == nil }

func (d *DeviceNoWidgets) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
}

func (d *DeviceNoWidgets) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
}

var _ model.Model = (*DeviceNoWidgets)(nil) // Verify compile-time implementation

type fakeNoWidgetsPresenter struct {
	record model.Model
}

func (f *fakeNoWidgetsPresenter) Title() string                  { return "No Widgets" }
func (f *fakeNoWidgetsPresenter) SearchPlaceholder() string      { return "Search" }
func (f *fakeNoWidgetsPresenter) Record() model.Model            { return f.record }
func (f *fakeNoWidgetsPresenter) Items() []view.Item             { return nil }
func (f *fakeNoWidgetsPresenter) Filter(term string) []view.Item { return nil }
func (f *fakeNoWidgetsPresenter) Reload(done func(error))             { done(nil) }
func (f *fakeNoWidgetsPresenter) Selected() string               { return "" }
func (f *fakeNoWidgetsPresenter) Select(id string) model.Model   { return nil }
func (f *fakeNoWidgetsPresenter) Deselect()                      {}

// Case 1: New with a model without widgets fails
func TestConsumer_NewNoWidgets(t *testing.T) {
	p := &fakeNoWidgetsPresenter{record: &DeviceNoWidgets{}}
	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err == nil {
		t.Error("expected New with widget-less record to fail, but it succeeded")
	}
	if v != nil {
		t.Error("expected returned view to be nil on error")
	}
}

// Case 2: The list operation on wiring is ListOp (reloaded on Init)
func TestConsumer_ListOp(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if fb.Calls == 0 {
		t.Errorf("expected Presenter to be reloaded on Init")
	}
}

// Case 3: The list renders cards returned by Items
func TestConsumer_ListRendersCards(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
			&Device{Id: "23", Name: "Device Two", Ip: "192.168.1.2"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if len(v.Presenter.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.Presenter.Items()))
	}
	if v.list.Count() != 2 {
		t.Errorf("expected 2 rendered cards, got %d", v.list.Count())
	}
}

// Case 4: Selecting a card populates the form using form.LoadValues
func TestConsumer_SelectPopulatesForm(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
			&Device{Id: "23", Name: "Device Two", Ip: "192.168.1.2"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	v.selectAction(view.Item{ID: "12"})

	// Check form values. "id" is a PK — form.New hides it by default now
	// (see form.New's ShowField comment), so it is never one of f.Inputs
	// and there is nothing id-specific left to assert here directly: the
	// name/ip checks below only pass for device "12"'s actual record, which
	// is proof enough that selectAction resolved and loaded the right one.
	f := v.form
	nameInput := f.Input("name")
	ipInput := f.Input("ip")

	if len(nameInput.GetValues()) == 0 || nameInput.GetValues()[0] != "Device One" {
		t.Errorf("expected name input to be 'Device One', got %v", nameInput.GetValues())
	}
	if len(ipInput.GetValues()) == 0 || ipInput.GetValues()[0] != "192.168.1.1" {
		t.Errorf("expected ip input to be '192.168.1.1', got %v", ipInput.GetValues())
	}

	// Nil record on Fill should reset the form
	v.selectAction(view.Item{ID: "unknown"})
	if len(nameInput.GetValues()) != 0 && nameInput.GetValues()[0] != "" {
		t.Errorf("expected name input to be reset/empty, got %v", nameInput.GetValues())
	}
}

// Case 5: Save calls Save with form data, not original Record
func TestConsumer_SaveWithFormData(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	f := v.form
	f.SetValues("id", "12")
	f.SetValues("name", "New Name")
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSaved = func(err error) {
		saveDoneCalled = true
		saveErr = err
	}

	saver, ok := p.(view.Saver)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Saver")
	}
	v.saveAction(saver)

	if !saveDoneCalled {
		t.Fatal("expected OnSaved hook to be called")
	}
	if saveErr != nil {
		t.Fatalf("expected no save error, got: %v", saveErr)
	}

	if len(fb.SavedRecords) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(fb.SavedRecords))
	}
	dev, ok := fb.SavedRecords[0].(*Device)
	if !ok {
		t.Fatalf("expected *Device saved record, got %T", fb.SavedRecords[0])
	}
	if dev.Name != "New Name" {
		t.Errorf("expected saved device name 'New Name', got %q", dev.Name)
	}
	if dev.Ip != "10.0.0.1" {
		t.Errorf("expected saved device ip '10.0.0.1', got %q", dev.Ip)
	}
}

// Case 6: Save with invalid form doesn't call presenter and returns error
func TestConsumer_SaveInvalidForm(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	f := v.form
	f.SetValues("id", "12")
	f.SetValues("name", "") // empty name violates NotNull
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSaved = func(err error) {
		saveDoneCalled = true
		saveErr = err
	}

	saver, ok := p.(view.Saver)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Saver")
	}
	v.saveAction(saver)

	if !saveDoneCalled {
		t.Error("expected OnSaved hook to be called")
	}
	if saveErr == nil {
		t.Error("expected validation error, got nil")
	}
	if len(fb.SavedRecords) != 0 {
		t.Error("Save was called on fake backend but form was invalid")
	}
}

// Case 7: Delete calls Delete on presenter
func TestConsumer_DeleteSelected(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "123", Name: "Device One", Ip: "192.168.1.1"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Select the loaded device so the presenter indexes and registers selection
	v.selectAction(view.Item{ID: "123"})

	var deleteDoneCalled bool
	var deleteErr error
	var deletedIDs []string
	v.OnDeleted = func(ids []string, err error) {
		deleteDoneCalled = true
		deleteErr = err
		deletedIDs = ids
	}

	deleter, ok := p.(view.Deleter)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Deleter")
	}
	v.deleteAction(deleter, "123")

	if !deleteDoneCalled {
		t.Error("expected OnDeleted hook to be called")
	}
	if deleteErr != nil {
		t.Errorf("expected no delete error, got: %v", deleteErr)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != "123" {
		t.Errorf("expected hook deleted ids to be ['123'], got %v", deletedIDs)
	}

	if len(fb.DeletedIDs) != 1 || fb.DeletedIDs[0] != "123" {
		t.Errorf("expected backend deleted ids to be ['123'], got %v", fb.DeletedIDs)
	}
}

// Case 8: Delete can return error from presenter
func TestConsumer_DeleteNoSelection(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "non-existent", Name: "Device One", Ip: "192.168.1.1"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Select the item so it is registered
	v.selectAction(view.Item{ID: "non-existent"})

	// Set fb.Err *after* successful Init/Reload
	expectedErr := fmt.Errf("no selection")
	fb.Err = expectedErr

	var deleteDoneCalled bool
	var deleteErr error
	v.OnDeleted = func(ids []string, err error) {
		deleteDoneCalled = true
		deleteErr = err
	}

	deleter, ok := p.(view.Deleter)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Deleter")
	}
	v.deleteAction(deleter, "non-existent")

	if !deleteDoneCalled {
		t.Error("expected OnDeleted hook to be called")
	}
	if deleteErr == nil {
		t.Error("expected delete error because presenter/caller returned error, got nil")
	}
}

// Case 9: Presenter error on list is propagated to Reload caller
func TestConsumer_ListErrorPropagated(t *testing.T) {
	expectedErr := fmt.Errf("network connection failed")
	fb := &conformance.FakeLister{
		Err: expectedErr,
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	var receivedErr error
	v.Reload(func(err error) { receivedErr = err })
	if receivedErr == nil {
		t.Error("expected Reload to deliver the presenter error, but got nil")
	} else if receivedErr.Error() != expectedErr.Error() {
		t.Errorf("expected error '%s', got '%s'", expectedErr.Error(), receivedErr.Error())
	}
}

// Case 10: Search filters cards by Label and Description (case-insensitive)
func TestConsumer_SearchFiltering(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Frontend Device", Ip: "192.168.1.10"},
			&Device{Id: "23", Name: "Backend Server", Ip: "10.0.0.5"},
			&Device{Id: "34", Name: "Database Instance", Ip: "mysql-production"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})
	v.Reload(nil) // fetch and populate items

	// 1. Initial state (no search term) - expect all 3 items
	if v.list.Count() != 3 {
		t.Errorf("expected 3 items initially, got %d", v.list.Count())
	}

	// 2. Search by Label ("backend") case-insensitive - expect only 1 item
	v.search.Set("backend")
	v.filter()
	if v.list.Count() != 1 {
		t.Errorf("expected 1 item for 'backend', got %d", v.list.Count())
	}

	// 3. Search by Description ("mysql") case-insensitive - expect only 1 item
	v.search.Set("MYSQL")
	v.filter()
	if v.list.Count() != 1 {
		t.Errorf("expected 1 item for 'MYSQL', got %d", v.list.Count())
	}

	// 4. Search with no matches ("invalid-term") - expect 0 items
	v.search.Set("invalid-term")
	v.filter()
	if v.list.Count() != 0 {
		t.Errorf("expected 0 items for 'invalid-term', got %d", v.list.Count())
	}
}

// Case 11: the form lock is gone. Selecting a row loads it EDITABLE — there is
// no Save button, auto-save persists every field commit, so nothing needs
// protecting; and new/undo always leave the form editable too. The "disabled"
// gate that used to appear on selection is the regression this guards.
func TestConsumer_NoLockGating(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{ParentID: "my-id", Presenter: p, IDs: testIDs}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// New/blank form (nothing selected): editable.
	html := v.form.Render().String()
	if fmt.Contains(html, "disabled") {
		t.Error("expected the blank/new form to be editable, got disabled fields")
	}

	// Selecting an existing row: still editable — the lock no longer exists.
	v.selectAction(view.Item{ID: "12"})
	html = v.form.Render().String()
	if fmt.Contains(html, "disabled") {
		t.Error("expected selecting a row to leave the form editable (no lock), got disabled fields")
	}

	// editAction (kept for the conformance driver): also editable, always.
	v.editAction("12")
	html = v.form.Render().String()
	if fmt.Contains(html, "disabled") {
		t.Error("expected editAction to leave the form editable")
	}

	// Re-select the same row (simulating a fresh row click): still editable.
	v.selectAction(view.Item{ID: "12"})
	html = v.form.Render().String()
	if fmt.Contains(html, "disabled") {
		t.Error("expected re-selecting the row to leave the form editable")
	}

	// undoAction ("↺"): editable as ever.
	v.undoAction()
	html = v.form.Render().String()
	if fmt.Contains(html, "disabled") {
		t.Error("expected undoAction to leave the form editable")
	}
}

// Case 12: ⋮ -> Eliminar opens the confirmation modal instead of deleting
// immediately; only confirmDeleteAction (the modal's "Eliminar" button)
// actually deletes. Dismissing the modal without confirming leaves the
// record untouched.
func TestConsumer_DeleteRequiresConfirmation(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{ParentID: "my-id", Presenter: p, IDs: testIDs}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// deleteRequest (⋮ -> Eliminar) only opens the modal — no delete yet.
	v.deleteRequest("12")
	if len(fb.DeletedIDs) != 0 {
		t.Error("expected deleteRequest to only open the confirmation modal, not delete")
	}
	if v.deleteID.Get() != "12" {
		t.Errorf("expected pending delete id '12', got %q", v.deleteID.Get())
	}
	if v.deleteLabel.Get() != "Device One" {
		t.Errorf("expected pending delete label 'Device One', got %q", v.deleteLabel.Get())
	}

	// Dismissing without confirming (e.g. Cancelar/backdrop/×) must not delete.
	v.confirmDelete.Close()
	if len(fb.DeletedIDs) != 0 {
		t.Error("expected closing the modal without confirming to not delete")
	}

	// confirmDeleteAction (the modal's "Eliminar" button) performs the delete.
	v.confirmDeleteAction()
	if len(fb.DeletedIDs) != 1 {
		t.Errorf("expected exactly 1 device_delete call after confirming, got %d", len(fb.DeletedIDs))
	}
}

// Case 13: pressing "+" must flip the toggle to its "↺" (cancel) state even
// though nothing is selected — active() (not v.selected alone) is what the
// icon/toggleAction branch reads. Regression test for the reported bug: the
// button stayed on "+" after pressing it.
func TestConsumer_NewFlipsToggleActive(t *testing.T) {
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
		},
	}
	p := view.New(fb, &Device{})

	cfg := Config{ParentID: "my-id", Presenter: p, IDs: testIDs}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if v.active() {
		t.Error("expected idle state (nothing selected, nothing composing) to be inactive")
	}

	v.newAction()
	if !v.active() {
		t.Error("expected \"+\" to flip active() to true so the toggle shows cancel/undo")
	}

	// toggleAction must read active(), not v.selected alone, or it would call
	// newAction again (idempotent, but the wrong branch) instead of undoAction.
	v.toggleAction()
	if v.active() {
		t.Error("expected toggling while composing (nothing selected yet) to undo, not re-enter new")
	}

	// Selecting an existing row is the other source of active().
	v.selectAction(view.Item{ID: "12"})
	if !v.active() {
		t.Error("expected selecting a row to flip active() to true")
	}
}

// Case 14: Config.NewRecord seeds the form when "+" (newAction) is triggered.
func TestConsumer_NewRecord_SeedsForm(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
		NewRecord: func() model.Model {
			return &Device{Name: "Seeded Patient", Ip: "10.0.0.99"}
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	v.newAction()

	f := v.form
	nameInput := f.Input("name")
	ipInput := f.Input("ip")

	if len(nameInput.GetValues()) == 0 || nameInput.GetValues()[0] != "Seeded Patient" {
		t.Errorf("expected name input to be 'Seeded Patient', got %v", nameInput.GetValues())
	}
	if len(ipInput.GetValues()) == 0 || ipInput.GetValues()[0] != "10.0.0.99" {
		t.Errorf("expected ip input to be '10.0.0.99', got %v", ipInput.GetValues())
	}
}

// Case 15: Config without NewRecord keeps default empty form behavior on newAction.
func TestConsumer_NewRecord_NilDefault(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
		NewRecord: nil,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	f := v.form
	f.SetValues("name", "Previous Unsaved Entry")

	v.newAction()

	nameInput := f.Input("name")
	if len(nameInput.GetValues()) != 0 && nameInput.GetValues()[0] != "" {
		t.Errorf("expected empty name input when NewRecord is nil, got %v", nameInput.GetValues())
	}
}

// Case 16: NewRecord called multiple times returns fresh record instances,
// preventing previous draft edits from leaking into subsequent drafts.
func TestConsumer_NewRecord_ReturnsFreshInstance(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{})

	n := 0
	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
		IDs:       testIDs,
		NewRecord: func() model.Model {
			n++
			return &Device{Name: fmt.Sprintf("Draft %d", n), Ip: "192.168.0.1"}
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// First draft
	v.newAction()
	f := v.form
	if f.Input("name").GetValues()[0] != "Draft 1" {
		t.Errorf("expected 'Draft 1', got %v", f.Input("name").GetValues())
	}

	// User edits the first draft
	f.SetValues("name", "Draft 1 Edited")

	// Second draft
	v.newAction()
	if f.Input("name").GetValues()[0] != "Draft 2" {
		t.Errorf("expected 'Draft 2' for fresh draft, got %v", f.Input("name").GetValues())
	}
}
