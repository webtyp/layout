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

var _ Notifier = (*Platform)(nil)
