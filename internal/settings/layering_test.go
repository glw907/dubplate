package settings_test

import (
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
// takes plain values everywhere else.
var configHolders = []string{
	"auth",
	"backup",
	"serve",
	"store",
	"webapp",
	"intake",
	"ops/local",
}

// configImports are the packages whose import marks a file as reading
// configuration.
var configImports = []string{
	"github.com/glw907/dubplate/internal/config",
	"github.com/glw907/dubplate/internal/settings",
}

// TestConfigurationStaysInConfigFiles holds configuration to config.go
// and doctor.go in each package that declares settings. A package with
// no config.go that imports the configuration package fails too, so a
// wrong path in the list cannot pass without reading anything.
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
				file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}
				for _, spec := range file.Imports {
					imported, err := strconv.Unquote(spec.Path.Value)
					if err != nil {
						t.Fatalf("%s: import path %s: %v", path, spec.Path.Value, err)
					}
					if !slices.Contains(configImports, imported) {
						continue
					}
					if name != "config.go" && name != "doctor.go" {
						t.Errorf("%s imports %s; only config.go and doctor.go read configuration", path, imported)
					}
					if name == "config.go" && imported == configImports[0] {
						configReads = true
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
