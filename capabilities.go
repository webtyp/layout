package layout

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/fmt"
	"webtyp.com/lang"
	"webtyp.com/msgtype"
)

// Duration is how long a notification stays before dismissing itself.
type Duration struct {
	millis func(msg string) int // nil = persistent
	auto   bool                 // built by Auto(): the caller left the length to the chassis
}

// IsAuto reports whether the duration was created with Auto().
func (d Duration) IsAuto() bool { return d.auto }

// HasTimer reports whether the duration auto-dismisses after a positive number of milliseconds.
func (d Duration) HasTimer() bool { return d.millis != nil }

// Millis computes the duration in milliseconds for msg, or 0 if persistent.
func (d Duration) Millis(msg string) int {
	if d.millis == nil {
		return 0
	}
	return d.millis(msg)
}

// Auto sizes the duration to the message (~350ms per word plus 2.5s, floored at 4s and capped at 10s).
func Auto() Duration {
	return Duration{auto: true, millis: func(msg string) int {
		words := fmt.Count(msg, " ") + 1
		ms := 2500 + words*350
		if ms < 4000 {
			return 4000
		}
		if ms > 10000 {
			return 10000
		}
		return ms
	}}
}

// Persistent keeps the notification until the user dismisses it.
func Persistent() Duration { return Duration{} }

// For pins an exact duration in milliseconds.
func For(ms int) Duration {
	return Duration{millis: func(string) int { return ms }}
}

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

// Notifier raises a notification in the chassis.
type Notifier interface {
	Notify(t msgtype.Type, msg string, d Duration)
}

// UsesNotifier is an optional capability of a UIModule that raises notifications.
type UsesNotifier interface {
	UseNotifier(n Notifier)
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
