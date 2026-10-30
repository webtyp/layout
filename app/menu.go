package app

import (
	"webtyp.com/ddl"
	"webtyp.com/erd"
	"webtyp.com/fmt"
	"webtyp.com/icons/terminal"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/layout/apiexplorer"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/schemaview"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/rpc"
	"webtyp.com/svg"
	"webtyp.com/view"
)

var (
	helpAPIExplorer = lang.Text("Every operation of this app: who may call it, and a form to try it.")
	helpSchemaView  = lang.Text("Every table of this app, its references and the model errors to fix.")
)

// idDevTools is the menu id of the «Developer tools» group. A group has no permission of its own.
const idDevTools = "devtools"

const (
	labelDevTools lang.Text = "Developer tools"
	helpDevTools  lang.Text = "Tools to inspect how this app is built."
)

// Tools are the developer screens under «Developer tools». A nil field leaves
// that screen out; Tools{} leaves the whole group out — an app with no API and
// no database shows no developer tools.
// Example (an app with a server): app.Tools{API: apiexplorer.FromServer(), Database: schemaview.FromServer()}
type Tools struct {
	API      *apiexplorer.Config
	Database *schemaview.Source
}

// LocalTools builds the tools of a page with no server — a module's demo — with
// the pieces the server uses: the routes its operations declare (rpc.Routes)
// and its migration steps (erd.ReadSteps). No modules leaves the explorer out;
// no steps leaves the diagram out.
// Example: tools, err := app.LocalTools([]router.OperationModule{mod}, ddl.Step{Name: "shop", Run: migrate.Migrate})
func LocalTools(modules []router.OperationModule, steps ...ddl.Step) (Tools, error) {
	var t Tools
	if len(modules) > 0 {
		routes, err := rpc.Routes(rpc.DefaultPrefix, modules...)
		if err != nil {
			return Tools{}, err
		}
		t.API = apiexplorer.FromTable(router.Table(routes, nil))
	}
	if len(steps) > 0 {
		d, err := erd.ReadSteps(steps...)
		if err != nil {
			return Tools{}, err
		}
		t.Database = schemaview.FromReport(d.Report())
	}
	return t, nil
}

// Menu is what layout/app hands to any layout chassis (such as platformd or a custom layout):
// the modules and the visibility rule.
type Menu struct {
	Modules []layout.UIModule
	CanView func(id string) bool
}

type screenPerm struct {
	id         string
	permission model.Resource
}

const errNoOptionSource = "layout/app: a field takes its options from a model no screen lists: "

// leafView pairs a built list screen with its CrudView, so selects can be wired after
// every screen exists. buildScreen appends one per list leaf.
type leafView struct {
	view *crudview.CrudView
	pres view.Presenter
}

// wireOptions fills every select of every list screen from the presenter that lists the
// field's Ref/Options model, and keeps it in step: the source's OnChange refills it.
func wireOptions(leaves []leafView) error {
	for _, target := range leaves {
		for _, f := range target.pres.Record().Schema() {
			src := f.Options
			if src == nil {
				src = f.Ref
			}
			if src == nil {
				continue
			}
			source, ok := sourceFor(leaves, src.Name)
			if !ok {
				return fmt.Err(errNoOptionSource + target.pres.Record().ModelName() + "." + f.Name + " → " + src.Name)
			}
			fill := func(v *crudview.CrudView, field string, s view.Presenter) func() {
				return func() {
					items := s.Items()
					opts := make([]fmt.KeyValue, 0, len(items))
					for _, it := range items {
						opts = append(opts, fmt.KeyValue{Key: it.ID, Value: it.Label})
					}
					v.SetOptions(field, opts...)
				}
			}(target.view, f.Name, source)
			source.OnChange(fill)
			fill()
		}
	}
	return nil
}

func sourceFor(leaves []leafView, modelName string) (view.Presenter, bool) {
	for _, l := range leaves {
		if l.pres.Record().ModelName() == modelName {
			return l.pres, true
		}
	}
	return nil, false
}

// BuildMenu validates every screen and builds its layout.UIModule tree. allows answers whether the
// signed-in user holds a permission (in an app: func(r model.Resource) bool { return profile.Can(r, model.Read) }).
// tools configures the «Developer tools» group, added last; the group appears only with at least one tool.
func BuildMenu(screens []view.Screen, ids model.IDGenerator, allows func(model.Resource) bool, tools Tools) (Menu, error) {
	for _, s := range screens {
		if err := s.Validate(); err != nil {
			return Menu{}, err
		}
	}

	var perms []screenPerm
	leaves := make([]leafView, 0)
	mods := make([]layout.UIModule, 0, len(screens))
	for _, s := range screens {
		m, err := buildScreen(s, ids, &perms, &leaves)
		if err != nil {
			return Menu{}, err
		}
		mods = append(mods, m)
	}

	if err := wireOptions(leaves); err != nil {
		return Menu{}, err
	}

	var devs []layout.UIModule
	if tools.API != nil {
		explorer, err := apiexplorer.New(tools.API)
		if err != nil {
			return Menu{}, err
		}
		devs = append(devs, layout.Describe(explorer, helpAPIExplorer))
		perms = append(perms, screenPerm{id: explorer.ModelName(), permission: router.ResourceIntrospection})
	}
	if tools.Database != nil {
		schema, err := schemaview.New(tools.Database)
		if err != nil {
			return Menu{}, err
		}
		devs = append(devs, layout.Describe(schema, helpSchemaView))
		perms = append(perms, screenPerm{id: schema.ModelName(), permission: erd.ResourceSchema})
	}
	if len(devs) > 0 {
		devtools := layout.NewGroupUIModule(idDevTools, lang.Translate(labelDevTools).String(), terminal.Ref, devs)
		mods = append(mods, layout.Describe(devtools, helpDevTools))
	}

	canView := func(id string) bool {
		if allows == nil {
			return false
		}
		for _, p := range perms {
			if p.id == id {
				return allows(p.permission)
			}
		}
		return false
	}

	return Menu{
		Modules: mods,
		CanView: canView,
	}, nil
}

func buildScreen(s view.Screen, ids model.IDGenerator, perms *[]screenPerm, leaves *[]leafView) (layout.UIModule, error) {
	if len(s.Children) > 0 {
		children := make([]layout.UIModule, 0, len(s.Children))
		for _, c := range s.Children {
			ch, err := buildScreen(c, ids, perms, leaves)
			if err != nil {
				return nil, err
			}
			children = append(children, ch)
		}
		grp := layout.NewGroupUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), children)
		return layout.Describe(grp, s.Help), nil
	}

	if s.Component != nil {
		*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
		m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), s.Component)
		return layout.Describe(m, s.Help), nil
	}

	v, err := crudview.New(crudview.Config{
		ParentID:  s.ID,
		Presenter: s.Presenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}

	*leaves = append(*leaves, leafView{view: v, pres: s.Presenter})

	activate := func() {
		for _, f := range s.Presenter.Record().Schema() {
			src := f.Options
			if src == nil {
				src = f.Ref
			}
			if src == nil {
				continue
			}
			if _, isInput := f.Type.(input.Input); !isInput {
				continue
			}
			source, ok := sourceFor(*leaves, src.Name)
			if ok {
				source.Reload(nil)
			}
		}
		v.Activate()
	}

	*perms = append(*perms, screenPerm{id: s.ID, permission: s.Permission})
	m := layout.NewUIModule(s.ID, lang.Translate(s.Label).String(), svg.Icon(s.Icon), v, activate)
	if desc := s.Presenter.Description(); desc != "" {
		m = layout.Describe(m, desc)
	}
	return m, nil
}
