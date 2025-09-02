package liveflux

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/samber/lo"
)

// Component defines the contract for a server-driven UI component.
//
// Lifecycle:
// - New(alias) via registry (framework sets the component's Alias via SetAlias(alias))
// - SetID(...) during first mount (framework assigns a per-instance ID)
// - Mount(ctx, params) on first initialization
// - Handle(ctx, action, form) on user actions
// - Render(ctx) returns the current HTML (hb.Tag)
//
// Name is assigned by the framework when the instance is created from the registry.
// ID and SetID are used by the framework to track component instances and are
// assigned on mount.
// Mount should initialize default state. Handle should mutate state based on actions.
// Render must be deterministic based on current state.
//
// Example usage:
//
//	type Counter struct { liveflux.Base; Count int }
//	func (c *Counter) Mount(ctx context.Context, params map[string]string) error { c.Count = 0; return nil }
//	func (c *Counter) Handle(ctx context.Context, action string, data url.Values) error { if action=="inc" { c.Count++ }; return nil }
//	func (c *Counter) Render(ctx context.Context) hb.TagInterface { return hb.Div().Textf("%d", c.Count) }
//	liveflux.RegisterByAlias("counter", func() liveflux.ComponentInterface { return &Counter{} })
//
// See handler.go for the HTTP entry point.
type ComponentInterface interface {
	// GetAlias returns the stable routing key (TYPE identifier) used by the client
	// and registry to select which component to construct/route to.
	// Component authors MUST implement this.
	GetAlias() string

	// SetAlias sets the component's alias (TYPE identifier). The framework sets this
	// once during construction in the registry. Implementations should treat alias
	// as immutable; Base enforces set-once semantics.
	SetAlias(alias string)

	// GetID returns the instance ID (per-mount INSTANCE identifier).
	GetID() string

	// SetID sets the instance ID (per-mount INSTANCE identifier). This is called by
	// the framework during mount (see handler.go).
	SetID(id string)

	Mount(ctx context.Context, params map[string]string) error
	Handle(ctx context.Context, action string, data url.Values) error
	Render(ctx context.Context) hb.TagInterface
}

// Backward-compatible alias to avoid breaking existing references.
type Component = ComponentInterface

// Base provides minimal implementation for alias and ID handling.
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
