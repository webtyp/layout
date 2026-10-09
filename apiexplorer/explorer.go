package apiexplorer

import (
	"webtyp.com/components/searchbar"
	"webtyp.com/fetch"
	"webtyp.com/json"
	"webtyp.com/router"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/svg"
)

// Explorer is a module of the platform shell: proven at compile time.
var _ platformd.UIModule = (*Explorer)(nil)

const NameAPIExplorer widget.Name = "apiexplorer"

var (
	clsRoot        = NameAPIExplorer.Root()
	clsHeader      = NameAPIExplorer.Class("header")
	clsTable       = NameAPIExplorer.Class("table")
	clsRow         = NameAPIExplorer.Class("row")
	clsRowExpanded = NameAPIExplorer.Class("row-expanded")
	clsOrphan      = NameAPIExplorer.Class("orphan")
	clsPublic      = NameAPIExplorer.Class("public")

	clsColMethod     = NameAPIExplorer.Class("col-method")
	clsColPath       = NameAPIExplorer.Class("col-path")
	clsColAccess     = NameAPIExplorer.Class("col-access")
	clsColPermission = NameAPIExplorer.Class("col-permission")
	clsColRoles      = NameAPIExplorer.Class("col-roles")

	clsError = NameAPIExplorer.Class("error")

	clsTryForm = NameAPIExplorer.Class("try-form")
)

type Config struct {
	RoutesURL string
}

type Explorer struct {
	Element
	cfg Config

	routesURL string

	tableData *router.RouteTable
	errorMsg  *SignalString

	filter *searchbar.SearchBar
	search *SignalString

	expandedRow *SignalString
	rowsNodes   *SignalNodes
}

func New(cfg Config) (*Explorer, error) {
	e := &Explorer{
		cfg:         cfg,
		errorMsg:    NewString(""),
		search:      NewString(""),
		expandedRow: NewString(""),
		rowsNodes:   NewNodes(),
	}
	if e.cfg.RoutesURL == "" {
		e.routesURL = router.IntrospectionPath
	} else {
		e.routesURL = e.cfg.RoutesURL
	}
	return e, nil
}

func (e *Explorer) ModelName() string { return "api_explorer" }
func (e *Explorer) Label() string     { return lang.Translate("API Explorer").String() }
func (e *Explorer) Icon() svg.Icon    { return svg.Icon("terminal") }
func (e *Explorer) View() Component   { return e }

func (e *Explorer) WidgetName() widget.Name { return NameAPIExplorer }
func (e *Explorer) WidgetKind() widget.Kind { return widget.Region }

func (e *Explorer) Init(ctx Ctx) {
	e.filter = &searchbar.SearchBar{Placeholder: lang.Text("Search routes")}
	e.filter.OnFilterChange(func(term string) {
		e.search.Set(term)
		e.updateRows()
	})
}

// Activate is called by the shell when the module is opened: only then is /_routes fetched,
// and each activation refreshes it (routes and roles change while the app runs).
func (e *Explorer) Activate() {
	e.errorMsg.Set("")
	e.load()
}

func (e *Explorer) load() {
	req := fetch.Get(e.routesURL)
	req.Send(func(resp *fetch.Response, err error) {
		if err != nil {
			e.errorMsg.Set(err.Error())
			return
		}
		if resp.Status != 200 {
			if resp.Status == 403 {
				e.errorMsg.Set(lang.Translate("you need permission to view the API").String())
			} else {
				e.errorMsg.Set(lang.Translate("Error").String() + " " + Sprintf("%d", resp.Status))
			}
			return
		}

		var table router.RouteTable
		err = json.Decode(resp.Body(), &table)
		if err != nil {
			e.errorMsg.Set(lang.Translate("Failed to decode routes").String() + ": " + err.Error())
			return
		}
		e.tableData = &table
		e.updateRows()
	})
}

func (e *Explorer) sortedAndFilteredRoutes() []router.RouteRecord {
	if e.tableData == nil {
		return nil
	}

	term := Convert(e.search.Get()).ToLower().String()

	var orphans []router.RouteRecord
	var others []router.RouteRecord

	for _, r := range e.tableData.Routes {
		if term != "" {
			sub := Convert(r.Method + " " + r.Path).ToLower().String()
			if !Contains(sub, term) {
				continue
			}
		}
		if r.Orphan() {
			orphans = append(orphans, r)
		} else {
			others = append(others, r)
		}
	}

	return append(orphans, others...)
}

func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	b := make([]byte, n)
	bp := copy(b, elems[0])
	for _, s := range elems[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}

func (e *Explorer) updateRows() {
	routes := e.sortedAndFilteredRoutes()
	var nodes []*Element

	expandedID := e.expandedRow.Get()

	for _, r := range routes {
		rowID := r.Method + ":" + r.Path

		isExpanded := expandedID == rowID

		row := Div().Set(clsRow.AsAttr())

		if r.Orphan() {
			row.Set(clsOrphan.AsAttr())
		}
		if r.Access == model.AccessPublic.String() {
			row.Set(clsPublic.AsAttr())
		}

		if isExpanded {
			row.Set(clsRowExpanded.AsAttr())
		}

		row.OnClick(func(evt Event) {
			if e.expandedRow.Get() == rowID {
				e.expandedRow.Set("")
			} else {
				e.expandedRow.Set(rowID)
			}
			e.updateRows()
		})

		methodCol := Div().Set(clsColMethod.AsAttr()).Text(r.Method)
		pathCol := Div().Set(clsColPath.AsAttr()).Text(r.Path)
		accessCol := Div().Set(clsColAccess.AsAttr()).Text(r.Access)

		permStr := ""
		if r.Resource != "" && r.Action != "" {
			permStr = r.Resource + ":" + r.Action
		}
		permCol := Div().Set(clsColPermission.AsAttr()).Text(permStr)

		rolesStr := ""
		if !r.PolicyKnown {
			rolesStr = "—" // with tooltip "the server did not describe its policy"
		} else {
			if len(r.Roles) == 0 {
				rolesStr = lang.Translate("nobody").String()
			} else {
				rolesStr = joinStrings(r.Roles, ", ")
			}
		}
		rolesCol := Div().Set(clsColRoles.AsAttr()).Text(rolesStr)
		if !r.PolicyKnown {
			rolesCol.Attr("title", lang.Translate("the server did not describe its policy").String())
		}

		headerRow := Div().Child(methodCol, pathCol, accessCol, permCol, rolesCol)
		row.Child(headerRow)

		if isExpanded {
			tryPanel := buildTryPanel(r)
			row.Child(tryPanel)
		}

		nodes = append(nodes, row)
	}

	e.rowsNodes.Set(nodes)
}

func (e *Explorer) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())

	header := Div().Set(clsHeader.AsAttr())
	if e.filter != nil {
		header.Child(e.filter.Render())
	}
	root.Child(header)

	errDiv := Div().Set(clsError.AsAttr()).BindText(e.errorMsg)
	root.Child(errDiv)

	tableContainer := Div().Set(clsTable.AsAttr()).BindChildren(e.rowsNodes)
	root.Child(tableContainer)

	return root
}
