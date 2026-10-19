package crudview

import (
	"strings"
	"testing"

	"webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/view"
)

// fakeScope is a scope control whose marker appears in the list header and
// which also satisfies widget.Filterable — the test asserts crudview does NOT
// wire it to the list filter despite that capability.
type fakeScope struct {
	dom.Element
	filterCalls int
}

func (f *fakeScope) OnFilterChange(fn func(term string)) { f.filterCalls++ }

func (f *fakeScope) String() string { return "<div data-testid='fake-scope'></div>" }
func (f *fakeScope) Render() *dom.Element {
	return Div().Attr("data-testid", "fake-scope")
}

// The Scope slot lands in the list's header (the aside controls band), above the
// filter — not in the article and not in any title band (there is none).
func TestScope_RendersAboveTheFilter(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v := &CrudView{Presenter: p, Scope: &fakeScope{}, Filter: &fakeFilterMarker{}}
	v.Init(&fakeCtx{})

	html := v.Render().String()

	scope := strings.Index(html, "data-testid='fake-scope'")
	filter := strings.Index(html, "data-testid='fake-filter'")
	aside := strings.Index(html, "rp__aside-header")
	if scope < 0 || filter < 0 || aside < 0 {
		t.Fatalf("expected the scope and the filter inside the aside header, markup:\n%s", html)
	}
	if !(aside < scope && scope < filter) {
		t.Errorf("expected the scope control above the filter inside the list header, markup:\n%s", html)
	}
	if strings.Contains(html, "<h1") || strings.Contains(html, "rp__controls") {
		t.Errorf("crudview must not draw a title band, markup:\n%s", html)
	}
}

// fakeFilterMarker is a plain control that marks where the filter lands.
type fakeFilterMarker struct{ dom.Element }

func (f *fakeFilterMarker) String() string { return "<div data-testid='fake-filter'></div>" }
func (f *fakeFilterMarker) Render() *dom.Element {
	return Div().Attr("data-testid", "fake-filter")
}

// No Scope adds nothing: the list header is then just what it always was.
func TestScope_NilAddsNothing(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})

	v := &CrudView{Presenter: p, Scope: nil}
	v.Init(&fakeCtx{})
	html := v.Render().String()

	if strings.Contains(html, "fake-scope") {
		t.Errorf("expected no scope control with Scope nil, markup:\n%s", html)
	}
}

// A Scope that is ALSO widget.Filterable must NOT receive OnFilterChange from
// crudview — changing the scope re-scopes data, it does not filter a term (see
// Config.Scope doc). Asserting 0 calls proves the no-auto-wiring rule.
func TestScope_NotWiredToFilter(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	ctx := &fakeScope{}
	v := &CrudView{Presenter: p, Scope: ctx}
	v.Init(&fakeCtx{})

	if ctx.filterCalls != 0 {
		t.Fatalf("expected crudview NOT to wire Scope to the list filter, got %d wiring calls", ctx.filterCalls)
	}
}

// Config.Scope flows through New into the CrudView and renders in the list header.
func TestScope_FromConfigRenders(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v, err := New(Config{ParentID: "ctx-parent", Presenter: p, IDs: testIDs, Scope: &fakeScope{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&fakeCtx{})

	html := v.Render().String()
	if !strings.Contains(html, "data-testid='fake-scope'") {
		t.Errorf("expected Scope from Config to render, markup:\n%s", html)
	}
}
