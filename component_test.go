package liveflux

import (
	"testing"
)

// minimal component for Base behavior tests
type baseComp struct{ Base }

func (c *baseComp) GetAlias() string { return c.Base.GetAlias() }

func TestBase_SetAlias_Once(t *testing.T) {
	var b Base
	if b.GetAlias() != "" {
		t.Fatalf("new Base should have empty alias")
	}
	b.SetAlias("first")
	if got := b.GetAlias(); got != "first" {
		t.Fatalf("expected alias 'first', got %q", got)
	}
	// Setting empty should not change
	b.SetAlias("")
	if got := b.GetAlias(); got != "first" {
		t.Fatalf("alias changed on empty SetAlias: %q", got)
	}
	// Setting a new value after set-once should be ignored
	b.SetAlias("second")
	if got := b.GetAlias(); got != "first" {
		t.Fatalf("alias should be set-once, got %q", got)
	}
}

func TestBase_ID_GetSet(t *testing.T) {
	var b Base
	if b.GetID() != "" {
		t.Fatalf("new Base should have empty id")
	}
	b.SetID("abc")
	if got := b.GetID(); got != "abc" {
		t.Fatalf("expected id 'abc', got %q", got)
	}
}

func TestBase_Redirect_API(t *testing.T) {
	var b Base
	// SetRedirect immediate
	b.SetRedirect("/home")
	if url := b.TakeRedirect(); url != "/home" {
		t.Fatalf("TakeRedirect got %q want %q", url, "/home")
	}
	// After taking, should be cleared
	if url := b.TakeRedirect(); url != "" {
		t.Fatalf("TakeRedirect should clear redirect, got %q", url)
	}

	// Redirect with delay
	b.Redirect("/next", 3)
	if url := b.TakeRedirect(); url != "/next" {
		t.Fatalf("TakeRedirect got %q want %q", url, "/next")
	}
	d := b.TakeRedirectDelaySeconds()
	if d != 3 {
		t.Fatalf("delay got %d want %d", d, 3)
	}
	// Delay resets after taking
	if d2 := b.TakeRedirectDelaySeconds(); d2 != 0 {
		t.Fatalf("delay should reset to 0 after take, got %d", d2)
	}

	// Negative delay normalized to 0
	b.SetRedirect("/n")
	b.SetRedirectDelaySeconds(-5)
	if d3 := b.TakeRedirectDelaySeconds(); d3 != 0 {
		t.Fatalf("negative delay should normalize to 0, got %d", d3)
	}
}
