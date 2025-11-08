package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// SearchComponent demonstrates live search using data-flux-trigger
type SearchComponent struct {
	liveflux.Base
	Query   string
	Results []string
}

// Implement TargetRenderer to use targeted updates
var _ liveflux.TargetRenderer = (*SearchComponent)(nil)

// GetAlias returns the component alias
func (c *SearchComponent) GetAlias() string {
	return "search"
}

// Mount initializes the component
func (c *SearchComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Query = ""
	c.Results = []string{}
	return nil
}

// Handle processes actions from the client
func (c *SearchComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "search":
		c.Query = data.Get("query")
		c.handleSearch()
		// Mark the results as dirty for targeted update
		c.MarkTargetDirty("results")
	}
	return nil
}

// RenderTargets returns only the changed fragments
func (c *SearchComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
	var targets []liveflux.TargetFragment

	if c.IsDirty("results") {
		targets = append(targets, liveflux.TargetFragment{
			Selector: "#search-results",
			Content:  c.renderResults(),
		})
	}

	return targets
}

// handleSearch performs the search action
func (c *SearchComponent) handleSearch() {
	// Simulate search - filter from a static dataset
	allItems := []string{
		"Apple", "Apricot", "Avocado",
		"Banana", "Blueberry", "Blackberry",
		"Cherry", "Cranberry", "Coconut",
		"Date", "Dragonfruit", "Durian",
		"Elderberry", "Fig", "Grape",
		"Grapefruit", "Guava", "Honeydew",
		"Kiwi", "Lemon", "Lime",
		"Mango", "Melon", "Nectarine",
		"Orange", "Papaya", "Peach",
		"Pear", "Pineapple", "Plum",
		"Pomegranate", "Raspberry", "Strawberry",
		"Tangerine", "Watermelon",
	}

	c.Results = []string{}
	if c.Query != "" {
		query := strings.ToLower(c.Query)
		for _, item := range allItems {
			if strings.Contains(strings.ToLower(item), query) {
				c.Results = append(c.Results, item)
			}
		}
	}
}

// Render renders the component
func (c *SearchComponent) Render(ctx context.Context) hb.TagInterface {
	content := hb.Div().
		Class("card").
		Children([]hb.TagInterface{
			hb.Div().Class("card-header").
				Child(hb.H5().Class("mb-0").Text("Live Search Demo")),
			hb.Div().Class("card-body").
				Children([]hb.TagInterface{
					// Search input with trigger
					hb.Div().Class("mb-3").
						Children([]hb.TagInterface{
							hb.Label().Class("form-label").Text("Search fruits:"),
							hb.Input().
								Type("text").
								Name("query").
								Value(c.Query).
								Class("form-control").
								Placeholder("Type to search...").
								Attr("data-flux-trigger", "keyup changed delay:3000ms").
								Attr("data-flux-action", "search"),
							hb.Small().Class("form-text text-muted").
								Text("Searches as you type with 3000ms debounce"),
						}),
					// Results container with ID for targeted updates
					hb.Div().
						Attr("id", "search-results").
						Child(c.renderResults()),
				}),
		})

	return c.Root(content)
}

// renderResults renders the search results
func (c *SearchComponent) renderResults() hb.TagInterface {
	if c.Query == "" {
		return hb.Div().Class("text-muted").Text("Start typing to search...")
	}

	if len(c.Results) == 0 {
		return hb.Div().Class("alert alert-info").Text(fmt.Sprintf("No results found for \"%s\"", c.Query))
	}

	items := []hb.TagInterface{}
	for _, result := range c.Results {
		items = append(items, hb.Li().Class("list-group-item").Text(result))
	}

	return hb.Div().
		Children([]hb.TagInterface{
			hb.P().Class("text-muted").Text(fmt.Sprintf("Found %d result(s):", len(c.Results))),
			hb.Ul().Class("list-group").Children(items),
		})
}
