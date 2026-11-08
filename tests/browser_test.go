//go:build wasm

package tests

import (
	"testing"

	"webtyp.com/appkit"
	"webtyp.com/view"
)

func TestBrowser_ShapeOfApp(t *testing.T) {
	app := appkit.App{
		Log:      func(...any) {},
		Timezone: "America/Santiago",
		Modules:  []appkit.Module{{Name: "x", Browser: view.ModuleBrowser{}}},
	}
	if app.Modules[0].Name != "x" {
		t.Fatalf("Name = %q, want x", app.Modules[0].Name)
	}
}
