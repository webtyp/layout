package layout

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/dom"
	"webtyp.com/svg"
)

// NewUIModule returns an implementation of UIModule with an optional onActivate hook.
func NewUIModule(id, label string, icon svg.Icon, view dom.Component, onActivate ...func()) UIModule {
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
	view       dom.Component
	onActivate func()
}

func (m *uiModule) ModelName() string   { return m.id }
func (m *uiModule) Label() string       { return m.label }
func (m *uiModule) Icon() svg.Icon      { return m.icon }
func (m *uiModule) View() dom.Component { return m.view }
func (m *uiModule) Activate() {
	if m.onActivate != nil {
		m.onActivate()
	} else if act, ok := m.view.(interface{ Activate() }); ok {
		act.Activate()
	}
}

func (m *uiModule) UseNotifier(n Notifier) {
	if u, ok := m.view.(UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *uiModule) Badge() *countbadge.CountBadge {
	if b, ok := m.view.(Badged); ok {
		return b.Badge()
	}
	return nil
}

var _ UIModule = (*uiModule)(nil)
var _ UsesNotifier = (*uiModule)(nil)
var _ Badged = (*uiModule)(nil)

// NewGroupUIModule returns an implementation of GroupModule that wraps child UIModules.
func NewGroupUIModule(id, label string, icon svg.Icon, children []UIModule) UIModule {
	return &uiGroupModule{
		id:       id,
		label:    label,
		icon:     icon,
		children: children,
	}
}

type uiGroupModule struct {
	id       string
	label    string
	icon     svg.Icon
	children []UIModule
}

func (m *uiGroupModule) ModelName() string    { return m.id }
func (m *uiGroupModule) Label() string        { return m.label }
func (m *uiGroupModule) Icon() svg.Icon       { return m.icon }
func (m *uiGroupModule) View() dom.Component  { return nil }
func (m *uiGroupModule) Children() []UIModule { return m.children }
func (m *uiGroupModule) Activate() {
	if len(m.children) > 0 {
		m.children[0].Activate()
	}
}

func (m *uiGroupModule) UseNotifier(n Notifier) {
	for _, ch := range m.children {
		if u, ok := ch.(UsesNotifier); ok {
			u.UseNotifier(n)
		}
	}
}

var _ GroupModule = (*uiGroupModule)(nil)
var _ UsesNotifier = (*uiGroupModule)(nil)
