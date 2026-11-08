package platformd

import (
	"webtyp.com/layout"
	"webtyp.com/view"
)

type (
	Duration      = view.Duration
	Badged        = layout.Badged
	Described     = layout.Described
	Step          = layout.Step
	AlertedModule = layout.AlertedModule
	ModuleAlert   = layout.ModuleAlert
	Notifier      = view.Notifier
	UsesNotifier  = view.UsesNotifier
	GroupModule   = layout.GroupModule
)

var (
	Auto       = view.Auto
	Persistent = view.Persistent
	For        = view.For
	Describe   = layout.Describe
)

var _ Notifier = (*Platform)(nil)
