package liveflux

import (
	"context"
	"encoding/json"
	"reflect"
)

// Event represents a dispatched event with a name and optional data payload.
type Event struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// EventDispatcher manages event dispatching and listening within components.
type EventDispatcher struct {
	// events holds the list of events to be dispatched after the current request
	events []Event
	// listeners maps event names to handler functions
	listeners map[string][]EventListener
}

// EventListener is a function that handles an event.
type EventListener func(ctx context.Context, event Event) error

// NewEventDispatcher creates a new event dispatcher.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		events:    []Event{},
		listeners: make(map[string][]EventListener),
	}
}

// Dispatch queues an event to be sent to the client and other components.
// Usage: dispatcher.Dispatch("post-created", map[string]interface{}{"title": "My Post"})
func (ed *EventDispatcher) Dispatch(name string, data ...map[string]interface{}) {
	event := Event{Name: name}
	if len(data) > 0 && data[0] != nil {
		event.Data = data[0]
	}
	ed.events = append(ed.events, event)
}

// DispatchToAlias queues an event to be sent to a specific component alias.
// This is handled by the client-side runtime which filters events by alias.
func (ed *EventDispatcher) DispatchToAlias(componentAlias string, eventName string, data ...map[string]any) {
	event := Event{Name: eventName}
	if len(data) > 0 && data[0] != nil {
		event.Data = data[0]
	} else {
		event.Data = make(map[string]any)
	}
	// Add target alias metadata
	event.Data["__target"] = componentAlias
	ed.events = append(ed.events, event)
}

// DispatchToAliasAndID queues an event to be sent to a specific component alias and ID.
// The client-side runtime filters events by both alias and component ID.
func (ed *EventDispatcher) DispatchToAliasAndID(componentAlias string, componentID string, eventName string, data ...map[string]any) {
	event := Event{Name: eventName}
	if len(data) > 0 && data[0] != nil {
		event.Data = data[0]
	} else {
		event.Data = make(map[string]any)
	}
	// Add target alias and ID metadata
	event.Data["__target"] = componentAlias
	event.Data["__target_id"] = componentID
	ed.events = append(ed.events, event)
}

// On registers an event listener for the given event name.
// Multiple listeners can be registered for the same event.
func (ed *EventDispatcher) On(name string, handler EventListener) {
	ed.listeners[name] = append(ed.listeners[name], handler)
}

// TriggerLocal triggers event listeners registered on this dispatcher (server-side only).
// This allows components to respond to events during the same request cycle.
func (ed *EventDispatcher) TriggerLocal(ctx context.Context, event Event) error {
	handlers, ok := ed.listeners[event.Name]
	if !ok {
		return nil
	}
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// TakeEvents returns all queued events and clears the queue.
func (ed *EventDispatcher) TakeEvents() []Event {
	events := ed.events
	ed.events = []Event{}
	return events
}

// HasEvents returns true if there are queued events.
func (ed *EventDispatcher) HasEvents() bool {
	return len(ed.events) > 0
}

// EventsJSON returns the queued events as a JSON string.
func (ed *EventDispatcher) EventsJSON() string {
	events := ed.TakeEvents()
	if len(events) == 0 {
		return "[]"
	}
	data, err := json.Marshal(events)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// EventAware is an interface for components that support event dispatching.
type EventAware interface {
	GetEventDispatcher() *EventDispatcher
}

// OnAttribute represents the #[On] attribute pattern from Livewire.
// Components can use struct tags or method naming conventions to declare event listeners.
type OnAttribute struct {
	EventName string
	Method    string
}

// ParseOnAttributes scans a component for methods with "On" prefix and registers them.
// Convention: OnEventName(ctx context.Context, event Event) error
// Example: OnPostCreated handles "post-created" event
func ParseOnAttributes(component ComponentInterface) []OnAttribute {
	attrs := []OnAttribute{}
	t := reflect.TypeOf(component)

	// We need to scan the pointer type's methods, not the struct type
	// Methods are defined on the pointer receiver

	// Scan for methods with "On" prefix
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if len(method.Name) > 2 && method.Name[:2] == "On" {
			// Convert OnPostCreated -> post-created
			eventName := toKebabCase(method.Name[2:])
			attrs = append(attrs, OnAttribute{
				EventName: eventName,
				Method:    method.Name,
			})
		}
	}

	return attrs
}

// toKebabCase converts PascalCase to kebab-case
// Example: PostCreated -> post-created
func toKebabCase(s string) string {
	if s == "" {
		return ""
	}
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32) // Convert to lowercase
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// RegisterEventListeners automatically registers event listeners based on method naming.
func RegisterEventListeners(component ComponentInterface, dispatcher *EventDispatcher) {
	attrs := ParseOnAttributes(component)
	v := reflect.ValueOf(component)

	for _, attr := range attrs {
		methodName := attr.Method
		eventName := attr.EventName

		// Get the method
		method := v.MethodByName(methodName)
		if !method.IsValid() {
			continue
		}

		// Create a closure that calls the method
		dispatcher.On(eventName, func(ctx context.Context, event Event) error {
			// Call the method with ctx and event
			results := method.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(event),
			})

			// Check if method returned an error
			if len(results) > 0 && !results[0].IsNil() {
				if err, ok := results[0].Interface().(error); ok {
					return err
				}
			}
			return nil
		})
	}
}
