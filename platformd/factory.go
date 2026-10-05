package platformd

import (
	. "webtyp.com/dom"
	"webtyp.com/svg"
)

// NewUIModule returns an implementation of UIModule with an optional onActivate hook.
func NewUIModule(id, label string, icon svg.Icon, view Component, onActivate ...func()) UIModule {
	var act func()
	if len(onActivate) > 0 {
		act = onActivate[0]
	}
	return &uiModule{
		id:         id,
		label:      label,
		icon:       icon,
		view:       view,
		onActivate: act,
	}
}

type uiModule struct {
	id         string
	label      string
	icon       svg.Icon
	view       Component
	onActivate func()
}

func (m *uiModule) ModelName() string { return m.id }
func (m *uiModule) Label() string     { return m.label }
func (m *uiModule) Icon() svg.Icon    { return m.icon }
func (m *uiModule) View() Component   { return m.view }
func (m *uiModule) Activate() {
	if m.onActivate != nil {
		m.onActivate()
	} else if act, ok := m.view.(interface{ Activate() }); ok {
		act.Activate()
	}
}
