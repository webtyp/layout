package crudview

import (
	"testing"

	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// A list reached by picking in Scope (one patient's records) has no search bar when the host says so; by default it
// keeps the search bar New installs.
func TestHideSearch_DropsTheDefaultSearchBar(t *testing.T) {
	hidden, err := New(Config{ParentID: "hidden", Presenter: view.New(&conformance.FakeLister{}, &Device{}), IDs: testIDs, HideSearch: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if hidden.Filter != nil {
		t.Errorf("HideSearch: Filter = %T, want none", hidden.Filter)
	}
	shown, err := New(Config{ParentID: "shown", Presenter: view.New(&conformance.FakeLister{}, &Device{}), IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if shown.Filter == nil {
		t.Error("default: the search bar must be installed")
	}
}
