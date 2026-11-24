package settings_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// configHolders are the packages that declare settings. Each reads
// configuration in config.go, and in doctor.go where it has one, and
// takes plain values and setting readers everywhere else.
var configHolders = []string{
	"auth",
	"backup",
	"serve",
	"store",
	"webapp",
	"intake",
	"ops/jobs",
}

const (
	configImport   = "github.com/glw907/dubplate/internal/config"
	settingsImport = "github.com/glw907/dubplate/internal/settings"
)

// readerNames are the settings names a file outside config.go and
// doctor.go may use: the reader type its config.go hands it and the
// error a failed read returns. Any other name, the accessor above all,
// reads configuration.
var readerNames = []string{"Reader", "ReadError"}

// TestConfigurationStaysInConfigFiles holds configuration to config.go
// and doctor.go in each package that declares settings. Another file
// may import internal/settings only to name [settings.Reader] or
// [settings.ReadError], and may not import internal/config at all. A
// package with no config.go that imports the configuration package
// fails too, so a wrong path in the list cannot pass without reading
// anything.
func TestConfigurationStaysInConfigFiles(t *testing.T) {
	for _, pkg := range configHolders {
		t.Run(pkg, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", filepath.FromSlash(pkg), "*.go"))
			if err != nil {
				t.Fatalf("glob %s: %v", pkg, err)
			}
			fset := token.NewFileSet()
			configReads := false
			for _, path := range files {
				name := filepath.Base(path)
				if strings.HasSuffix(name, "_test.go") {
					continue
				}
				file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}
				configFile := name == "config.go" || name == "doctor.go"
				for _, spec := range file.Imports {
					imported, err := strconv.Unquote(spec.Path.Value)
					if err != nil {
						t.Fatalf("%s: import path %s: %v", path, spec.Path.Value, err)
					}
					switch {
					case imported == configImport && name == "config.go":
						configReads = true
					case configFile:
					case imported == configImport:
						t.Errorf("%s imports %s; only config.go and doctor.go read configuration", path, imported)
					case imported == settingsImport:
						for _, used := range settingsNames(file, spec) {
							if !slices.Contains(readerNames, used) {
								t.Errorf("%s uses settings.%s; only config.go and doctor.go read configuration", path, used)
							}
						}
					}
				}
			}
			if !configReads {
				if _, err := os.Stat(filepath.Join("..", filepath.FromSlash(pkg))); err != nil {
					t.Fatalf("package %s: %v", pkg, err)
				}
				t.Errorf("package %s has no config.go importing internal/config", pkg)
			}
		})
	}
}

// settingsNames returns each name file selects from the package spec
// imports, under the name the import binds. A dot or blank import binds
// no selector, so it reports "." or "_" itself, which no allowed name
// matches.
func settingsNames(file *ast.File, spec *ast.ImportSpec) []string {
	local := "settings"
	if spec.Name != nil {
		local = spec.Name.Name
	}
	if local == "." || local == "_" {
		return []string{local}
	}
	var names []string
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == local {
				names = append(names, sel.Sel.Name)
			}
		}
		return true
	})
	return names
}
