package liveflux

import (
	"context"
	"net/url"
	"testing"

	"github.com/dracory/hb"
)

func TestEventDispatcher_Dispatch(t *testing.T) {
	ed := NewEventDispatcher()

	ed.Dispatch("test-event", map[string]interface{}{"key": "value"})

	if !ed.HasEvents() {
		t.Fatal("expected events to be queued")
	}

	events := ed.TakeEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Name != "test-event" {
		t.Errorf("expected event name 'test-event', got '%s'", events[0].Name)
	}

	if events[0].Data["key"] != "value" {
		t.Errorf("expected data key='value', got '%v'", events[0].Data["key"])
	}

	// After taking, should be empty
	if ed.HasEvents() {
		t.Error("expected no events after taking")
	}
}

func TestEventDispatcher_EventsJSONDrainsQueue(t *testing.T) {
	ed := NewEventDispatcher()

	ed.Dispatch("sample", map[string]interface{}{"key": "value"})

	first := ed.EventsJSON()
	if first == "[]" {
		t.Fatalf("expected first EventsJSON call to contain data, got %s", first)
	}

	second := ed.EventsJSON()
	if second != "[]" {
		t.Fatalf("expected second EventsJSON call to be empty, got %s", second)
	}
}

func TestEventDispatcher_On(t *testing.T) {
	ed := NewEventDispatcher()
	called := false

	ed.On("test-event", func(ctx context.Context, event Event) error {
		called = true
		return nil
	})

	event := Event{Name: "test-event"}
	err := ed.TriggerLocal(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected listener to be called")
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PostCreated", "post-created"},
		{"UserUpdated", "user-updated"},
		{"Simple", "simple"},
		{"", ""},
		{"ABC", "a-b-c"},
	}

	for _, tt := range tests {
		result := toKebabCase(tt.input)
		if result != tt.expected {
			t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

type TestComponent struct {
	Base
	EventDispatcher     *EventDispatcher
	OnPostCreatedCalled bool
	OnUserUpdatedCalled bool
}

func (tc *TestComponent) Mount(ctx context.Context, params map[string]string) error {
	return nil
}

func (tc *TestComponent) Handle(ctx context.Context, action string, data url.Values) error {
	return nil
}

func (tc *TestComponent) Render(ctx context.Context) hb.TagInterface {
	return hb.Div()
}

func (tc *TestComponent) OnPostCreated(ctx context.Context, event Event) error {
	tc.OnPostCreatedCalled = true
	return nil
}

func (tc *TestComponent) OnUserUpdated(ctx context.Context, event Event) error {
	tc.OnUserUpdatedCalled = true
	return nil
}

func TestParseOnAttributes(t *testing.T) {
	comp := &TestComponent{}
	attrs := ParseOnAttributes(comp)

	if len(attrs) < 2 {
		t.Fatalf("expected at least 2 attributes, got %d", len(attrs))
	}

	// Check that we found the OnPostCreated and OnUserUpdated methods
	foundPost := false
	foundUser := false
	for _, attr := range attrs {
		if attr.EventName == "post-created" && attr.Method == "OnPostCreated" {
			foundPost = true
		}
		if attr.EventName == "user-updated" && attr.Method == "OnUserUpdated" {
			foundUser = true
		}
	}

	if !foundPost {
		t.Error("expected to find OnPostCreated -> post-created")
	}
	if !foundUser {
		t.Error("expected to find OnUserUpdated -> user-updated")
	}
}

func TestRegisterEventListeners(t *testing.T) {
	comp := &TestComponent{
		EventDispatcher: NewEventDispatcher(),
	}

	RegisterEventListeners(comp, comp.EventDispatcher)

	// Trigger the post-created event
	event := Event{Name: "post-created"}
	err := comp.EventDispatcher.TriggerLocal(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !comp.OnPostCreatedCalled {
		t.Error("expected OnPostCreated to be called")
	}

	// Trigger the user-updated event
	event2 := Event{Name: "user-updated"}
	err = comp.EventDispatcher.TriggerLocal(context.Background(), event2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !comp.OnUserUpdatedCalled {
		t.Error("expected OnUserUpdated to be called")
	}
}

func TestEventDispatcher_DispatchToKind(t *testing.T) {
	ed := NewEventDispatcher()

	ed.DispatchToKind("dashboard", "refresh", map[string]any{"count": 5})

	events := ed.TakeEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Data["__target"] != "dashboard" {
		t.Errorf("expected __target='dashboard', got '%v'", events[0].Data["__target"])
	}

	if events[0].Data["count"] != 5 {
		t.Errorf("expected count=5, got '%v'", events[0].Data["count"])
	}
}

func TestEventDispatcher_DispatchToKindAndID(t *testing.T) {
	ed := NewEventDispatcher()

	ed.DispatchToKindAndID("dashboard", "component-123", "refresh", map[string]any{"count": 2})

	events := ed.TakeEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Data["__target"] != "dashboard" {
		t.Errorf("expected __target='dashboard', got '%v'", events[0].Data["__target"])
	}

	if events[0].Data["__target_id"] != "component-123" {
		t.Errorf("expected __target_id='component-123', got '%v'", events[0].Data["__target_id"])
	}

	if events[0].Data["count"] != 2 {
		t.Errorf("expected count=2, got '%v'", events[0].Data["count"])
	}
}
