package liveflux

import (
	"testing"
)

func TestBase_SetKind_Once(t *testing.T) {
	var b Base
	if b.GetKind() != "" {
		t.Fatalf("new Base should have empty kind")
	}
	b.SetKind("first")
	if got := b.GetKind(); got != "first" {
		t.Fatalf("expected kind 'first', got %q", got)
	}
	// Setting empty should not change
	b.SetKind("")
	if got := b.GetKind(); got != "first" {
		t.Fatalf("kind changed on empty SetKind: %q", got)
	}
	// Setting a new value after set-once should be ignored
	b.SetKind("second")
	if got := b.GetKind(); got != "first" {
		t.Fatalf("kind should be set-once, got %q", got)
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
	b.SetKind("foo.bar")
	b.SetID("123")

	// Dispatch -> no targeting metadata
	b.Dispatch("alpha", map[string]any{"value": 1})

	// DispatchToKind -> kind metadata
	b.DispatchToKind("other.kind", "beta", map[string]any{"value": 2})

	// DispatchToKindAndID -> kind & id metadata
	b.DispatchToKindAndID("other.kind", "other-id", "gamma", map[string]any{"value": 3})

	// DispatchSelf -> kind & id metadata for current component
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

	if target := events[1].Data["__target"]; target != "other.kind" {
		t.Fatalf("event 1 __target got %v want %v", target, "other.kind")
	}

	if events[2].Name != "gamma" {
		t.Fatalf("event 2 name got %q want %q", events[2].Name, "gamma")
	}

	if target := events[2].Data["__target"]; target != "other.kind" {
		t.Fatalf("event 2 __target got %v want %v", target, "other.kind")
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
