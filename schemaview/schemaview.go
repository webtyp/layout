// Package schemaview is the «Database» screen: the app's database diagram (GET erd.SchemaPath) and the list of
// model errors, each one jumping to its table.
package schemaview

import (
	"webtyp.com/components/searchbar"
	"webtyp.com/erd"
	"webtyp.com/fetch"
	"webtyp.com/icons/database"
	"webtyp.com/json"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
)

// View is a module of the platform shell: proven at compile time.
var _ layout.UIModule = (*View)(nil)

const NameSchemaView widget.Name = "schemaview"

var (
	clsRoot     = NameSchemaView.Root()
	clsHeader   = NameSchemaView.Class("header")
	clsError    = NameSchemaView.Class("error")
	clsBody     = NameSchemaView.Class("body")
	clsFindings = NameSchemaView.Class("findings")
	clsSummary  = NameSchemaView.Class("summary")
	clsFinding  = NameSchemaView.Class("finding")
	clsWhere    = NameSchemaView.Class("finding-where")
	clsWhy      = NameSchemaView.Class("finding-why")
	clsBad      = NameSchemaView.Class("finding-error")
	clsCanvas   = NameSchemaView.Class("canvas")
)

type View struct {
	Element
	errorMsg *SignalString
	search   *SignalString
	filter   *searchbar.SearchBar
	report   *erd.Report
	svgNodes *SignalNodes
	rows     *SignalNodes
}

func New() *View {
	return &View{errorMsg: NewString(""), search: NewString(""), svgNodes: NewNodes(), rows: NewNodes()}
}

func (v *View) ModelName() string       { return string(erd.ResourceSchema) }
func (v *View) Label() string           { return lang.Translate("Database").String() }
func (v *View) Icon() svg.Icon          { return database.Ref }
func (v *View) View() Component         { return v }
func (v *View) WidgetName() widget.Name { return NameSchemaView }
func (v *View) WidgetKind() widget.Kind { return widget.Region }

func (v *View) Init(ctx Ctx) {
	v.filter = &searchbar.SearchBar{Placeholder: lang.Text("Search warnings by table or column")}
	v.filter.OnFilterChange(func(term string) {
		v.search.Set(term)
		v.updateRows()
		v.jumpTo(term)
	})
}

// Activate fetches the diagram each time the screen opens (the schema changes when the app is redeployed).
func (v *View) Activate() {
	v.errorMsg.Set("")
	fetch.Get(erd.SchemaPath).Send(func(resp *fetch.Response, err error) {
		if err != nil {
			v.errorMsg.Set(err.Error())
			return
		}
		if resp.Status == 403 {
			v.errorMsg.Set(lang.Translate("no access to", "database diagram").String())
			return
		}
		if resp.Status != 200 {
			v.errorMsg.Set(lang.Translate("could not load", "database diagram").String() + " " + Sprintf("%d", resp.Status))
			return
		}
		var rep erd.Report
		if err := json.Decode(resp.Body(), &rep); err != nil {
			v.errorMsg.Set(lang.Translate("could not load", "database diagram").String() + ": " + err.Error())
			return
		}
		v.report = &rep
		// The SVG is produced by this app's own server (erd), never by a user: trusted markup. The
		// server has no dictionary, so each band title's word «Module» is translated here.
		svg := Convert(rep.SVG).Replace(">"+string(erd.ModuleLabel)+" ", ">"+lang.Translate(erd.ModuleLabel).String()+" ").String()
		v.svgNodes.Set([]*Element{Div().Raw(Trust(svg))})
		v.updateRows()
	})
}

// updateRows lists the findings whose table or column contains the search term; errors first (erd sorts them).
func (v *View) updateRows() {
	if v.report == nil {
		return
	}
	term := Convert(v.search.Get()).ToLower().String()
	errs, warns := 0, 0
	for _, f := range v.report.Findings {
		if f.Rule.Level() == erd.LevelError {
			errs++
		} else {
			warns++
		}
	}
	// «2 errors · 7 warnings»: how much there is to fix before reading the list.
	summary := Convert(errs).String() + " " + lang.Translate("errors").String() + " · " +
		Convert(warns).String() + " " + lang.Translate("warnings").String()
	rows := []*Element{Div().Set(clsSummary.AsAttr()).Text(summary)}
	for _, f := range v.report.Findings {
		where := f.Table
		if f.Column != "" {
			where += "." + f.Column
		}
		if term != "" && !Contains(Convert(where).ToLower().String(), term) {
			continue
		}
		text := lang.Translate(f.Rule.Text()).String()
		if f.Target != "" {
			text += " (" + f.Target + ")"
		}
		row := Div().Set(clsFinding.AsAttr())
		if f.Rule.Level() == erd.LevelError {
			row.Set(clsBad.AsAttr())
		}
		table := f.Table
		row.Child(Div().Set(clsWhere.AsAttr()).Text(where), Div().Set(clsWhy.AsAttr()).Text(text)).OnClick(func(Event) { v.jumpTo(table) })
		rows = append(rows, row)
	}
	v.rows.Set(rows)
}

// jumpTo scrolls the diagram to the table named exactly name; nothing happens when there is none.
func (v *View) jumpTo(name string) {
	if ref, ok := Get(erd.IDPrefix + name); ok {
		ref.ScrollIntoView()
	}
}

func (v *View) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())
	header := Div().Set(clsHeader.AsAttr())
	if v.filter != nil {
		header.Child(v.filter.Render())
	}
	root.Child(header)
	root.Child(Div().Set(clsError.AsAttr()).BindText(v.errorMsg))
	// The diagram takes the room; the findings are a rail beside it, so clicking one keeps the table in view.
	root.Child(Div().Set(clsBody.AsAttr()).Child(
		Div().Set(clsCanvas.AsAttr()).BindChildren(v.svgNodes),
		Div().Set(clsFindings.AsAttr()).BindChildren(v.rows),
	))
	return root
}
