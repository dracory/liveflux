package liveflux

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/hb"
)

type phComp struct{ Base }

func (c *phComp) GetKind() string                                  { return "phComp" }
func (c *phComp) Mount(context.Context, map[string]string) error   { return nil }
func (c *phComp) Handle(context.Context, string, url.Values) error { return nil }
func (c *phComp) Render(context.Context) hb.TagInterface           { return hb.Div() }

func TestPlaceholderByKind_BuildsDivWithParams(t *testing.T) {
	kind := "users.list"
	tag := PlaceholderByKind(kind, map[string]string{"foo": "bar", "": "ignore"})
	h := tag.ToHTML()
	// Basic attributes
	if !strings.Contains(h, DataFluxMount+"=\"1\"") {
		t.Fatalf("expected data-flux-mount=1, got: %s", h)
	}
	if !strings.Contains(h, DataFluxComponentKind+"=\""+kind+"\"") {
		t.Fatalf("expected data-flux-component-kind with kind, got: %s", h)
	}
	if !strings.Contains(h, DataFluxParam+"-foo=\"bar\"") {
		t.Fatalf("expected data param attr, got: %s", h)
	}
	if !strings.Contains(h, "Loading "+kind+"...") {
		t.Fatalf("expected loading label, got: %s", h)
	}
}

func TestPlaceholder_NilComponent(t *testing.T) {
	h := Placeholder(nil).ToHTML()
	if h != "component missing" {
		t.Fatalf("expected 'component missing', got: %q", h)
	}
}

func TestPlaceholder_ComponentNoKind(t *testing.T) {
	c := &phComp{}
	h := Placeholder(c).ToHTML()
	if h != "component has no kind" {
		t.Fatalf("expected 'component has no kind', got: %q", h)
	}
}

func TestPlaceholder_ComponentWithKind_Registered(t *testing.T) {
	// Register this type with a unique kind and ensure Placeholder uses it
	unique := "test." + NewID()
	RegisterByKind(unique, &phComp{}) // unique kind

	c := &phComp{}
	h := Placeholder(c, map[string]string{"x": "1"}).ToHTML()
	if !strings.Contains(h, "data-flux-component-kind=\""+unique+"\"") {
		t.Fatalf("expected placeholder to use registered kind %q, got: %s", unique, h)
	}
	if !strings.Contains(h, "data-flux-param-x=\"1\"") {
		t.Fatalf("expected param to be rendered, got: %s", h)
	}
}
