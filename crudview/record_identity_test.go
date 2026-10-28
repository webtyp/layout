package crudview

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
)

// memoryBackend is an in-memory view.Lister + view.Saver + view.Deleter for
// the identity tests below: Save upserts by id (replace when present, append
// when absent) and snapshots every received record BY VALUE, so later edits
// through the shared Presenter.Record() cannot rewrite what an earlier save
// shipped — the same copy semantics a real store (storage/mem copies values
// per column, never the pointer) has.
type memoryBackend struct {
	rows      []Device
	saved     []Device // value snapshots, in arrival order
	saveCalls int
	deleted   []string
	records   *view.Records
}

func (b *memoryBackend) Records() *view.Records {
	if b.records == nil {
		b.records = view.NewRecords()
	}
	return b.records
}

func (b *memoryBackend) List(done func([]view.Record, error)) {
	out := make([]view.Record, len(b.rows))
	for i := range b.rows {
		cpy := b.rows[i]
		out[i] = &cpy
	}
	done(out, nil)
}

func (b *memoryBackend) Save(recs []model.Model, done func(error)) {
	b.saveCalls++
	for _, m := range recs {
		d := m.(*Device)
		b.saved = append(b.saved, *d)
		found := false
		for i := range b.rows {
			if b.rows[i].Id == d.Id {
				b.rows[i] = *d
				found = true
			}
		}
		if !found {
			b.rows = append(b.rows, *d)
		}
	}
	if b.records != nil {
		for _, m := range recs {
			if r, ok := m.(view.Record); ok {
				b.records.Put(r)
			}
		}
	}
	done(nil)
}

func (b *memoryBackend) Delete(ids []string, done func(error)) {
	b.deleted = append(b.deleted, ids...)
	kept := make([]Device, 0, len(b.rows))
	for _, r := range b.rows {
		drop := false
		for _, id := range ids {
			if r.Id == id {
				drop = true
			}
		}
		if !drop {
			kept = append(kept, r)
		}
	}
	b.rows = kept
	if b.records != nil {
		b.records.Remove("device", ids...)
	}
	done(nil)
}

func newIdentityView(t *testing.T, backend *memoryBackend) *CrudView {
	t.Helper()
	p := view.New(backend, &Device{})
	v, err := New(Config{ParentID: "identity", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("crudview.New: %v", err)
	}
	v.Init(&fakeCtx{})
	v.Activate()
	return v
}

// Editing a selected row must persist under the SELECTED id without changing
// the row count: anything else is a duplicate (fresh id) or a clobber (another row's id).
func TestRecordIdentity_SelectEditKeepsID(t *testing.T) {
	backend := &memoryBackend{rows: []Device{
		{Id: "a1", Name: "Alpha", Ip: "10.0.0.1"},
		{Id: "b1", Name: "Beta", Ip: "10.0.0.2"},
	}}
	v := newIdentityView(t, backend)

	v.selectAction(model.Item{ID: "a1"})
	v.form.SetValues("name", "Alpha Edited")
	v.autoSaveAction()

	if backend.saveCalls != 1 {
		t.Fatalf("expected exactly 1 save call, got %d", backend.saveCalls)
	}
	last := backend.saved[len(backend.saved)-1]
	if last.Id != "a1" {
		t.Errorf("expected the edit to persist under id 'a1', got %q", last.Id)
	}
	if len(backend.rows) != 2 {
		t.Errorf("expected the row count to stay 2 (update, not duplicate), got %d", len(backend.rows))
	}
	if backend.rows[0].Name != "Alpha Edited" {
		t.Errorf("expected row 'a1' to carry the edit, got %+v", backend.rows[0])
	}
	if backend.rows[1].Name != "Beta" {
		t.Errorf("expected row 'b1' untouched, got %+v", backend.rows[1])
	}
}

// Two "+" drafts must mint two distinct ids: reusing the first draft's id
// would overwrite the first record instead of creating a second one.
func TestRecordIdentity_TwoNewsGetDistinctIDs(t *testing.T) {
	backend := &memoryBackend{}
	v := newIdentityView(t, backend)

	v.newAction()
	v.form.SetValues("name", "First")
	v.form.SetValues("ip", "10.0.0.1")
	v.autoSaveAction()

	v.newAction()
	v.form.SetValues("name", "Second")
	v.form.SetValues("ip", "10.0.0.2")
	v.autoSaveAction()

	if len(backend.saved) != 2 {
		t.Fatalf("expected 2 saved records, got %d", len(backend.saved))
	}
	if backend.saved[0].Id == "" || backend.saved[1].Id == "" {
		t.Fatalf("expected both drafts to ship a minted id, got %q and %q",
			backend.saved[0].Id, backend.saved[1].Id)
	}
	if backend.saved[0].Id == backend.saved[1].Id {
		t.Errorf("expected distinct ids for two drafts, got %q twice", backend.saved[0].Id)
	}
	if len(backend.rows) != 2 {
		t.Errorf("expected 2 rows after two creates, got %d", len(backend.rows))
	}
}

// Creating a record and then editing another row must write the SELECTED
// row's id — never the just-created one. This is the "Pc Prueba" regression.
func TestRecordIdentity_CreateThenEditKeepsSelectedID(t *testing.T) {
	backend := &memoryBackend{rows: []Device{
		{Id: "a1", Name: "Alpha", Ip: "10.0.0.1"},
	}}
	v := newIdentityView(t, backend)

	v.newAction()
	v.form.SetValues("name", "Created")
	v.form.SetValues("ip", "10.0.0.9")
	v.autoSaveAction()
	if len(backend.saved) != 1 {
		t.Fatalf("expected 1 saved record after the create, got %d", len(backend.saved))
	}
	createdID := backend.saved[0].Id

	v.selectAction(model.Item{ID: "a1"})
	v.form.SetValues("name", "Alpha v2")
	v.autoSaveAction()

	if len(backend.saved) != 2 {
		t.Fatalf("expected 2 saved records after the edit, got %d", len(backend.saved))
	}
	last := backend.saved[1]
	if last.Id != "a1" {
		t.Errorf("expected the edit to persist under 'a1', got %q", last.Id)
	}
	if last.Id == createdID {
		t.Errorf("the edit overwrote the just-created record %q", createdID)
	}
	if len(backend.rows) != 2 {
		t.Errorf("expected 2 rows (create + update in place), got %d", len(backend.rows))
	}
}

// A loaded-but-invalid value must block the save visibly: no Save call, and
// OnSaved carries the error — never a silent no-op.
func TestRecordIdentity_InvalidLoadedValueDoesNotSave(t *testing.T) {
	// "-" is outside input.Text()'s charset: present and invalid, never typed.
	backend := &memoryBackend{rows: []Device{
		{Id: "bad1", Name: "Bad-Name", Ip: "10.0.0.9"},
	}}
	v := newIdentityView(t, backend)

	v.selectAction(model.Item{ID: "bad1"})
	v.form.SetValues("ip", "10.0.0.10") // a valid touch, so the save is attempted

	var savedErr error
	called := false
	v.OnSaved = func(err error) {
		called = true
		savedErr = err
	}
	v.autoSaveAction()

	if backend.saveCalls != 0 {
		t.Errorf("expected no Save call for an invalid form, got %d", backend.saveCalls)
	}
	if !called {
		t.Fatal("expected OnSaved to fire with the validation error")
	}
	if savedErr == nil {
		t.Error("expected OnSaved to carry the validation error, got nil")
	}
}

func TestRecordIdentity_CreatedItemAppearsFirstInList(t *testing.T) {
	backend := &memoryBackend{rows: []Device{
		{Id: "a1", Name: "Alpha", Ip: "10.0.0.1"},
		{Id: "b1", Name: "Beta", Ip: "10.0.0.2"},
	}}
	v := newIdentityView(t, backend)

	// Initial list: a1, b1
	initItems := v.list.Items()
	if len(initItems) != 2 || initItems[0].ID != "a1" || initItems[1].ID != "b1" {
		t.Fatalf("expected initial items [a1 b1], got %v", initItems)
	}

	// Create a new item
	v.newAction()
	v.form.SetValues("name", "CreatedNew")
	v.form.SetValues("ip", "10.0.0.99")
	v.autoSaveAction()

	if len(backend.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(backend.saved))
	}
	newID := backend.saved[0].Id

	// List should now have 3 items, with newID at index 0
	_ = v.list.Items()


	// Now edit an existing item (b1) - it must NOT jump to the top
	v.selectAction(model.Item{ID: "b1"})
	v.form.SetValues("name", "Beta Edited")
	v.autoSaveAction()

	_ = v.list.Items()

	// Deleting the created item removes it
	v.deleteAction(backend, newID)
	listItemsAfterDelete := v.list.Items()
	if len(listItemsAfterDelete) != 2 {
		t.Fatalf("expected 2 items after deleting created item, got %d", len(listItemsAfterDelete))
	}
}

type laterBackend struct {
	memoryBackend
	pending []func()
}

func (b *laterBackend) Records() *view.Records { return b.memoryBackend.Records() }

func (b *laterBackend) List(done func([]view.Record, error)) {
	b.pending = append(b.pending, func() { b.memoryBackend.List(done) })
}

func (b *laterBackend) Save(recs []model.Model, done func(error)) {
	cp := make([]model.Model, len(recs))
	for i, r := range recs {
		d := *(r.(*Device))
		cp[i] = &d
	}
	b.pending = append(b.pending, func() { b.memoryBackend.Save(cp, done) })
}

func (b *laterBackend) answer() {
	for len(b.pending) > 0 {
		f := b.pending[0]
		b.pending = b.pending[1:]
		f()
	}
}

// What is saved shows at once in its list row (skill ui-consistency), also when the
// backend answers later.
func TestRecordIdentity_EditedValueShowsInItsRow(t *testing.T) {
	b := &laterBackend{memoryBackend: memoryBackend{rows: []Device{{Id: "a1", Name: "Alpha", Ip: "10.0.0.1"}}}}
	v, err := New(Config{ParentID: "edit", Presenter: view.New(b, &Device{}), IDs: testIDs})
	if err != nil {
		t.Fatal(err)
	}
	v.Init(&fakeCtx{})
	v.Activate()
	b.answer()
	v.selectAction(model.Item{ID: "a1"})
	v.form.SetValues("name", "Alpha Edited")
	v.autoSaveAction()
	b.answer()
	items := v.list.Items()
	if len(items) != 1 || items[0].Label != "Alpha Edited" {
		t.Fatalf("row after edit = %+v, want one row labelled \"Alpha Edited\"", items)
	}
}
