// Root-level test (justified): pure logic tests must test unexported functions, which the tests/ package cannot access.
package apiexplorer

import (
	"testing"
	"webtyp.com/router"
)

func TestSortingAndFiltering(t *testing.T) {
	e, _ := New(FromTable(router.RouteTable{}))

	e.tableData = &router.RouteTable{
		Routes: []router.RouteRecord{
			{Method: "GET", Path: "/a", Access: "public"},
			{Method: "POST", Path: "/b", Access: "guarded", PolicyKnown: true, Roles: []string{}}, // Orphan
			{Method: "GET", Path: "/c", Access: "guarded", PolicyKnown: true, Roles: []string{"admin"}},
		},
	}

	routes := e.sortedAndFilteredRoutes()
	if len(routes) != 3 {
		t.Fatalf("want 3 routes, got %d", len(routes))
	}
	if routes[0].Path != "/b" {
		t.Errorf("expected orphan /b first, got %s", routes[0].Path)
	}

	e.search.Set("GET")
	routesFiltered := e.sortedAndFilteredRoutes()
	if len(routesFiltered) != 2 {
		t.Fatalf("want 2 routes, got %d", len(routesFiltered))
	}
	if routesFiltered[0].Path != "/a" && routesFiltered[1].Path != "/a" {
		t.Errorf("expected /a in results")
	}
}

// A route that RequiresMFA must say so in the access column: the explorer is
// the security posture of the server, and "guarded" alone hides the second factor.
func TestAccessLabelShowsMFA(t *testing.T) {
	if got := accessLabel(router.RouteRecord{Access: "guarded", MFA: true}); got != "guarded"+labelMFA {
		t.Errorf("MFA route label = %q, want %q", got, "guarded"+labelMFA)
	}
	if got := accessLabel(router.RouteRecord{Access: "public"}); got != "public" {
		t.Errorf("non-MFA route label = %q, want %q", got, "public")
	}
}

func TestNewWithoutSource(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for New(nil), got nil")
	}
	if err.Error() != errNoSource {
		t.Fatalf("expected error '%s', got '%s'", errNoSource, err.Error())
	}
}

func TestFromTableCannotTry(t *testing.T) {
	cfg := FromTable(router.RouteTable{})
	if cfg.canTry {
		t.Error("expected FromTable().canTry to be false")
	}
	srvCfg := FromServer()
	if !srvCfg.canTry {
		t.Error("expected FromServer().canTry to be true")
	}
}

func TestFromTableLoads(t *testing.T) {
	table := router.RouteTable{Routes: []router.RouteRecord{{Method: "GET", Path: "/a"}}}
	cfg := FromTable(table)

	var loaded router.RouteTable
	var loadedErr error
	called := false

	cfg.load(func(t router.RouteTable, err error) {
		loaded = t
		loadedErr = err
		called = true
	})

	if !called {
		t.Fatal("expected load to call done synchronously")
	}
	if loadedErr != nil {
		t.Fatalf("expected no error, got %v", loadedErr)
	}
	if len(loaded.Routes) != 1 || loaded.Routes[0].Path != "/a" {
		t.Errorf("expected loaded route /a, got %v", loaded.Routes)
	}
}
