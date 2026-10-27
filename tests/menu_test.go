package tests

import (
	"webtyp.com/dom"
	"webtyp.com/router"

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

	if len(m.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(m.Modules))
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

	if len(m.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(m.Modules))
	}

	desc, ok := m.Modules[0].(layout.Described)
	if !ok {
		t.Fatalf("expected module to be Described")
	}
	if desc.Description() != "Help" {
		t.Errorf("expected Description 'Help', got '%s'", desc.Description())
	}
}

type fakeComp struct{ dom.Element }

func TestBuildMenu_ComponentLeaf(t *testing.T) {
	screens := []view.Screen{
		{
			ID:         "booking",
			Label:      "Booking",
			Icon:       "laptop",
			Permission: "reservation",
			Help:       "Book.",
			Component:  &fakeComp{},
		},
	}
	ids := testIDGen{}

	allows := func(r model.Resource) bool { return string(r) == "reservation" }
	m, err := app.BuildMenu(screens, ids, allows)
	if err != nil {
		t.Fatal(err)
	}

	if m.Modules[0].ModelName() != "booking" {
		t.Errorf("expected module ModelName 'booking', got '%s'", m.Modules[0].ModelName())
	}
	if m.Modules[0].View() != screens[0].Component {
		t.Error("expected View() to return the same component")
	}

	if !m.CanView("booking") {
		t.Error("expected CanView('booking') to be true")
	}

	desc, ok := m.Modules[0].(layout.Described)
	if !ok {
		t.Fatalf("expected module to satisfy layout.Described")
	}
	if desc.Description() != "Book." {
		t.Errorf("expected Description 'Book.', got '%s'", desc.Description())
	}
}

func TestBuildMenu_AddsAPIExplorerLast(t *testing.T) {
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

	allows := func(r model.Resource) bool { return string(r) == string(router.ResourceIntrospection) }
	m, err := app.BuildMenu(screens, ids, allows)
	if err != nil {
		t.Fatal(err)
	}

	last := m.Modules[len(m.Modules)-1]
	if last.ModelName() != string(router.ResourceIntrospection) {
		t.Errorf("expected last module to be API explorer, got '%s'", last.ModelName())
	}

	if !m.CanView(string(router.ResourceIntrospection)) {
		t.Error("expected CanView for api explorer to be true when allows returns true")
	}

	allowsFalse := func(r model.Resource) bool { return false }
	m2, _ := app.BuildMenu(screens, ids, allowsFalse)
	if m2.CanView(string(router.ResourceIntrospection)) {
		t.Error("expected CanView for api explorer to be false when allows returns false")
	}
}
