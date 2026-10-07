//go:build !wasm

package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type langFile struct {
	Default   string              `json:"default"`
	Languages []string            `json:"languages"`
	Keys      map[string][]string `json:"keys"`
}

func TestLangJson(t *testing.T) {
	b, err := os.ReadFile("../lang.json")
	if err != nil {
		t.Fatalf("failed to read lang.json: %v", err)
	}

	var lf langFile
	if err := json.Unmarshal(b, &lf); err != nil {
		t.Fatalf("failed to parse lang.json: %v", err)
	}

	numLangs := len(lf.Languages)
	if numLangs == 0 {
		t.Errorf("expected languages in lang.json")
	}

	for k, v := range lf.Keys {
		if len(v) != numLangs {
			t.Errorf("key %q has %d translations, expected %d", k, len(v), numLangs)
		}
	}

	// Read all go files to check for keys
	goFiles := []string{}
	err = filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk files: %v", err)
	}

	var allCode string
	for _, file := range goFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		allCode += string(content) + "\n"
	}

	for k := range lf.Keys {
		// skip formatting args
		if strings.Contains(k, "%s") {
			continue
		}
		if !strings.Contains(allCode, "\""+k+"\"") {
			t.Errorf("key %q not found in code", k)
		}
	}
}
