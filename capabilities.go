package layout

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/lang"
	"webtyp.com/view"
)

// Badged is an optional capability of a UIModule: a count the nav rail draws
// over the module's entry.
type Badged interface {
	Badge() *countbadge.CountBadge
}

// Described is an optional capability of a UIModule: what it is, for someone who has never used it, plus
// optional first-steps guidance.
type Described interface {
	Description() lang.Text
	Steps() []Step
}

// Step is one setup prerequisite shown with a checkmark once Done is true.
type Step struct {
	Label lang.Text
	Done  bool
}

// AlertedModule is an optional capability of a UIModule: it has alerts a person has not seen yet.
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}

// ModuleAlert is one row an alert strip draws.
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}

// GroupModule is an optional capability of a UIModule indicating that it groups
// nested child modules.
type GroupModule interface {
	UIModule
	Children() []UIModule
}

// Describe returns m with the Described capability, preserving UsesNotifier, Badged, and GroupModule.
func Describe(m UIModule, text lang.Text, steps ...Step) UIModule {
	if gm, ok := m.(GroupModule); ok {
		return &describedGroup{
			GroupModule: gm,
			text:        text,
			steps:       steps,
		}
	}
	return &describedModule{
		UIModule: m,
		text:     text,
		steps:    steps,
	}
}

type describedModule struct {
	UIModule
	text  lang.Text
	steps []Step
}

func (m *describedModule) Description() lang.Text { return m.text }
func (m *describedModule) Steps() []Step          { return m.steps }

func (m *describedModule) UseNotifier(n view.Notifier) {
	if u, ok := m.UIModule.(view.UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *describedModule) Badge() *countbadge.CountBadge {
	if b, ok := m.UIModule.(Badged); ok {
		return b.Badge()
	}
	return nil
}

type describedGroup struct {
	GroupModule
	text  lang.Text
	steps []Step
}

func (m *describedGroup) Description() lang.Text { return m.text }
func (m *describedGroup) Steps() []Step          { return m.steps }

func (m *describedGroup) UseNotifier(n view.Notifier) {
	if u, ok := m.GroupModule.(view.UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *describedGroup) Badge() *countbadge.CountBadge {
	if b, ok := m.GroupModule.(Badged); ok {
		return b.Badge()
	}
	return nil
}

