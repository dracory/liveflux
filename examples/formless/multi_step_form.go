package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// MultiStepForm demonstrates data-flux-include across multiple sections
type MultiStepForm struct {
	liveflux.Base
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Submitted bool
}

func (c *MultiStepForm) GetAlias() string { return "formless.multi-step" }

func (c *MultiStepForm) Mount(ctx context.Context, params map[string]string) error {
	return nil
}

func (c *MultiStepForm) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "submit" {
		c.FirstName = data.Get("first_name")
		c.LastName = data.Get("last_name")
		c.Email = data.Get("email")
		c.Phone = data.Get("phone")
		c.Submitted = true
	}
	return nil
}

func (c *MultiStepForm) Render(ctx context.Context) hb.TagInterface {
	if c.Submitted {
		summary := hb.Div().Class("alert alert-success").
			Child(hb.H4().Text("Registration Complete!")).
			Child(hb.P().Text(fmt.Sprintf("Name: %s %s", c.FirstName, c.LastName))).
			Child(hb.P().Text(fmt.Sprintf("Email: %s", c.Email))).
			Child(hb.P().Text(fmt.Sprintf("Phone: %s", c.Phone)))

		return c.Root(summary)
	}

	content := hb.Div().Class("card").
		Child(hb.Div().Class("card-header").
			Child(hb.H5().Class("mb-0").Text("Multi-Step Registration"))).
		Child(hb.Div().Class("card-body").
			Child(hb.P().Class("text-muted").
				Text("This form demonstrates data-flux-include collecting fields from multiple sections outside the component root.")).
			Child(hb.Button().
				Class("btn btn-success btn-lg").
				Attr(liveflux.DataFluxAction, "submit").
				Attr(liveflux.DataFluxInclude, liveflux.IncludeSelectors("#step-1", "#step-2")).
				Text("Complete Registration")))

	return c.Root(content)
}
