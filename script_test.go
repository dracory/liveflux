package liveflux

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gouniverse/hb"
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
		readJS(t, "mount.js") + "\n" +
		readJS(t, "handlers.js") + "\n" +
		readJS(t, "bootstrap.js")

	got := JS()
	if got != expected {
		// Keep failure output compact but helpful
		t.Fatalf("JS() content mismatch with expected concatenation of js/*.js files")
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
