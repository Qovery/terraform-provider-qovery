//go:build unit && !integration

package main_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const forbiddenImportPrefix = "github.com/hashicorp/terraform-plugin-sdk"

// TestNoTerraformPluginSDKImports guards against reintroducing the deprecated
// terraform-plugin-sdk/v2 test harness. The acceptance tests run on
// terraform-plugin-testing, and mixing both harnesses in one package panics at
// init because each registers the same -sweep flag.
func TestNoTerraformPluginSDKImports(t *testing.T) {
	t.Parallel()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("expected go.mod at module root %s: %v", root, err)
	}

	var offenders []string
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "bin" || name == "vendor" || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, forbiddenImportPrefix) {
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				offenders = append(offenders, rel)
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("QOV-2300: provider tests use terraform-plugin-testing, so %s must not be imported; %d offending files:\n%s",
			forbiddenImportPrefix, len(offenders), strings.Join(offenders, "\n"))
	}
}
