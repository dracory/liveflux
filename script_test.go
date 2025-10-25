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
	endpoint := DefaultEndpoint
	modules := []string{
		strings.TrimSpace(injectNamespaceConstants(readJS(t, "liveflux_namespace_create.js"), endpoint)),
		strings.TrimSpace(readJS(t, "liveflux_util.js")),
		strings.TrimSpace(readJS(t, "liveflux_events.js")),
		strings.TrimSpace(readJS(t, "liveflux_network.js")),
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

func TestInjectNamespaceConstants(t *testing.T) {
	endpoint := "/custom"
	js := injectNamespaceConstants(livefluxNamespaceCreateJS, endpoint)

	constants := []string{
		DataFluxAction,
		DataFluxDispatchTo,
		DataFluxComponent,
		DataFluxComponentID,
		DataFluxID,
		DataFluxMount,
		DataFluxParam,
		DataFluxRoot,
		DataFluxSubmit,
		DataFluxWS,
		DataFluxWSURL,
	}

	for _, c := range constants {
		if !strings.Contains(js, c) {
			t.Fatalf("namespace constants missing %q", c)
		}
	}

	if !strings.Contains(js, endpoint) {
		t.Fatalf("namespace script missing endpoint %q", endpoint)
	}

	expecteds := []string{
		// constants
		`const dataFluxAction = "` + DataFluxAction + `";`,
		`const dataFluxDispatchTo = "` + DataFluxDispatchTo + `";`,
		`const dataFluxComponent = "` + DataFluxComponent + `";`,
		`const dataFluxComponentID = "` + DataFluxComponentID + `";`,
		`const dataFluxID = "` + DataFluxID + `";`,
		`const dataFluxMount = "` + DataFluxMount + `";`,
		`const dataFluxParam = "` + DataFluxParam + `";`,
		`const dataFluxRoot = "` + DataFluxRoot + `";`,
		`const dataFluxSubmit = "` + DataFluxSubmit + `";`,
		`const dataFluxWS = "` + DataFluxWS + `";`,
		`const dataFluxWSURL = "` + DataFluxWSURL + `";`,
		`const endpoint = "` + endpoint + `";`,

		// added to window.liveflux object
		`dataFluxAction,`,
		`dataFluxDispatchTo,`,
		`dataFluxComponent,`,
		`dataFluxComponentID,`,
		`dataFluxID,`,
		`dataFluxMount,`,
		`dataFluxParam,`,
		`dataFluxRoot,`,
		`dataFluxSubmit,`,
		`dataFluxWS,`,
		`dataFluxWSURL,`,
		`endpoint,`,
	}

	for _, e := range expecteds {
		if !strings.Contains(js, e) {
			t.Fatalf("namespace constants missing %q", e)
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
