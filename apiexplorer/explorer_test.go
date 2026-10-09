// Root-level test (justified): pure logic tests must test unexported functions, which the tests/ package cannot access.
package apiexplorer

import (
	"testing"
	"webtyp.com/router"
)

func TestSortingAndFiltering(t *testing.T) {
	e, _ := New(Config{})

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
