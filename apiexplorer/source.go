package apiexplorer

import (
	"webtyp.com/fetch"
	"webtyp.com/json"
	"webtyp.com/lang"
	"webtyp.com/router"

	. "webtyp.com/fmt"
)

const errNoSource = "apiexplorer: no source: use FromServer or FromTable"

// Config says where the explorer reads its route table and whether a route can
// be tried. Build it with FromServer or FromTable.
type Config struct {
	load   func(done func(router.RouteTable, error))
	canTry bool
}

// FromServer reads the table the app's server serves at router.IntrospectionPath;
// each route can be tried against that server.
// Example: app.Tools{API: apiexplorer.FromServer()}
func FromServer() *Config {
	return &Config{load: fetchRoutes(router.IntrospectionPath), canTry: true}
}

// FromTable shows a table built in process — a module's demo, which has no
// server, so routes cannot be tried.
// Example: apiexplorer.FromTable(router.Table(routes, nil))
func FromTable(t router.RouteTable) *Config {
	return &Config{load: func(done func(router.RouteTable, error)) { done(t, nil) }}
}

func fetchRoutes(url string) func(done func(router.RouteTable, error)) {
	return func(done func(router.RouteTable, error)) {
		fetch.Get(url).Send(func(resp *fetch.Response, err error) {
			if err != nil {
				done(router.RouteTable{}, err)
				return
			}
			if resp.Status == 403 {
				done(router.RouteTable{}, Err(lang.Translate("you need permission to view the API").String()))
				return
			}
			if resp.Status != 200 {
				done(router.RouteTable{}, Err(lang.Translate("Error").String()+" "+Sprintf("%d", resp.Status)))
				return
			}
			var table router.RouteTable
			if err := json.Decode(resp.Body(), &table); err != nil {
				done(router.RouteTable{}, Err(lang.Translate("Failed to decode routes").String()+": "+err.Error()))
				return
			}
			done(table, nil)
		})
	}
}
