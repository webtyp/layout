package crudview

import (
	"strings"
	"testing"

	"webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/view"
)

// fakeContext is a context control whose marker appears in the title band and
// which also satisfies widget.Filterable — the test asserts crudview does NOT
// wire it to the list filter despite that capability.
type fakeContext struct {
	dom.Element
	filterCalls int
}

func (f *fakeContext) OnFilterChange(fn func(term string)) { f.filterCalls++ }

func (f *fakeContext) String() string { return "<div data-testid='fake-context'></div>" }
func (f *fakeContext) Render() *dom.Element {
	return Div().Attr("data-testid", "fake-context")
}

// The Context slot lands in the title band (below the H1), not in the aside.
func TestContext_RendersInHeadBand(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v := &CrudView{Title: "CRUD", Presenter: p, Context: &fakeContext{}}
	v.Init(&fakeCtx{})

	html := v.Render().String()

	if !strings.Contains(html, "class='rp__controls'") {
		t.Errorf("expected the title-controls band, markup:\n%s", html)
	}
	if !strings.Contains(html, "data-testid='fake-context'") {
		t.Errorf("expected the context control inside the title band, markup:\n%s", html)
	}
	// The Context band must sit BELOW the H1 and must NOT appear inside the
	// aside header (the Filter slot).
	if strings.Index(html, "</h1>") > strings.Index(html, "data-testid='fake-context'") {
		t.Errorf("expected the context control to render below the H1 title, markup:\n%s", html)
	}
	if !strings.Contains(html, "fake-context") && strings.Contains(html, "aside-header") {
		t.Errorf("context must not leak into the aside filter band")
	}
}

// No Context paints no title band at all — identical to today's markup.
func TestContext_NilNoBand(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})

	withCtx := &CrudView{Title: "CRUD", Presenter: p, Context: nil}
	withCtx.Init(&fakeCtx{})
	html := withCtx.Render().String()

	if strings.Contains(html, "rp__controls") {
		t.Errorf("expected no title-controls band with Context nil, markup:\n%s", html)
	}
}

// A Context that is ALSO widget.Filterable must NOT receive OnFilterChange
// from crudview — changing context re-scopes data, it does not filter a term
// (see Config.Context doc). Asserting 0 calls proves the no-auto-wiring rule.
func TestContext_NotWiredToFilter(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	ctx := &fakeContext{}
	v := &CrudView{Title: "CRUD", Presenter: p, Context: ctx}
	v.Init(&fakeCtx{})

	if ctx.filterCalls != 0 {
		t.Fatalf("expected crudview NOT to wire Context to the list filter, got %d wiring calls", ctx.filterCalls)
	}
}

// Config.Context flows through New into the CrudView and renders in the band.
func TestContext_FromConfigRenders(t *testing.T) {
	p := view.New(fakeListBackend(), &Device{})
	v, err := New(Config{ParentID: "ctx-parent", Presenter: p, IDs: testIDs, Context: &fakeContext{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&fakeCtx{})

	html := v.Render().String()
	if !strings.Contains(html, "data-testid='fake-context'") {
		t.Errorf("expected Context from Config to render, markup:\n%s", html)
	}
}
