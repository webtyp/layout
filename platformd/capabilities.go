package platformd

import "webtyp.com/layout"

type (
	Duration      = layout.Duration
	Badged        = layout.Badged
	Described     = layout.Described
	Step          = layout.Step
	AlertedModule = layout.AlertedModule
	ModuleAlert   = layout.ModuleAlert
	Notifier      = layout.Notifier
	UsesNotifier  = layout.UsesNotifier
	GroupModule   = layout.GroupModule
)

var (
	Auto       = layout.Auto
	Persistent = layout.Persistent
	For        = layout.For
	Describe   = layout.Describe
)

var _ Notifier = (*Platform)(nil)
