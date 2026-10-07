// Root-level test (justified): exercises confirmDelete and deleteLabel — opening the delete confirmation through the public UI requires a loaded record and pointer events.
//go:build wasm

package crudview

import (
	"os"
	"syscall/js"
	"testing"

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
	os.Exit(m.Run())
}

func TestCrudView_DeleteConfirm_Language(t *testing.T) {
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)

	v := &CrudView{Title: "Test CRUD"}
	v.Init(&mockCtx{})
	v.deleteLabel.Set("Device One")
	v.confirmDelete.Open()

	htmlOut := v.confirmDelete.Render().String()

	for _, want := range []string{
		"Confirmar",
		"Eliminar «Device One»?",
		"Esta acción no se puede deshacer.",
		"Cancelar",
	} {
		if !fmt.Contains(htmlOut, want) {
			t.Errorf("expected ES dialog text %q, got: %s", want, htmlOut)
		}
	}
}
