package liveflux

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/gouniverse/hb"
)

type testSSRComp struct {
	Base
	mounted map[string]string
}

func (c *testSSRComp) GetAlias() string { return "test.ssr-comp" }
func (c *testSSRComp) Mount(_ context.Context, params map[string]string) error {
	c.mounted = params
	return nil
}
func (c *testSSRComp) Handle(_ context.Context, _ string, _ url.Values) error { return nil }
func (c *testSSRComp) Render(_ context.Context) hb.TagInterface {
	return hb.Div().Attr("data-id", c.GetID()).Text("ok")
}

type errSSRComp struct{ Base }

func (c *errSSRComp) GetAlias() string                                       { return "test.err-comp" }
func (c *errSSRComp) Mount(_ context.Context, _ map[string]string) error     { return errors.New("boom") }
func (c *errSSRComp) Handle(_ context.Context, _ string, _ url.Values) error { return nil }
func (c *errSSRComp) Render(_ context.Context) hb.TagInterface {
	return hb.Div().Text("should not be used")
}

func TestSSR_SetsID_CallsMount_Persists(t *testing.T) {
	oldStore := StoreDefault
	defer func() { StoreDefault = oldStore }()
	StoreDefault = NewMemoryStore()

	c := &testSSRComp{}
	html := SSRHTML(c, map[string]string{"a": "1", "b": "2"})
	if c.GetID() == "" {
		t.Fatalf("expected SSR to set an ID on the component")
	}
	if c.mounted == nil || c.mounted["a"] != "1" || c.mounted["b"] != "2" {
		t.Fatalf("expected Mount to receive params, got: %#v", c.mounted)
	}

	// SSRHTML should be equivalent to SSR(...).ToHTML()
	if html2 := SSR(c).ToHTML(); html2 != html {
		t.Fatalf("SSRHTML and SSR(...).ToHTML() mismatch")
	}

	// Persisted in store
	if got, ok := StoreDefault.Get(c.GetID()); !ok || got == nil {
		t.Fatalf("expected component to be stored after SSR")
	}
}

func TestSSR_MountError(t *testing.T) {
	oldStore := StoreDefault
	defer func() { StoreDefault = oldStore }()
	StoreDefault = NewMemoryStore()

	c := &errSSRComp{}
	tag := SSR(c)
	h := tag.ToHTML()
	if c.GetID() == "" {
		t.Fatalf("expected ID to be set before Mount even on error")
	}
	if got, ok := StoreDefault.Get(c.GetID()); ok || got != nil {
		t.Fatalf("did not expect component to be stored on mount error")
	}
	if wantSub := "mount error:"; !strings.Contains(h, wantSub) {
		t.Fatalf("expected SSR to return mount error HTML containing %q, got: %s", wantSub, h)
	}
}
