//go:build wasm

package crudview

import (
	"strings"
	"testing"

	"webtyp.com/view"
	"webtyp.com/view/conformance"
)

// searchFirstDevice is Device declared with ListsOnlyAfterSearch.
type searchFirstDevice struct{ Device }

func (*searchFirstDevice) ListsOnlyAfterSearch() bool { return true }

func TestCrudView_Wasm_SearchFirstSaysTypeToSearch(t *testing.T) {
	fb := &conformance.FakeLister{PageSize: view.Page{First: 0, Size: 20}}
	p := view.New(fb, &searchFirstDevice{}, view.WithTitle("Wasm Test"), view.WithEmpty("No items msg"))
	v := &CrudView{Presenter: p}
	v.Init(&mockCtxWasm{})
	v.Reload(nil)

	rendered := v.Render().String()
	if !strings.Contains(rendered, `data-open='true'>Type to search</p>`) {
		t.Errorf("a search-first list with no text must say «Type to search». Got:\n%s", rendered)
	}
	if strings.Contains(rendered, `data-open='true'>No items msg</p>`) {
		t.Errorf("the empty message must stay closed while waiting for text")
	}

	v.search.Set("does-not-exist")
	v.filter()
	rendered = v.Render().String()
	if !strings.Contains(rendered, `data-open='true'>No items msg</p>`) || strings.Contains(rendered, `data-open='true'>Type to search</p>`) {
		t.Errorf("with text and no rows, the empty message shows and «Type to search» does not. Got:\n%s", rendered)
	}
}
