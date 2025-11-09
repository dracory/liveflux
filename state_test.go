package liveflux

import (
	"context"
	"net/url"
	"testing"

	"github.com/dracory/hb"
)

// minimal component for store tests
type storeComp struct{ Base }

func (c *storeComp) GetKind() string                                { return "test.store-comp" }
func (c *storeComp) Mount(context.Context, map[string]string) error { return nil }
func (c *storeComp) Handle(context.Context, string, url.Values) error {
	return nil
}
func (c *storeComp) Render(context.Context) hb.TagInterface { return hb.Div() }

func TestMemoryStore_GetEmpty(t *testing.T) {
	s := NewMemoryStore()
	if got, ok := s.Get("missing"); ok || got != nil {
		t.Fatalf("expected empty store to return (nil,false)")
	}
}

func TestMemoryStore_SetIgnoresEmptyID(t *testing.T) {
	s := NewMemoryStore()
	c := &storeComp{} // ID empty
	s.Set(c)
	if got, ok := s.Get(""); ok || got != nil {
		t.Fatalf("expected Set to ignore components with empty ID")
	}
}

func TestMemoryStore_SetGetAndDelete(t *testing.T) {
	s := NewMemoryStore()
	c := &storeComp{}
	c.SetID("abc123")
	s.Set(c)

	got, ok := s.Get("abc123")
	if !ok || got == nil {
		t.Fatalf("expected to retrieve stored component")
	}
	if got.GetID() != "abc123" {
		t.Fatalf("unexpected component retrieved: %v", got.GetID())
	}

	s.Delete("abc123")
	if got2, ok2 := s.Get("abc123"); ok2 || got2 != nil {
		t.Fatalf("expected component to be deleted")
	}
}

func TestStoreDefault_IsInitialized(t *testing.T) {
	if StoreDefault == nil {
		t.Fatalf("StoreDefault should be initialized")
	}
}
