package rightpanel_test

import (
	"strings"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/layout/rightpanel"
)

// stubModule implements Module for tests.
type stubModule struct{ name string }

func (s stubModule) ModelName() string { return s.name }

// stubComponent implements dom.Component for tests.
type stubComponent struct{ html string }

func (s *stubComponent) GetID() string             { return "stub" }
func (s *stubComponent) SetID(_ string)            {}
func (s *stubComponent) String() string            { return s.html }
func (s *stubComponent) Children() []dom.Component { return nil }

func TestRightPanel_RenderHTML_WithAllSlots(t *testing.T) {
	panel := &rightpanel.RightPanel{
		Module:        stubModule{"users"},
		Title:         "Users",
		Head:          &stubComponent{"<span>badge</span>"},
		HeadControls:  &stubComponent{"<select></select>"},
		Article:       &stubComponent{"<table></table>"},
		AsideControls: &stubComponent{"<input type=search>"},
		Aside:         &stubComponent{"<ul></ul>"},
		AsideFooter:   &stubComponent{"<button></button>"},
	}

	el := panel.Render()
	html := el.String()

	checks := []struct {
		label, want string
	}{
		{"wrapper class", "class='rp'"},
		{"main class", "class='rp__main'"},
		{"header class", "class='rp__header'"},
		{"title row", "class='rp__title-row'"},
		{"title class", "class='rp__title'"},
		{"h1 title", "Users</h1>"},
		{"Head slot", "<span>badge</span>"},
		{"HeadControls slot", "<select></select>"},
		{"article class", "class='rp__article'"},
		{"Article slot", "<table></table>"},
		{"aside class", "class='rp__aside'"},
		{"aside header", "class='rp__aside-header'"},
		{"AsideControls slot", "<input type=search>"},
		{"aside content", "class='rp__aside-content'"},
		{"Aside slot", "<ul></ul>"},
		{"aside footer", "class='rp__aside-footer'"},
		{"AsideFooter slot", "<button></button>"},
	}

	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("[%s] expected %q in HTML:\n%s", c.label, c.want, html)
		}
	}
}

func TestRightPanel_RenderHTML_AsideOmittedWhenNil(t *testing.T) {
	panel := &rightpanel.RightPanel{
		Module:  stubModule{"orders"},
		Title:   "Orders",
		Article: &stubComponent{"<table></table>"},
		// No AsideControls, no Aside, no AsideFooter
	}

	html := panel.Render().String()

	if strings.Contains(html, "rp__aside") {
		t.Error("expected rp__aside to be absent when all aside slots are nil")
	}
}

func TestRightPanel_AsideRendersForFooterAlone(t *testing.T) {
	panel := &rightpanel.RightPanel{
		Module:      stubModule{"checkout"},
		Title:       "Checkout",
		AsideFooter: &stubComponent{"<button>Buy</button>"},
	}

	html := panel.Render().String()

	for _, want := range []string{"class='rp__aside'", "class='rp__aside-footer'", "<button>Buy</button>"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in HTML with only AsideFooter set:\n%s", want, html)
		}
	}
}
