// Root-level test (justified): exercises confirmDelete and deleteLabel — opening the delete confirmation through the public UI requires a loaded record and pointer events.
//go:build wasm

package crudview

import (
	"os"
	"syscall/js"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/lang"
)

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

func TestMain(m *testing.M) {
	doc := js.Global().Get("document")
	el := doc.Call("createElement", "script")
	el.Set("type", "application/json")
	el.Set("id", lang.ScriptID)
	el.Set("textContent", `{"default":"es","languages":["es"],"keys":{"Cancel":["Cancelar"],"Confirm":["Confirmar"],"Delete":["Eliminar"],"This":["Esta"],"action":["acción"],"be":["se"],"cannot":["no"],"undone.":["puede deshacer."]}}`)
	doc.Get("head").Call("appendChild", el)
	app := doc.Call("createElement", "div")
	app.Set("id", "app")
	doc.Get("body").Call("appendChild", app)
	os.Exit(m.Run())
}

func TestCrudView_DeleteConfirm_Language(t *testing.T) {
	// In WASM the dialog's body is filled when it is mounted (Show), so it is
	// mounted and read from the live DOM.
	openDialog := func() string {
		v := &CrudView{Title: "Test CRUD"}
		v.Init(&mockCtx{})
		v.deleteLabel.Set("Device One")
		if err := dom.Render("app", v.confirmDelete); err != nil {
			t.Fatalf("mount dialog: %v", err)
		}
		v.confirmDelete.Open() // after mounting: Init (run by Render) closes it
		return js.Global().Get("document").Call("getElementById", "app").Get("textContent").String()
	}

	// The page dictionary (inserted by TestMain) translates the same chrome.
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)
	html := openDialog()
	for _, want := range []string{
		"Confirmar",
		"Eliminar «Device One»?",
		"Esta acción no se puede deshacer.",
		"Cancelar",
	} {
		if !fmt.Contains(html, want) {
			t.Errorf("expected ES dialog text %q, got: %s", want, html)
		}
	}
}
