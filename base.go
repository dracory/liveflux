package liveflux

import (
	"fmt"

	"github.com/dracory/hb"
	"github.com/samber/lo"
)

// Base provides minimal implementation for kind and ID handling
// and is exported so downstream packages can embed `liveflux.Base`
// for shared behavior.
type Base struct {
	// kind is the TYPE identifier (registry key). Set once by the framework.
	kind string

	// id is the INSTANCE identifier (unique per mount). Set during mount.
	id string

	// redirect, if set by a component during Handle, signals the handler to instruct
	// the client to navigate. Consumed via TakeRedirect().
	redirect string
	// redirectAfterSeconds optionally delays the redirect on the client.
	redirectAfterSeconds int

	// eventDispatcher manages event dispatching and listening for this component.
	eventDispatcher *EventDispatcher

	// dirtyTargets tracks which DOM targets need to be updated.
	// Components can use MarkTargetDirty to signal that specific selectors changed.
	dirtyTargets map[string]bool
}

// GetKind returns the component's kind.
func (b *Base) GetKind() string {
	return b.kind
}

// SetKind sets the component's kind only once (no-op if already set or empty input).
func (b *Base) SetKind(kind string) {
	if b.kind == "" && kind != "" {
		b.kind = kind
	}
}

// GetID returns the component's instance ID.
func (b *Base) GetID() string {
	return b.id
}

// SetID sets the component's instance ID.
func (b *Base) SetID(id string) {
	b.id = id
}

// Redirect requests a client-side redirect with an optional delay in seconds.
// If delaySeconds is not provided, 0 is used (immediate).
func (b *Base) Redirect(url string, delaySeconds ...int) {
	b.SetRedirect(url)
	d := lo.FirstOr(delaySeconds, 0)
	b.SetRedirectDelaySeconds(d)
}

// SetRedirect requests an immediate client-side redirect to the given URL on next response.
func (b *Base) SetRedirect(url string) { b.redirect = url; b.redirectAfterSeconds = 0 }

// SetRedirectDelay sets an optional delay (seconds) for a previously-set redirect.
// Negative values are treated as zero.
func (b *Base) SetRedirectDelaySeconds(seconds int) {
	if seconds < 0 {
		seconds = 0
	}
	b.redirectAfterSeconds = seconds
}

// TakeRedirect returns the pending redirect URL (if any) and clears it.
func (b *Base) TakeRedirect() string {
	url := b.redirect
	b.redirect = ""
	return url
}

// TakeRedirectDelaySeconds returns the pending redirect delay seconds (if any) and resets it.
func (b *Base) TakeRedirectDelaySeconds() int {
	d := b.redirectAfterSeconds
	b.redirectAfterSeconds = 0
	return d
}

// Root returns the standard Liveflux component root element with required
// attributes and hidden fields, and appends the provided content as a child.
// This avoids repeating boilerplate in every component's Render method.
func (b *Base) Root(content hb.TagInterface) hb.TagInterface {
	root := hb.Div().
		// data-flux-root is checked by the client to determine if this is a Liveflux component.
		Attr(DataFluxRoot, "1").
		// data-flux-component-kind is the component kind (KIND identifier).
		Attr(DataFluxComponentKind, b.GetKind()).
		// data-flux-component-id is the component instance ID (INSTANCE identifier).
		Attr(DataFluxComponentID, b.GetID())

	if content != nil {
		root = root.Child(content)
	}
	return root
}

// GetEventDispatcher returns the component's event dispatcher, creating it if needed.
func (b *Base) GetEventDispatcher() *EventDispatcher {
	if b.eventDispatcher == nil {
		b.eventDispatcher = NewEventDispatcher()
	}
	return b.eventDispatcher
}

// Dispatch queues an event to be sent to the client and other components.
// Usage: component.Dispatch("post-created", map[string]any{"id": 1, "title": "My Post"})
func (b *Base) Dispatch(eventName string, data ...map[string]any) {
	b.GetEventDispatcher().Dispatch(eventName, data...)
}

// DispatchToKind queues an event to be sent to a specific component kind.
// Usage: component.DispatchToKind("users.list", "post-created", map[string]any{"id": 1, "title": "My Post"})
func (b *Base) DispatchToKind(componentKind string, eventName string, data ...map[string]any) {
	b.GetEventDispatcher().DispatchToKind(componentKind, eventName, data...)
}

// DispatchToKindAndID queues an event to be sent to a specific component kind and ID.
// Usage: component.DispatchToKindAndID("users.list", someID, "post-updated", map[string]any{"id": 1})
func (b *Base) DispatchToKindAndID(componentKind string, componentID string, eventName string, data ...map[string]any) {
	fmt.Printf("Dispatching to kind and ID: %s %s %s %v\n", componentKind, componentID, eventName, data)
	b.GetEventDispatcher().DispatchToKindAndID(componentKind, componentID, eventName, data...)
}

// DispatchSelf queues an event to be sent only to the current component.
// Usage: component.DispatchSelf("post-created", map[string]any{"id": 1, "title": "My Post"})
func (b *Base) DispatchSelf(eventName string, data ...map[string]any) {
	dataUpdated := lo.FirstOr(data, map[string]any{})
	dataUpdated["__self"] = true
	b.GetEventDispatcher().DispatchToKindAndID(b.GetKind(), b.GetID(), eventName, dataUpdated)
}

// MarkTargetDirty marks a specific DOM target as needing an update.
// This is used in conjunction with TargetRenderer to track which fragments changed.
// Usage: component.MarkTargetDirty("#cart-total")
func (b *Base) MarkTargetDirty(selector string) {
	if b.dirtyTargets == nil {
		b.dirtyTargets = make(map[string]bool)
	}
	b.dirtyTargets[selector] = true
}

// IsDirty checks if a specific target has been marked dirty.
func (b *Base) IsDirty(selector string) bool {
	if b.dirtyTargets == nil {
		return false
	}
	return b.dirtyTargets[selector]
}

// ClearDirtyTargets clears all dirty target markers.
func (b *Base) ClearDirtyTargets() {
	b.dirtyTargets = make(map[string]bool)
}

// GetDirtyTargets returns a copy of the dirty targets map.
func (b *Base) GetDirtyTargets() map[string]bool {
	if b.dirtyTargets == nil {
		return make(map[string]bool)
	}
	result := make(map[string]bool, len(b.dirtyTargets))
	for k, v := range b.dirtyTargets {
		result[k] = v
	}
	return result
}
