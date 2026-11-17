package tests

import (
	"testing"

	"webtyp.com/layout/app"
	"webtyp.com/model"
)

func TestMenu_ViewAsListsModuleUses(t *testing.T) {
	var selected model.ModuleUse
	tools := app.Tools{
		ViewAs: &app.ViewAsConfig{
			Uses: []model.ModuleUse{model.ManagesEverything, model.SeesOwnRecords},
			OnSelectUse: func(u model.ModuleUse) {
				selected = u
			},
		},
	}
	menu, err := app.BuildMenu(nil, nil, func(model.Resource) bool { return true }, tools)
	if err != nil {
		t.Fatal(err)
	}
	if len(menu.Modules) != 1 {
		t.Fatalf("expected devtools group in menu, got %d modules", len(menu.Modules))
	}
	tools.ViewAs.OnSelectUse(model.SeesOwnRecords)
	if selected != model.SeesOwnRecords {
		t.Fatalf("expected SeesOwnRecords, got %v", selected)
	}
}
