package platformd

import (
	"webtyp.com/components/countbadge"
	"webtyp.com/msgtype"
)

// Badged is an optional capability of a UIModule: a count the nav rail draws
// over the module's entry. The module owns the signals; the chassis only paints.
type Badged interface {
	Badge() *countbadge.CountBadge
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

var _ Notifier = (*Platform)(nil)
