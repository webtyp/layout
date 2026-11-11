//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/layout/platformd"
)

func TestMenu_Scrolls(t *testing.T) {
	p := &platformd.Platform{}
	cssStr := p.RenderCSS().String()

	if d := declsFor(cssStr, ".pd__drawer-panel"); !strings.Contains(d, "overflow-y: auto;") {
		t.Errorf(".pd__drawer-panel must be the scroll region, decls:\n%s", d)
	}
	if d := declsFor(cssStr, ".pd__navbar"); !strings.Contains(d, "overflow-y: auto;") {
		t.Errorf(".pd__navbar must be the scroll region, decls:\n%s", d)
	}
}
