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

func TestBase_DispatchHelpers(t *testing.T) {
	var b Base
	b.SetAlias("foo.bar")
	b.SetID("123")

	// Dispatch -> no targeting metadata
	b.Dispatch("alpha", map[string]any{"value": 1})

	// DispatchToAlias -> alias metadata
	b.DispatchToAlias("other.alias", "beta", map[string]any{"value": 2})

	// DispatchToAliasAndID -> alias & id metadata
	b.DispatchToAliasAndID("other.alias", "other-id", "gamma", map[string]any{"value": 3})

	// DispatchSelf -> alias & id metadata for current component
	b.DispatchSelf("delta", map[string]any{"value": 4})

	events := b.GetEventDispatcher().TakeEvents()
	if len(events) != 4 {
		t.Fatalf("expected 4 queued events, got %d", len(events))
	}

	if events[0].Name != "alpha" {
		t.Fatalf("event 0 name got %q want %q", events[0].Name, "alpha")
	}

	if events[0].Data["value"] != 1 {
		t.Fatalf("event 0 value got %v want %v", events[0].Data["value"], 1)
	}

	if events[1].Name != "beta" {
		t.Fatalf("event 1 name got %q want %q", events[1].Name, "beta")
	}

	if target := events[1].Data["__target"]; target != "other.alias" {
		t.Fatalf("event 1 __target got %v want %v", target, "other.alias")
	}

	if events[2].Name != "gamma" {
		t.Fatalf("event 2 name got %q want %q", events[2].Name, "gamma")
	}

	if target := events[2].Data["__target"]; target != "other.alias" {
		t.Fatalf("event 2 __target got %v want %v", target, "other.alias")
	}

	if targetID := events[2].Data["__target_id"]; targetID != "other-id" {
		t.Fatalf("event 2 __target_id got %v want %v", targetID, "other-id")
	}

	if events[3].Name != "delta" {
		t.Fatalf("event 3 name got %q want %q", events[3].Name, "delta")
	}

	if target := events[3].Data["__target"]; target != "foo.bar" {
		t.Fatalf("event 3 __target got %v want %v", target, "foo.bar")
	}

	if targetID := events[3].Data["__target_id"]; targetID != "123" {
		t.Fatalf("event 3 __target_id got %v want %v", targetID, "123")
	}
}
