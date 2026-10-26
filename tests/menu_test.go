package tests

import (
	"strings"
	"testing"

	"webtyp.com/layout"
	"webtyp.com/layout/app"
	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

type testIDGen struct{}

func (testIDGen) NewID() string { return "test-id" }

func TestBuildMenu_InvalidScreenIsAnError(t *testing.T) {
	// A leaf without a permission
	screens := []view.Screen{
		{
			ID:        "leaf",
			Label:     "Leaf",
			Icon:      "laptop",
			Presenter: view.New(&conformance.FakeLister{}, &conformance.MockRecord{}),
		},
	}
	ids := testIDGen{}
	_, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return true })
	if err == nil {
		t.Fatal("expected error for leaf screen without Permission, got nil")
	}
	if !strings.Contains(err.Error(), "needs a Permission") {
		t.Fatalf("expected error containing 'needs a Permission', got %v", err)
	}
}

func TestBuildMenu_LeafVisibleByItsPermission(t *testing.T) {
	screens := []view.Screen{
		{
			ID:         "x",
			Label:      "X",
			Icon:       "laptop",
			Permission: "service",
			Presenter:  view.New(&conformance.FakeLister{}, &conformance.MockRecord{}),
		},
	}
	ids := testIDGen{}

	// allows true only for "service"
	allowsService := func(r model.Resource) bool { return string(r) == "service" }
	m, err := app.BuildMenu(screens, ids, allowsService)
	if err != nil {
		t.Fatal(err)
	}
	if !m.CanView("x") {
		t.Error("expected CanView('x') to be true when allows returns true")
	}

	// allows always false
	allowsFalse := func(r model.Resource) bool { return false }
	m2, err := app.BuildMenu(screens, ids, allowsFalse)
	if err != nil {
		t.Fatal(err)
	}
	if m2.CanView("x") {
		t.Error("expected CanView('x') to be false when allows always false")
	}
}

func TestBuildMenu_GroupHasNoOwnPermission(t *testing.T) {
	screens := []view.Screen{
		{
			ID:    "g",
			Label: "Group",
			Icon:  "laptop",
			Help:  "help",
			Children: []view.Screen{
				{
					ID:         "g.a",
					Label:      "A",
					Icon:       "laptop",
					Permission: "a",
					Presenter:  view.New(&conformance.FakeLister{}, &conformance.MockRecord{}),
				},
			},
		},
	}
	ids := testIDGen{}

	m, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return string(r) == "a" })
	if err != nil {
		t.Fatal(err)
	}
	if m.CanView("g") {
		t.Error("expected CanView('g') to be false (groups have no permission)")
	}
	if !m.CanView("g.a") {
		t.Error("expected CanView('g.a') to be true")
	}

	if len(m.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(m.Modules))
	}

	gm, ok := m.Modules[0].(layout.GroupModule)
	if !ok {
		t.Fatalf("expected module to be a GroupModule, got %T", m.Modules[0])
	}
	if len(gm.Children()) != 1 {
		t.Fatalf("expected 1 child in group, got %d", len(gm.Children()))
	}
}

func TestBuildMenu_DescriptionFromPresenter(t *testing.T) {
	presenter := view.New(&conformance.FakeLister{}, &conformance.MockRecord{}, view.WithDescription("Help"))

	screens := []view.Screen{
		{
			ID:         "leaf",
			Label:      "Leaf",
			Icon:       "laptop",
			Permission: "p",
			Presenter:  presenter,
		},
	}
	ids := testIDGen{}

	m, err := app.BuildMenu(screens, ids, func(r model.Resource) bool { return true })
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(m.Modules))
	}

	desc, ok := m.Modules[0].(layout.Described)
	if !ok {
		t.Fatalf("expected module to be Described")
	}
	if desc.Description() != "Help" {
		t.Errorf("expected Description 'Help', got '%s'", desc.Description())
	}
}
