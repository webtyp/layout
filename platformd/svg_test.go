//go:build !wasm

package platformd_test

import (
	"strings"
	"testing"
	"webtyp.com/layout/platformd"
)

func TestPlatform_IconSvg_HasRequiredIcons(t *testing.T) {
	p := &platformd.Platform{}
	sprite := p.IconSvg()
	if sprite == nil {
		t.Fatal("IconSvg() returned nil")
	}

	// Content icons (home, products, info) live with their module packages —
	// the chassis only draws its own chrome glyphs.
	s := sprite.String()
	required := []string{"pd-user", "pd-brand", "pd-menu"}
	for _, id := range required {
		if !strings.Contains(s, `id="`+id+`"`) {
			t.Errorf("missing icon: %s", id)
		}
	}
}

func TestPlatform_IconSvg_HasCurrentColor(t *testing.T) {
	p := &platformd.Platform{}
	sprite := p.IconSvg()
	s := sprite.String()
	if !strings.Contains(s, "currentColor") {
		t.Error("icons must use fill=currentColor or stroke=currentColor")
	}
}
