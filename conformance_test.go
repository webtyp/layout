//go:build !wasm

// The external test package, not layout: Guard C reflects over
// rightpanel.RightPanel, and rightpanel imports layout — an internal test file
// would close that cycle and Go forbids it. The source-reading guards work the
// same from either package.
package layout_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/layout/rightpanel"
)

// eachStyleFile walks every css.go in the repository and hands its parsed AST
// to fn. Skips docs/, web/ and .git.
func eachStyleFile(t *testing.T, fn func(path string, node *ast.File)) {
	t.Helper()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			switch info.Name() {
			case "docs", "web", ".git", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Base(path) != "css.go" {
			return nil
		}
		fset := token.NewFileSet()
		node, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return fmt.Errorf("failed to parse %s: %w", path, perr)
		}
		fn(path, node)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// eachGoFile walks every non-test .go file in the repository, like
// eachStyleFile but without the css.go filter.
func eachGoFile(t *testing.T, fn func(path string, node *ast.File)) {
	t.Helper()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			switch info.Name() {
			case "docs", "web", ".git", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		node, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return fmt.Errorf("failed to parse %s: %w", path, perr)
		}
		fn(path, node)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func appendOnce(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// TestOnlyOneOwnerOfTheGrid is the test that would have made the
// rightpanel/crudview duplication impossible. A layout primitive that arranges
// a module's top-level panels belongs to exactly one package: the skeleton.
// A second package emitting one is a second skeleton, which is how the first
// duplication started and how the next one would.
func TestOnlyOneOwnerOfTheGrid(t *testing.T) {
	// Split and MasterDetail arrange a module's panels. Grid and Stack are NOT
	// listed: they are ordinary rhythm inside any widget and banning them would
	// be noise.
	owners := map[string][]string{} // primitive → packages that emit it

	eachStyleFile(t, func(path string, node *ast.File) {
		pkg := filepath.Dir(path)
		ast.Inspect(node, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "Split", "MasterDetail":
				owners[sel.Sel.Name] = appendOnce(owners[sel.Sel.Name], pkg)
			}
			return true
		})
	})

	for primitive, pkgs := range owners {
		if len(pkgs) > 1 {
			t.Errorf("%s is emitted by %d packages (%v); the module frame has ONE owner.\n"+
				"If a package needs this layout, it must compose the skeleton that owns it "+
				"(rightpanel), not re-declare the primitive. See docs/ARQ_REFACTOR.md.",
				primitive, len(pkgs), pkgs)
		}
	}
}

// TestNoLocallyDeclaredSeams rejects the leaf-patch: a one-method interface
// declared here whose name matches a capability the ecosystem already names in
// widget/capability.go. Declaring it locally forks that contract, and the copy
// can never be reused — the failure mode described in docs/ARQ_REFACTOR.md §2.
func TestNoLocallyDeclaredSeams(t *testing.T) {
	// Method names that widget/capability.go already owns.
	upstream := map[string]string{
		"OnFilterChange": "widget.Filterable",
		"Select":         "widget.Selectable",
		"Dismiss":        "widget.Dismissible",
		"Expand":         "widget.Expandable",
	}

	eachGoFile(t, func(path string, node *ast.File) {
		ast.Inspect(node, func(n ast.Node) bool {
			it, ok := n.(*ast.InterfaceType)
			if !ok || it.Methods == nil {
				return true
			}
			var methods []*ast.Ident
			for _, field := range it.Methods.List {
				for _, name := range field.Names {
					methods = append(methods, name)
				}
			}
			if len(methods) != 1 {
				return true
			}
			name := methods[0].Name
			if replacement, ok := upstream[name]; ok {
				t.Errorf("%s declares a local interface with the single method %s; "+
					"use %s instead. A consumer never re-creates an upstream symbol "+
					"locally (construction harness, \"lego rules\").", path, name, replacement)
			}
			return true
		})
	})
}

// slotStub is a minimal Component that paints a uniquely-marked element, so a
// rendered slot can be told apart from an empty one. Both serialization paths
// must carry the marker: the SSR path renders a Component child through
// String(), the wasm mount path through Render().
type slotStub struct {
	dom.Element
	name string
}

func (s *slotStub) Render() *dom.Element {
	return dom.NewElement("div").Attr("data-slot", s.name)
}

func (s *slotStub) String() string {
	return "<div data-slot='" + s.name + "'></div>"
}

// TestEverySlotIsRendered guards the third failure mode: a slot added to the
// struct but never wired into Render(), which fails silently — the consumer
// sets it and nothing appears. Every exported Component field on RightPanel
// must show up in the rendered markup when populated.
func TestEverySlotIsRendered(t *testing.T) {
	rp := rightpanel.RightPanel{}
	tp := reflect.TypeOf(rp)
	pv := reflect.ValueOf(&rp).Elem()
	componentType := reflect.TypeOf((*dom.Component)(nil)).Elem()

	markers := map[string]string{} // field name → marker it must leave behind
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		if field.PkgPath != "" || !field.Type.AssignableTo(componentType) {
			continue
		}
		stub := &slotStub{name: field.Name}
		pv.Field(i).Set(reflect.ValueOf(stub))
		markers[field.Name] = "data-slot='" + field.Name + "'"
	}

	out := rp.Render().String()
	for name, marker := range markers {
		if !strings.Contains(out, marker) {
			t.Errorf("RightPanel.%s is declared but never rendered: a slot that "+
				"cannot be seen is a silent failure (construction harness, principle 6).",
				name)
		}
	}
}
