// Package layout provides pre-built, consistent, reusable UI layout skeletons for
// webtyp modules. A layout defines WHERE elements are placed (structure, grid, spacing)
// but NOT what those elements contain — that is the consumer's responsibility.
package layout

import (
	"webtyp.com/dom"
	"webtyp.com/svg"
)

// Module is the identity contract for the ecosystem.
// Any component that can be identified by a slug (for routing or keys)
// implements this interface.
type Module interface {
	ModelName() string
}

// UIModule is a module that provides its UI to any layout chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	Module               // identity: ModelName() → used as ID
	Label() string       // text in the nav rail
	Icon() svg.Icon      // chassis renders via the sprite
	View() dom.Component // module content
	Activate()           // called by chassis when the module is activated (on-demand loading)
}
