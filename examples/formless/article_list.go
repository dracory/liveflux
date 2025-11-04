package main

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// ArticleList demonstrates the same shared filters
type ArticleList struct {
	liveflux.Base
	Articles []string
	Search   string
	Category string
}

func (c *ArticleList) GetAlias() string { return "formless.article-list" }

func (c *ArticleList) Mount(ctx context.Context, params map[string]string) error {
	c.Articles = []string{"Getting Started", "Advanced Tips", "Best Practices", "Troubleshooting"}
	c.Search = params["search"]
	c.Category = params["category"]
	return nil
}

func (c *ArticleList) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "refresh" {
		c.Search = data.Get("search")
		c.Category = data.Get("category")
	}
	return nil
}

func (c *ArticleList) Render(ctx context.Context) hb.TagInterface {
	filtered := c.Articles
	if c.Search != "" {
		var result []string
		for _, a := range c.Articles {
			if contains(a, c.Search) {
				result = append(result, a)
			}
		}
		filtered = result
	}

	list := hb.Ul().Class("list-group")
	if len(filtered) == 0 {
		list.Child(hb.Li().Class("list-group-item").Text("No articles found"))
	} else {
		for _, a := range filtered {
			list.Child(hb.Li().Class("list-group-item").Text(a))
		}
	}

	card := hb.Div().Class("card").
		Child(hb.Div().Class("card-header").
			Child(hb.H5().Class("mb-0").Text("Articles"))).
		Child(hb.Div().Class("card-body").
			Child(list).
			Child(hb.Div().Class("mt-3").
				Child(hb.Button().
					Class("btn btn-primary").
					Attr(liveflux.DataFluxAction, "refresh").
					Attr(liveflux.DataFluxInclude, "#global-filters").
					Text("Refresh Articles"))))

	return c.Root(card)
}
