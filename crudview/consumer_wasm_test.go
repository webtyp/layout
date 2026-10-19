//go:build wasm

package crudview

import (
	"testing"

	"webtyp.com/components/searchbar"
	"webtyp.com/form"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

type mockCtxWasmConsumer struct{}

func (m *mockCtxWasmConsumer) OnCleanup(fn func()) {}

func TestConsumer_Wasm_Render(t *testing.T) {
	fb := &conformance.FakeLister{}
	p := view.New(fb, &Device{},
		view.WithTitle("Custom Search Title"),
		view.WithSearchPlaceholder("Custom Search..."))

	cfg := Config{
		ParentID:  "my-wasm-id",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error in New: %v", err)
	}

	v.Init(&mockCtxWasmConsumer{})

	// Check that the form was wired correctly
	f, ok := v.Form.(*form.Form)
	if !ok {
		t.Fatalf("expected form to be *form.Form, got %T", v.Form)
	}

	if f.ParentID() != "my-wasm-id" {
		t.Errorf("expected parent ID to be 'my-wasm-id', got '%s'", f.ParentID())
	}

	root := v.Render()
	if root == nil {
		t.Fatal("expected rendered root element to be non-nil")
	}

	// The presenter's placeholder travels with the filter control — the
	// default searchbar New installed — not in crudview's own markup.
	sb, ok := v.Filter.(*searchbar.SearchBar)
	if !ok {
		t.Fatalf("expected the default Filter to be a *searchbar.SearchBar, got %T", v.Filter)
	}
	if sb.Placeholder != "Custom Search..." {
		t.Errorf("expected the searchbar to carry the presenter's placeholder, got %q", sb.Placeholder)
	}
}
