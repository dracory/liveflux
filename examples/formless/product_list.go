package main

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// ProductList demonstrates data-flux-include with shared filters
type ProductList struct {
	liveflux.Base
	Products []string
	Search   string
	Category string
}

func (c *ProductList) GetAlias() string { return "formless.product-list" }

func (c *ProductList) Mount(ctx context.Context, params map[string]string) error {
	c.Products = []string{"Laptop", "Mouse", "Keyboard", "Monitor", "Headphones"}
	c.Search = params["search"]
	c.Category = params["category"]
	return nil
}

func (c *ProductList) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "refresh" {
		c.Search = data.Get("search")
		c.Category = data.Get("category")
	}
	return nil
}

func (c *ProductList) Render(ctx context.Context) hb.TagInterface {
	filtered := c.Products
	if c.Search != "" {
		var result []string
		for _, p := range c.Products {
			if contains(p, c.Search) {
				result = append(result, p)
			}
		}
		filtered = result
	}

	list := hb.Ul().Class("list-group")
	if len(filtered) == 0 {
		list.Child(hb.Li().Class("list-group-item").Text("No products found"))
	} else {
		for _, p := range filtered {
			list.Child(hb.Li().Class("list-group-item").Text(p))
		}
	}

	card := hb.Div().Class("card").
		Child(hb.Div().Class("card-header").
			Child(hb.H5().Class("mb-0").Text("Products"))).
		Child(hb.Div().Class("card-body").
			Child(list).
			Child(hb.Div().Class("mt-3").
				Child(hb.Button().
					Class("btn btn-primary").
					Attr(liveflux.DataFluxAction, "refresh").
					Attr(liveflux.DataFluxInclude, "#global-filters").
					Text("Refresh Products"))))

	return c.Root(card)
}
