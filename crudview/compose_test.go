package crudview

import (
	"strings"
	"testing"

	"webtyp.com/components/searchbar"
	"webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/model"
	"webtyp.com/view"
	"webtyp.com/view/conformance"
	"webtyp.com/widget"
)

// fakeFilter is a filter control with no markup beyond a marker attribute.
type fakeFilter struct {
	dom.Element
	sink func(term string)
}

func (f *fakeFilter) OnFilterChange(fn func(term string)) { f.sink = fn }

// String is what the stdlib serialization path renders for a Component child;
// Render is what the wasm mount pipeline uses. The marker appears through both.
func (f *fakeFilter) String() string { return "<div data-testid='fake-filter'></div>" }
func (f *fakeFilter) Render() *dom.Element {
	return Div().Attr("data-testid", "fake-filter")
}

// fakeListBackend returns a FakeLister seeded with three rows, like the demo store.
func fakeListBackend() *conformance.FakeLister {
	return &conformance.FakeLister{
		Rows: []model.Model{
			&Device{Id: "12", Name: "Frontend Device", Ip: "192.168.1.10"},
			&Device{Id: "23", Name: "Backend Server", Ip: "10.0.0.5"},
			&Device{Id: "34", Name: "Database Instance", Ip: "mysql-production"},
		},
	}
}

// The controller composes the rightpanel skeleton and no longer paints its own
// grid: the old crudview__detail/search classes are gone from the markup.
func TestCrudView_RendersThroughRightPanel(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Filter:    &fakeFilter{},
	}
	v.Init(&fakeCtx{})

	html := v.Render().String()

	for _, want := range []string{"class='rp'", "rp__main", "rp__aside",
		"rp__aside-header", "rp__aside-footer"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in markup:\n%s", want, html)
		}
	}
	for _, gone := range []string{"crudview__detail", "crudview__search"} {
		if strings.Contains(html, gone) {
			t.Errorf("expected no %q: rightpanel owns the frame", gone)
		}
	}
}

// The Filter slot lands in the aside's controls band.
func TestCrudView_FilterSlotIsRendered(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Filter:    &fakeFilter{},
	}
	v.Init(&fakeCtx{})

	html := v.Render().String()

	if !strings.Contains(html, "class='rp__aside-header'") {
		t.Errorf("expected the controls band, markup:\n%s", html)
	}
	if !strings.Contains(html, "data-testid='fake-filter'") {
		t.Errorf("expected the filter control inside the controls band, markup:\n%s", html)
	}
}

// No Filter paints no controls band at all — and no search input either.
func TestCrudView_NoFilterPaintsNoControlsBand(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v := &CrudView{Title: "CRUD", Presenter: p}
	v.Init(&fakeCtx{})

	html := v.Render().String()

	for _, gone := range []string{"rp__aside-header", "type='search'"} {
		if strings.Contains(html, gone) {
			t.Errorf("expected no %q without a Filter, markup:\n%s", gone, html)
		}
	}
}

// A widget.Filterable slot drives the list filter through the Init seam.
func TestCrudView_FilterableDrivesTheList(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	f := &fakeFilter{}
	v := &CrudView{Title: "CRUD", Presenter: p, Filter: f}
	v.Init(&fakeCtx{})
	v.Activate()

	if f.sink == nil {
		t.Fatal("expected Init to wire the filter control to the list")
	}

	// Narrow to the one matching row.
	f.sink("backend")
	if labels := cardLabels(v); len(labels) != 1 || labels[0] != "Backend Server" {
		t.Errorf("expected only Backend Server after sink(\"backend\"), got %v", labels)
	}

	// Clearing restores every row.
	f.sink("")
	if labels := cardLabels(v); len(labels) != 3 {
		t.Errorf("expected all 3 rows after sink(\"\"), got %v", labels)
	}
}

// New with no Filter installs the ergonomic default: a searchbar carrying the
// presenter's placeholder.
func TestCrudView_NewInstallsADefaultFilter(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v, err := New(Config{ParentID: "compose", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if v.Filter == nil {
		t.Fatal("expected New to install a default Filter")
	}
	if _, ok := v.Filter.(widget.Filterable); !ok {
		t.Fatalf("expected the default Filter to be widget.Filterable, got %T", v.Filter)
	}
	if _, ok := v.Filter.(*searchbar.SearchBar); !ok {
		t.Errorf("expected the default Filter to be a *searchbar.SearchBar, got %T", v.Filter)
	}
}
