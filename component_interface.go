package liveflux

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
)

// ComponentInterface defines the contract for a server-driven UI component.
//
// Lifecycle:
// - New(kind) via registry (framework sets the component's kind via SetKind(kind))
// - SetID(...) during first mount (framework assigns a per-instance ID)
// - Mount(ctx, params) on first initialization
// - Handle(ctx, action, form) on user actions
// - Render(ctx) returns the current HTML (hb.Tag)
//
// Kind is assigned by the framework when the instance is created from the registry.
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
//	liveflux.RegisterByKind("counter", func() liveflux.ComponentInterface { return &Counter{} })
//
// See handler.go for the HTTP entry point.
type ComponentInterface interface {
	// GetKind returns the stable routing key (TYPE identifier) used by the client
	// and registry to select which component to construct/route to.
	// Component authors MUST implement this.
	GetKind() string

	// SetKind sets the component's kind (TYPE identifier). The framework sets this
	// once during construction in the registry. Implementations should treat kind
	// as immutable; Base enforces set-once semantics.
	SetKind(kind string)

	// GetID returns the instance ID (per-mount INSTANCE identifier).
	GetID() string

	// SetID sets the instance ID (per-mount INSTANCE identifier). This is called by
	// the framework during mount (see handler.go).
	SetID(id string)

	Mount(ctx context.Context, params map[string]string) error
	Handle(ctx context.Context, action string, data url.Values) error
	Render(ctx context.Context) hb.TagInterface
}
