package app

import (
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/icons/terminal"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/model"
	"webtyp.com/svg"
)

const idViewAs = "devtools.view_as"

var (
	labelViewAs = lang.Text("View as")
	helpViewAs  = lang.Text("Switch the active module use in a demo or the active role in an app.")
)

// ViewAsConfig configures the «View as» developer tool: ModuleUses in a module demo, or Roles in an app (D3).
type ViewAsConfig struct {
	Uses         []model.ModuleUse
	Roles        []model.Role
	CurrentUse   model.ModuleUse
	CurrentRole  model.RoleCode
	OnSelectUse  func(model.ModuleUse)
	OnSelectRole func(model.RoleCode)
}

type viewAsModule struct {
	dom.Element
	cfg *ViewAsConfig
}

func newViewAsModule(cfg *ViewAsConfig) layout.UIModule {
	return &viewAsModule{cfg: cfg}
}

func (m *viewAsModule) ModelName() string   { return idViewAs }
func (m *viewAsModule) Label() string       { return lang.Translate(labelViewAs).String() }
func (m *viewAsModule) Icon() svg.Icon      { return terminal.Ref }
func (m *viewAsModule) View() dom.Component { return m }
func (m *viewAsModule) Activate()           {}

func (m *viewAsModule) Render() *dom.Element {
	box := html.Div().Class("view-as-panel")
	if m.cfg == nil {
		return box
	}
	for _, u := range m.cfg.Uses {
		use := u
		btn := html.Button().Text(useLabel(use))
		if m.cfg.OnSelectUse != nil {
			btn.OnClick(func(dom.Event) { m.cfg.OnSelectUse(use) })
		}
		box.Child(btn)
	}
	for _, r := range m.cfg.Roles {
		role := r
		btn := html.Button().Text(role.Name)
		if m.cfg.OnSelectRole != nil {
			btn.OnClick(func(dom.Event) { m.cfg.OnSelectRole(role.Code) })
		}
		box.Child(btn)
	}
	return box
}

func useLabel(u model.ModuleUse) string {
	switch u {
	case model.ManagesEverything:
		return "Manages everything"
	case model.WorksWithAllRecords:
		return "Works with all records"
	case model.SeesOwnRecords:
		return "Sees own records"
	case model.VisitsWithoutAccount:
		return "Visits without account"
	default:
		return ""
	}
}
