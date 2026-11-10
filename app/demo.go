//go:build wasm

package app

import (
	"webtyp.com/crud"
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/time"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

// Demo draws a module's demo with the chassis: the module's operations through a loopback signed in as
// Demo.ActingAs, its screens, and the developer tools from its operations and migrations. It never returns.
// The webtyp daemon generates the only caller:  func main() { layoutapp.Demo(seed.Demo) }
func Demo(build func() (view.Demo, error)) {
	d, err := build()
	if err == nil {
		err = d.Validate()
	}
	if err != nil {
		Append("body", P().Text(err.Error())) // a broken demo says why on the page, never a blank screen
		select {}
	}
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err) // never fails in the browser
	}
	zone, err := time.LoadZone(d.Timezone)
	if err != nil {
		Append("body", P().Text(err.Error()))
		select {}
	}
	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, d.ActingAs, d.Modules...))
	screens := d.Screens(view.BrowserDeps{Caller: caller, Zone: zone})
	tools, err := LocalTools(d.Modules, d.Migrations...)
	if err == nil {
		var menu Menu
		menu, err = BuildMenu(screens, ids, func(model.Resource) bool { return true }, tools)
		if err == nil {
			p := &platformd.Platform{
				AppName: d.Label + " — demo",
				Modules: menu.Modules,
				CanView: menu.CanView,
				User:    demoUser{id: d.ActingAs},
			}
			p.Init(nil)
			Append("body", p)
			select {}
		}
	}
	Append("body", P().Text(err.Error()))
	select {}
}

type demoUser struct{ id string }

func (u demoUser) UserName() string    { return u.id }
func (u demoUser) UserAvatar() string  { return "" }
func (u demoUser) UserRoles() []string { return nil }
