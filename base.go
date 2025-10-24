package liveflux

import (
	"log"

	"github.com/dracory/hb"
	"github.com/samber/lo"
)

// Base provides minimal implementation for alias and ID handling
// and is exported so downstream packages can embed `liveflux.Base`
// for shared behavior.
type Base struct {
	// alias is the TYPE identifier (registry key). Set once by the framework.
	alias string

	// id is the INSTANCE identifier (unique per mount). Set during mount.
	id string

	// redirect, if set by a component during Handle, signals the handler to instruct
	// the client to navigate. Consumed via TakeRedirect().
	redirect string
	// redirectAfterSeconds optionally delays the redirect on the client.
	redirectAfterSeconds int

	// eventDispatcher manages event dispatching and listening for this component.
	eventDispatcher *EventDispatcher
}

// GetAlias returns the component's alias.
func (b *Base) GetAlias() string {
	return b.alias
}

// SetAlias sets the component's alias only once (no-op if already set or empty input).
func (b *Base) SetAlias(alias string) {
	if b.alias == "" && alias != "" {
		b.alias = alias
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
		Attr("data-flux-root", "1").
		Child(hb.Input().Type("hidden").Name(FormComponent).Value(b.GetAlias())).
		Child(hb.Input().Type("hidden").Name(FormID).Value(b.GetID()))

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

// DispatchToAlias queues an event to be sent to a specific component alias.
// Usage: component.DispatchToAlias("users.list", "post-created", map[string]any{"id": 1, "title": "My Post"})
func (b *Base) DispatchToAlias(componentAlias string, eventName string, data ...map[string]any) {
	b.GetEventDispatcher().DispatchToAlias(componentAlias, eventName, data...)
}

// DispatchToAliasAndID queues an event to be sent to a specific component alias and ID.
// Usage: component.DispatchToAliasAndID("users.list", someID, "post-updated", map[string]any{"id": 1})
func (b *Base) DispatchToAliasAndID(componentAlias string, componentID string, eventName string, data ...map[string]any) {
	log.Println("Dispatching to alias and ID:", componentAlias, componentID, eventName, data)
	b.GetEventDispatcher().DispatchToAliasAndID(componentAlias, componentID, eventName, data...)
}

// DispatchSelf queues an event to be sent only to the current component.
// Usage: component.DispatchSelf("post-created", map[string]any{"id": 1, "title": "My Post"})
func (b *Base) DispatchSelf(eventName string, data ...map[string]any) {
	dataUpdated := lo.FirstOr(data, map[string]any{})
	dataUpdated["__self"] = true
	b.GetEventDispatcher().DispatchToAliasAndID(b.GetAlias(), b.GetID(), eventName, dataUpdated)
}
