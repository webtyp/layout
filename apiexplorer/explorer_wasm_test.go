//go:build wasm

// Root-level test (justified): pure logic tests must test unexported functions, which the tests/ package cannot access.
package apiexplorer

import (
	"strings"
	"syscall/js"
	"testing"
	. "webtyp.com/dom"
	"webtyp.com/router"
)

func TestExplorer_WASM_Mount(t *testing.T) {
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "test-explorer-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() {
		doc.Get("body").Call("removeChild", root)
	})

	e, err := New(Config{})
	if err != nil {
		t.Fatalf("failed to create Explorer: %v", err)
	}
	e.Init(nil)

	e.tableData = &router.RouteTable{
		Routes: []router.RouteRecord{
			{Method: "GET", Path: "/api/public", Access: "public"},
			{Method: "POST", Path: "/api/orphan", Access: "guarded", PolicyKnown: true, Roles: []string{}}, // Orphan
		},
	}
	e.updateRows()

	if err := Render("test-explorer-root", e); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	rows := doc.Call("querySelectorAll", ".apiexplorer__row")
	if rows.Get("length").Int() != 2 {
		t.Fatalf("expected 2 .apiexplorer__row elements, got %d", rows.Get("length").Int())
	}

	firstRow := rows.Call("item", 0)
	classes := firstRow.Get("className").String()
	if !strings.Contains(classes, "apiexplorer__orphan") {
		t.Errorf("expected first row to be orphan, got classes: %s", classes)
	}

	firstRow.Call("click")

	tryForm := doc.Call("querySelectorAll", ".apiexplorer__try-form")
	if tryForm.Get("length").Int() != 1 {
		t.Fatalf("expected 1 try-form after expansion, got %d", tryForm.Get("length").Int())
	}

	resultText := tryForm.Call("item", 0).Call("querySelector", ".apiexplorer__result-panel").Get("textContent").String()
	if resultText != "" {
		t.Errorf("expected empty result panel on expand, got %q", resultText)
	}
}
