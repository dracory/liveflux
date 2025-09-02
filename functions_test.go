package liveflux

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/hb"
)

// Types for DefaultAliasFromType tests
type counterType struct{ Base }

func (c *counterType) GetAlias() string                                                 { return "" }
func (c *counterType) Mount(ctx context.Context, params map[string]string) error        { return nil }
func (c *counterType) Handle(ctx context.Context, action string, data url.Values) error { return nil }
func (c *counterType) Render(ctx context.Context) hb.TagInterface                       { return hb.Div() }

type Liveflux struct{ Base }                                                         // matches package name case-insensitively
func (c *Liveflux) GetAlias() string                                                 { return "" }
func (c *Liveflux) Mount(ctx context.Context, params map[string]string) error        { return nil }
func (c *Liveflux) Handle(ctx context.Context, action string, data url.Values) error { return nil }
func (c *Liveflux) Render(ctx context.Context) hb.TagInterface                       { return hb.Div() }

func TestNewID_LengthCharsetUniqueness(t *testing.T) {
	allowed := "123456789bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ"
	seen := map[string]struct{}{}
	for i := 0; i < 200; i++ {
		id := NewID()
		if len(id) != 12 {
			t.Fatalf("expected len 12, got %d (%q)", len(id), id)
		}
		for _, r := range id {
			if !strings.ContainsRune(allowed, r) {
				t.Fatalf("id contains disallowed rune %q in %q", r, id)
			}
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestDefaultAliasFromType(t *testing.T) {
	if got := DefaultAliasFromType(nil); got != "" {
		t.Fatalf("expected empty for nil, got %q", got)
	}

	// Type name different from package => pkg.type-kebab, lowercased
	c := &counterType{}
	if got := DefaultAliasFromType(c); got != "liveflux.counter-type" {
		t.Fatalf("unexpected alias for counterType: %q", got)
	}

	// Type name equal to package (case-insensitive) => just package
	l := &Liveflux{}
	if got := DefaultAliasFromType(l); got != "liveflux" {
		t.Fatalf("unexpected alias for Liveflux: %q", got)
	}

	// Pointer vs value shouldn't matter
	var cp ComponentInterface = c
	if got := DefaultAliasFromType(cp); got != "liveflux.counter-type" {
		t.Fatalf("unexpected alias for pointer receiver: %q", got)
	}
}

func TestToKebab(t *testing.T) {
	cases := map[string]string{
		"SimpleCase":      "simple-case",
		"HTTPServer":      "http-server",
		"userID":          "user-id",
		"User_List":       "user-list",
		"User List":       "user-list",
		"already-kebab":   "already-kebab",
		"snake_case_word": "snake-case-word",
	}
	for in, want := range cases {
		if got := toKebab(in); got != want {
			t.Fatalf("toKebab(%q)=%q want %q", in, got, want)
		}
	}
}
