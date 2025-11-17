package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// SingleIncludeComponent demonstrates including a single element that
// lives outside the component using data-flux-include.
type SingleIncludeComponent struct {
	liveflux.Base

	LastQuery string
}

func (c *SingleIncludeComponent) GetKind() string {
	return "include.single"
}

func (c *SingleIncludeComponent) Mount(ctx context.Context, params map[string]string) error {
	c.LastQuery = "(none yet)"
	return nil
}

func (c *SingleIncludeComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "apply-filter":
		// global_query comes from the input rendered in main.go toolbar.
		q := data.Get("global_query")
		if q == "" {
			c.LastQuery = "(empty)"
		} else {
			c.LastQuery = q
		}
	}
	return nil
}

func (c *SingleIncludeComponent) Render(ctx context.Context) hb.TagInterface {
	return c.Root(
		hb.Div().Children([]hb.TagInterface{
			hb.Div().Children([]hb.TagInterface{
				hb.Span().Class("pill").Text("Uses data-flux-include on a single external field"),
			}),
			hb.Div().Class("status").Text(fmt.Sprintf("Last applied query: %s", c.LastQuery)),
			hb.Div().Style("margin-top: 8px;").Children([]hb.TagInterface{
				hb.Button().
					Class("btn").
					Attr(liveflux.DataFluxAction, "apply-filter").
					Attr(liveflux.DataFluxInclude, "#global-query").
					Text("Apply filter using toolbar input"),
			}),
		}),
	)
}
