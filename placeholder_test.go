package liveflux

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/gouniverse/hb"
)

type phComp struct{ Base }

func (c *phComp) GetAlias() string { return "" }
func (c *phComp) Mount(context.Context, map[string]string) error { return nil }
func (c *phComp) Handle(context.Context, string, url.Values) error { return nil }
func (c *phComp) Render(context.Context) hb.TagInterface { return hb.Div() }

func TestPlaceholderByAlias_BuildsDivWithParams(t *testing.T) {
	alias := "users.list"
	tag := PlaceholderByAlias(alias, map[string]string{"foo": "bar", "": "ignore"})
	h := tag.ToHTML()
	// Basic attributes
	if !strings.Contains(h, "data-lw-mount=\"1\"") {
		t.Fatalf("expected data-lw-mount=1, got: %s", h)
	}
	if !strings.Contains(h, "data-lw-component=\""+alias+"\"") {
		t.Fatalf("expected data-lw-component with alias, got: %s", h)
	}
	if !strings.Contains(h, "data-lw-param-foo=\"bar\"") {
		t.Fatalf("expected data param attr, got: %s", h)
	}
	if !strings.Contains(h, "Loading "+alias+"...") {
		t.Fatalf("expected loading label, got: %s", h)
	}
}

func TestPlaceholder_NilComponent(t *testing.T) {
	h := Placeholder(nil).ToHTML()
	if h != "component missing" {
		t.Fatalf("expected 'component missing', got: %q", h)
	}
}

func TestPlaceholder_ComponentNoAlias(t *testing.T) {
	c := &phComp{}
	h := Placeholder(c).ToHTML()
	if h != "component has no alias" {
		t.Fatalf("expected 'component has no alias', got: %q", h)
	}
}

func TestPlaceholder_ComponentWithAlias_Registered(t *testing.T) {
	// Register this type with a unique alias and ensure Placeholder uses it
	unique := "test." + NewID()
	RegisterByAlias(unique, func() ComponentInterface { return &phComp{} })

	c := &phComp{}
	h := Placeholder(c, map[string]string{"x": "1"}).ToHTML()
	if !strings.Contains(h, "data-lw-component=\""+unique+"\"") {
		t.Fatalf("expected placeholder to use registered alias %q, got: %s", unique, h)
	}
	if !strings.Contains(h, "data-lw-param-x=\"1\"") {
		t.Fatalf("expected param to be rendered, got: %s", h)
	}
}
