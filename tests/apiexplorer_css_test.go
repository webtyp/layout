//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/layout/apiexplorer"
)

// declsFor joins the declarations of every rule whose selector list names sel
// exactly, across all layers. The primitives layer groups selectors that share
// a recipe (".a, .b, .c { ... }"), so a substring search on one rule is not
// enough: the flag a part carries may live in a grouped rule.
func declsFor(cssStr, sel string) string {
	var out strings.Builder
	rest := cssStr
	for {
		open := strings.Index(rest, "{")
		if open == -1 {
			break
		}
		close := strings.Index(rest[open:], "}")
		if close == -1 {
			break
		}
		header := rest[:open]
		if nl := strings.LastIndexAny(header, "\n;}"); nl != -1 {
			header = header[nl+1:]
		}
		body := rest[open+1 : open+close]
		if strings.Contains(body, "{") {
			// A layer block opens here: step inside it.
			rest = rest[open+1:]
			continue
		}
		for _, s := range strings.Split(header, ",") {
			if strings.TrimSpace(s) == sel {
				out.WriteString(body)
			}
		}
		rest = rest[open+close+1:]
	}
	return out.String()
}

// TestAPIExplorer_RowsKeepTheirHeight guards the overlap seen in iam: the
// explorer fills the stage, the table is a flex column inside it, and a Stack
// emits min-height: 0. Without a scroll region and rows that refuse to shrink,
// 25 rows that need 32px each were squeezed into 18px and their text overlapped.
func TestAPIExplorer_RowsKeepTheirHeight(t *testing.T) {
	e, err := apiexplorer.New(apiexplorer.Config{})
	if err != nil {
		t.Fatal(err)
	}
	cssStr := e.RenderCSS().String()

	if d := declsFor(cssStr, ".apiexplorer__table"); !strings.Contains(d, "overflow-y: auto;") {
		t.Errorf(".apiexplorer__table must be the scroll region, decls:\n%s", d)
	}
	if d := declsFor(cssStr, ".apiexplorer__row"); !strings.Contains(d, "flex-shrink: 0;") {
		t.Errorf(".apiexplorer__row must not shrink below its content, decls:\n%s", d)
	}
	if d := declsFor(cssStr, ".apiexplorer__header"); !strings.Contains(d, "flex-shrink: 0;") {
		t.Errorf(".apiexplorer__header must not shrink when the table scrolls, decls:\n%s", d)
	}
	// The search bar spans the header: a Stack gives its children full width,
	// a Row left it at its 120px intrinsic width.
	if d := declsFor(cssStr, ".apiexplorer__header"); !strings.Contains(d, "flex-direction: column;") {
		t.Errorf(".apiexplorer__header must stack so the search bar spans it, decls:\n%s", d)
	}
}
