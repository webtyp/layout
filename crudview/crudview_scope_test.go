//go:build !wasm

package crudview

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// newTestCrudView builds a fully wired controller over the standard Device
// fixture: a presenter fed by FakeLister, Init'd and Reloaded so the list is
// populated. The filter scope below is whatever the fixture's Filter returns
// for a term, so a term with no matches empties the list entirely.
func newTestCrudView(t *testing.T) *CrudView {
	t.Helper()
	fb := &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Device One", Ip: "192.168.1.1"},
			&Device{Id: "23", Name: "Device Two", Ip: "192.168.1.2"},
		},
	}
	p := view.New(fb, &Device{})

	v, err := New(Config{ParentID: "my-id", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("newTestCrudView: New: %v", err)
	}
	v.Init(&fakeCtx{})
	v.Activate()
	return v
}

func TestFilterDropsASelectionItNoLongerShows(t *testing.T) {
	// The reported bug: pick patient A, open one of A's records, switch to
	// patient B. The list repopulates; the form must not keep A's record
	// loaded, editable and saveable against a list that no longer shows it.
	v := newTestCrudView(t)

	// Select something the current list contains.
	items := v.list.Items()
	if len(items) == 0 {
		t.Fatal("fixture must provide at least one item")
	}
	v.selectAction(items[0])
	if v.selected.Get() == "" {
		t.Fatal("precondition: a record must be selected")
	}

	// Move the scope so the filtered list can no longer contain it.
	v.search.Set("no-such-scope-xyz")
	v.filter()

	if got := v.selected.Get(); got != "" {
		t.Errorf("selection must be dropped when it leaves the filtered list, still holds %q", got)
	}
	if v.composing.Get() {
		t.Error("a composing draft must be dropped with the scope it belonged to")
	}
	if v.canDelete.Get() {
		t.Error("delete must be disarmed once nothing is selected")
	}
}

func TestFilterKeepsASelectionStillInScope(t *testing.T) {
	// The invariant must not overreach: narrowing a search that still matches
	// the open record has to leave it alone, or typing in a search box would
	// wipe the form mid-edit.
	v := newTestCrudView(t)
	items := v.list.Items()
	v.selectAction(items[0])
	want := v.selected.Get()

	v.search.Set("") // widest possible scope: everything matches
	v.filter()

	if got := v.selected.Get(); got != want {
		t.Errorf("a selection still in the list must survive the filter: want %q, got %q", want, got)
	}
}

func TestFilterOnAnEmptyControllerIsANoop(t *testing.T) {
	// Nothing selected, nothing composing: filtering must not touch the form
	// or bounce a phone back to the list for no reason.
	v := newTestCrudView(t)
	v.search.Set("anything")
	v.filter() // must not panic and must leave the resting state alone
	if v.selected.Get() != "" || v.composing.Get() {
		t.Error("filtering an empty controller must change nothing")
	}
}
