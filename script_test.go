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
	// Build the expected concatenation of embedded client JS parts
	expected := readJS(t, "util.js") + "\n" +
		readJS(t, "network.js") + "\n" +
		readJS(t, "events.js") + "\n" +
		readJS(t, "wire.js") + "\n" +
		readJS(t, "mount.js") + "\n" +
		readJS(t, "handlers.js") + "\n" +
		readJS(t, "bootstrap.js")

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
