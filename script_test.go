package liveflux

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dracory/hb"
)

// helper to read a file relative to this package directory
func readJS(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("js", rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func TestJSConcatenationOrder(t *testing.T) {
	modules := []string{
		strings.TrimSpace(readJS(t, "liveflux_util.js")),
		strings.TrimSpace(readJS(t, "liveflux_events.js")),
		strings.TrimSpace(readJS(t, "liveflux_network.js")),
		strings.TrimSpace(readJS(t, "liveflux_target.js")),
		strings.TrimSpace(readJS(t, "liveflux_triggers.js")),
		strings.TrimSpace(readJS(t, "liveflux_wire.js")),
		strings.TrimSpace(readJS(t, "liveflux_mount.js")),
		strings.TrimSpace(readJS(t, "liveflux_handlers.js")),
		strings.TrimSpace(readJS(t, "liveflux_bootstrap.js")),
		strings.TrimSpace(readJS(t, "liveflux_find.js")),
		strings.TrimSpace(readJS(t, "liveflux_dispatch.js")),
	}

	expected := strings.Join(modules, "\n")

	got := JS()
	// JS() now prepends a single-line config snippet ending with a newline.
	// Verify the payload after the first newline matches the expected concatenation.
	idx := strings.IndexByte(got, '\n')
	if idx < 0 {
		t.Fatalf("JS() did not contain expected config prefix with newline")
	}
	payload := got[idx+1:]
	if payload != expected {
		t.Fatalf("JS() content mismatch after config prefix")
	}
}

func TestJSIncludesConfigVariables(t *testing.T) {
	out := JS()

	configVariables := map[string]string{
		"dataFluxAction":        DataFluxAction,
		"dataFluxDispatchTo":    DataFluxDispatchTo,
		"dataFluxComponentKind": DataFluxComponentKind,
		"dataFluxComponentID":   DataFluxComponentID,
		"dataFluxMount":         DataFluxMount,
		"dataFluxParam":         DataFluxParam,
		"dataFluxSubmit":        DataFluxSubmit,
		"dataFluxWS":            DataFluxWS,
		"dataFluxWSURL":         DataFluxWSURL,
		"endpoint":              DefaultEndpoint,
		"redirectHeader":        RedirectHeader,
		"redirectAfterHeader":   RedirectAfterHeader,
		"credentials":           "",
	}

	for key, val := range configVariables {
		needle := `"` + key + `":"` + val + `"`
		if !strings.Contains(out, needle) {
			t.Fatalf("JS() config missing variable %q (%s)", key, val)
		}
	}

	extraChecks := map[string]string{
		"useWebSocket": "false",
		"headers":      "{}",
		"timeoutMs":    "0",
	}

	for key, val := range extraChecks {
		needle := `"` + key + `":` + val
		if !strings.Contains(out, needle) {
			t.Fatalf("JS() config missing variable %q (%s)", key, val)
		}
	}
}

func TestScriptWrapper(t *testing.T) {
	// Script() should be equivalent to hb.Script(JS())
	expected := hb.Script(JS()).ToHTML()
	got := Script().ToHTML()
	if got != expected {
		t.Fatalf("Script().ToHTML() mismatch with hb.Script(JS())")
	}
}

func TestJSIncludesRedirectHeadersConfig(t *testing.T) {
	out := JS()
	// The config JSON is embedded in the first line; assert it includes the header names.
	if !strings.Contains(out, `"redirectHeader":"`+RedirectHeader+`"`) {
		t.Fatalf("JS() missing redirectHeader config; out: %q", out[:min(200, len(out))])
	}
	if !strings.Contains(out, `"redirectAfterHeader":"`+RedirectAfterHeader+`"`) {
		t.Fatalf("JS() missing redirectAfterHeader config; out: %q", out[:min(200, len(out))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
