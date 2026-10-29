//go:build wasm

package schemaview

import (
	"syscall/js"
	"testing"
	. "webtyp.com/dom"
	"webtyp.com/erd"
	. "webtyp.com/html"
	"webtyp.com/lang"
)

func TestMain(m *testing.M) {
	js.Global().Get("document").Get("body").Call("insertAdjacentHTML", "beforeend", `<script id="`+lang.ScriptID+`" type="application/json">{"Database":"Database","Search warnings by table or column":"Search warnings by table or column"}</script>`)
	m.Run()
}

func TestSchemaView(t *testing.T) {
	v := New()
	// manually mock the state, Activate would fetch it from network
	v.report = &erd.Report{
		SVG: "<svg><g id=\"erd-table-user\"></g></svg>",
		Findings: []erd.Finding{
			{Rule: erd.RuleMissingTable, Table: "user", Column: "tenant_id", Target: "tenant"},
			{Rule: erd.RuleIsolatedTable, Table: "setting"},
		},
	}
	v.Init(nil)
	v.svgNodes.Set([]*Element{Div().Raw(Trust(v.report.SVG))})
	v.updateRows()

	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "test-schemaview-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() {
		doc.Get("body").Call("removeChild", root)
	})

	if err := Render("test-schemaview-root", v); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	findings := js.Global().Get("document").Call("querySelectorAll", ".schemaview__finding")
	if findings.Length() != 2 {
		t.Fatalf("expected 2 findings, got %d", findings.Length())
	}

	// Check that one of them has finding-error (RuleMissingTable is LevelError, RuleIsolatedTable is LevelWarning)
	hasError := false
	for i := 0; i < findings.Length(); i++ {
		el := findings.Index(i)
		if el.Get("classList").Call("contains", "schemaview__finding-error").Bool() {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected one finding to have finding-error class")
	}

	v.search.Set("tenant")
	v.updateRows()

	findings = js.Global().Get("document").Call("querySelectorAll", ".schemaview__finding")
	if findings.Length() != 1 {
		t.Fatalf("expected 1 finding after search, got %d", findings.Length())
	}

	tableUser := js.Global().Get("document").Call("getElementById", "erd-table-user")
	if tableUser.IsNull() || tableUser.IsUndefined() {
		t.Error("expected document.getElementById('erd-table-user') to exist")
	}
}
