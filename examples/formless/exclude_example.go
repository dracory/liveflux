package main

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// ExcludeExample demonstrates data-flux-exclude
type ExcludeExample struct {
	liveflux.Base
	Username string
	Bio      string
	Message  string
}

func (c *ExcludeExample) GetKind() string { return "formless.exclude-example" }

func (c *ExcludeExample) Mount(ctx context.Context, params map[string]string) error {
	c.Username = "john_doe"
	c.Bio = "Software developer"
	return nil
}

func (c *ExcludeExample) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "update-profile" {
		c.Username = data.Get("username")
		c.Bio = data.Get("bio")
		// Note: password field is excluded via data-flux-exclude
		c.Message = "Profile updated (password not changed)"
	}
	return nil
}

func (c *ExcludeExample) Render(ctx context.Context) hb.TagInterface {
	var alert hb.TagInterface
	if c.Message != "" {
		alert = hb.Div().Class("alert alert-info").Text(c.Message)
	}

	content := hb.Div().Class("card").
		Child(hb.Div().Class("card-header").
			Child(hb.H5().Class("mb-0").Text("Exclude Example"))).
		Child(hb.Div().Class("card-body").
			ChildIf(alert != nil, alert).
			Child(hb.P().Class("text-muted").
				Text("This demonstrates data-flux-exclude to omit sensitive fields.")).
			Child(hb.Button().
				Class("btn btn-primary").
				Attr(liveflux.DataFluxAction, "update-profile").
				Attr(liveflux.DataFluxInclude, "#user-form").
				Attr(liveflux.DataFluxExclude, ".sensitive").
				Text("Update Profile (without password)")))

	return c.Root(content)
}
