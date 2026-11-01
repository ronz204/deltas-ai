package sources

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/ronz204/deltas-ai"

// reach lists the layers each layer may import; the standard library is always fine, third-party only in adapter.
var reach = map[string][]string{
	"domain":  {"domain"},
	"usecase": {"domain", "usecase"},
	"gateway": {"domain", "usecase", "gateway"},
	"adapter": {"domain", "usecase", "adapter"},
	"harness": {}, // embedded data: it imports nothing from the module
}

// violation says why a package in layer may not import target, or "" when it may.
func violation(layer, target string) string {
	if !strings.Contains(strings.SplitN(target, "/", 2)[0], ".") {
		return ""
	}
	rest, inModule := strings.CutPrefix(target, module+"/")
	if !inModule {
		if layer == "adapter" {
			return ""
		}
		return "third-party imports belong in adapter"
	}
	inner, ok := strings.CutPrefix(rest, "sources/")
	if !ok {
		return "layers do not import scripts"
	}
	other := strings.SplitN(inner, "/", 2)[0]
	if other == "harness" {
		if layer == "gateway" {
			return ""
		}
		return "only gateway reads the embedded harness"
	}
	for _, allowed := range reach[layer] {
		if other == allowed {
			return ""
		}
	}
	return layer + " may not import " + other
}

func TestViolation(t *testing.T) {
	cases := []struct {
		name, layer, target string
		bad                 bool
	}{
		{"stdlib anywhere", "domain", "strings", false},
		{"domain imports domain", "domain", module + "/sources/domain/assertion", false},
		{"domain imports usecase", "domain", module + "/sources/usecase/verify", true},
		{"domain imports gateway", "domain", module + "/sources/gateway/files", true},
		{"domain imports third party", "domain", "github.com/spf13/cobra", true},
		{"usecase imports domain", "usecase", module + "/sources/domain/specfile", false},
		{"usecase imports gateway", "usecase", module + "/sources/gateway/runner", true},
		{"usecase imports adapter", "usecase", module + "/sources/adapter/cmd", true},
		{"gateway implements usecase ports", "gateway", module + "/sources/usecase/verify", false},
		{"gateway reads the harness", "gateway", module + "/sources/harness", false},
		{"gateway imports adapter", "gateway", module + "/sources/adapter/hook", true},
		{"adapter imports usecase", "adapter", module + "/sources/usecase/guard", false},
		{"adapter imports gateway", "adapter", module + "/sources/gateway/gitcli", true},
		{"adapter imports third party", "adapter", "github.com/spf13/cobra", false},
		{"usecase reads the harness", "usecase", module + "/sources/harness", true},
		{"harness imports domain", "harness", module + "/sources/domain/assertion", true},
		{"layers do not import scripts", "adapter", module + "/scripts/delta", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := violation(c.layer, c.target) != ""; got != c.bad {
				t.Fatalf("violation(%s, %s) = %q, want violation: %v", c.layer, c.target, violation(c.layer, c.target), c.bad)
			}
		})
	}
}

func TestLayers(t *testing.T) {
	err := filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
			return err
		}
		parts := strings.Split(filepath.ToSlash(p), "/")
		if len(parts) < 2 {
			return nil // files directly under sources belong to no layer
		}
		layer := parts[0]
		if _, known := reach[layer]; !known {
			t.Errorf("%s: %q is not a layer; use domain, usecase, gateway or adapter", p, layer)
			return nil
		}
		file, perr := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range file.Imports {
			target, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				return uerr
			}
			if why := violation(layer, target); why != "" {
				t.Errorf("%s imports %s: %s", filepath.ToSlash(p), target, why)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
