package platformd

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/lang"
	"webtyp.com/msgtype"
)

// Badged is an optional capability of a UIModule: a count the nav rail draws
// over the module's entry. The module owns the signals; the chassis only paints.
type Badged interface {
	Badge() *countbadge.CountBadge
}

// Described is an optional capability of a UIModule: what it is, for someone who has never used it, plus
// optional first-steps guidance. A future chassis feature draws it; this interface alone does not change
// anything rendered today.
type Described interface {
	Description() lang.Text // what this module is for, in plain words — never the internal structure
	Steps() []Step          // optional; nil/empty means no step list
}

// Step is one setup prerequisite shown with a checkmark once Done is true.
type Step struct {
	Label lang.Text
	Done  bool
}

// AlertedModule is an optional capability of a UIModule: it has alerts a person has not seen yet. A future
// chassis feature draws them; this interface alone does not change anything rendered today. ModuleAlert and
// this interface are deliberately minimal — alert_manager (a different repo) owns storage, "seen" state and
// delivery; this is only the shape a future chassis feature will need.
type AlertedModule interface {
	UnseenAlerts() []ModuleAlert
}

// ModuleAlert is one row a future alert strip will draw.
type ModuleAlert struct {
	ID      string
	Subject string
	OnView  func()
	OnSeen  func()
}

// Notifier raises a notification in the chassis. *Platform satisfies it.
type Notifier interface {
	Notify(t msgtype.Type, msg string, d Duration)
}

// UsesNotifier is an optional capability of a UIModule that raises
// notifications. The chassis hands itself over once, in Init, before any
// module renders.
type UsesNotifier interface {
	UseNotifier(n Notifier)
}

// GroupModule is an optional capability of a UIModule indicating that it groups
// nested child modules. The chassis renders a downward accordion in the nav rail
// and drawer for group modules.
type GroupModule interface {
	UIModule
	Children() []UIModule
}

// Describe returns m with the Described capability. The result keeps every optional capability
// m already had: it forwards UseNotifier (UsesNotifier) and Badge (Badged) to m when m implements
// them, and it is a GroupModule (forwarding Children) when m is one.
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

func (m *describedModule) UseNotifier(n Notifier) {
	if u, ok := m.UIModule.(UsesNotifier); ok {
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

func (m *describedGroup) UseNotifier(n Notifier) {
	if u, ok := m.GroupModule.(UsesNotifier); ok {
		u.UseNotifier(n)
	}
}

func (m *describedGroup) Badge() *countbadge.CountBadge {
	if b, ok := m.GroupModule.(Badged); ok {
		return b.Badge()
	}
	return nil
}

var _ Notifier = (*Platform)(nil)
