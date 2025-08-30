package liveflux

import (
	"context"

	"github.com/gouniverse/hb"
	"github.com/samber/lo"
)

// SSR mounts the component on the server and returns its rendered HB tag.
// Useful for SEO/static-first rendering while still enabling JS-driven updates
// after the client runtime hydrates.
func SSR(c ComponentInterface, params ...map[string]string) hb.TagInterface {
	p := lo.FirstOr(params, map[string]string{})
	if c == nil {
		return hb.Text("component missing")
	}

	ctx := context.Background()

	// Ensure instance has ID for subsequent client actions/state persistence
	if c.GetID() == "" {
		c.SetID(NewID())
	}

	// Initialize component state
	if err := c.Mount(ctx, p); err != nil {
		return hb.Div().Class("alert alert-danger").Text("mount error: " + err.Error())
	}

	// Persist for later actions
	StoreDefault.Set(c)

	return c.Render(ctx)
}

// SSRHTML mounts and renders the component, returning HTML as string.
func SSRHTML(c ComponentInterface, params ...map[string]string) string {
	tag := SSR(c, params...)
	if tag == nil {
		return ""
	}
	return tag.ToHTML()
}
