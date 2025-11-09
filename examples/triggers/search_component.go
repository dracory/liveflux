package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

const actionSearch = "search"
const actionClear = "clear"
const targetSearchClear = "wrapper-search-clear"
const targetSearchQuery = "wrapper-search-query"
const targetSearchResults = "wrapper-search-results"

// SearchComponent demonstrates live search using data-flux-trigger
type SearchComponent struct {
	liveflux.Base
	Query   string
	Results []string
}

// Implement TargetRenderer to use targeted updates
var _ liveflux.TargetRenderer = (*SearchComponent)(nil)

// GetKind returns the component kind
func (c *SearchComponent) GetKind() string {
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
	case actionSearch:
		c.Query = data.Get("query")
		c.handleSearch()
		// Mark the results as dirty for targeted update
		c.MarkTargetDirty(targetSearchResults)
	case actionClear:
		c.Query = ""
		c.Results = []string{}
		c.MarkTargetDirty(targetSearchResults)
		c.MarkTargetDirty(targetSearchQuery)
	}
	return nil
}

// RenderTargets returns only the changed fragments
func (c *SearchComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
	var targets []liveflux.TargetFragment

	if c.IsDirty(targetSearchResults) {
		targets = append(targets, liveflux.TargetFragment{
			Selector: liveflux.TargetID(targetSearchResults),
			Content:  c.renderSearchResults(),
			SwapMode: liveflux.SwapReplace,
		})
	}

	if c.IsDirty(targetSearchQuery) {
		targets = append(targets, liveflux.TargetFragment{
			Selector: liveflux.TargetID(targetSearchQuery),
			Content:  c.renderSearchQuery(),
			SwapMode: liveflux.SwapReplace,
		})
	}

	targets = append(targets, liveflux.TargetFragment{
		Selector: liveflux.TargetID(targetSearchClear),
		Content:  c.renderSearchClear(),
		SwapMode: liveflux.SwapReplace,
	})

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
		Child(hb.Div().
			Class("card-header").
			Child(hb.H5().
				Class("mb-0").
				Text("Live Search Demo"))).
		Child(hb.Div().
			Class("card-body").
			Child(c.renderSearchQuery()).
			Child(c.renderSearchResults()).
			Child(c.renderSearchClear()))

	return c.Root(content)
}

// renderSearchResults renders the search results
func (c *SearchComponent) renderSearchResults() hb.TagInterface {
	root := hb.Div().
		ID(targetSearchResults)

	if c.Query == "" {
		return root.Child(hb.Div().
			Class("text-muted").
			Text("Start typing to search..."))
	}

	if len(c.Results) == 0 {
		return root.Child(hb.Div().
			Class("alert alert-info").
			Text(fmt.Sprintf("No results found for \"%s\"", c.Query)))
	}

	items := []hb.TagInterface{}
	for _, result := range c.Results {
		items = append(items, hb.Li().Class("list-group-item").Text(result))
	}

	return root.Child(hb.Div().
		Children([]hb.TagInterface{
			hb.P().
				Class("text-muted").
				Text(fmt.Sprintf("Found %d result(s):", len(c.Results))),
			hb.Ul().
				Class("list-group").
				Children(items),
		}))
}

// renderSearchQuery renders the search query group
func (c *SearchComponent) renderSearchQuery() hb.TagInterface {
	label := hb.Label().
		Class("form-label").
		Text("Search fruits:")
	input := hb.Input().
		Type("text").
		Name("query").
		Value(c.Query).
		Class("form-control").
		Placeholder("Type to search...").
		Attr("data-flux-trigger", "keyup changed delay:500ms").
		Attr("data-flux-action", actionSearch)

	debounce := hb.Small().
		Class("form-text text-muted").
		Text("Searches as you type with 500ms debounce")

	groupSearch := hb.Div().
		ID(targetSearchQuery).
		Class("mb-3").
		Child(label).
		Child(input).
		Child(debounce)

	return groupSearch
}

func (c *SearchComponent) renderSearchClear() hb.TagInterface {
	button := hb.Button().
		Class("btn btn-danger").
		Text("Clear").
		Attr("data-flux-action", actionClear)

	return hb.Div().
		ID(targetSearchClear).
		ChildIf(len(c.Results) > 0, button)
}
