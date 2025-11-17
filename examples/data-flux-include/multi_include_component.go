package main

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// MultiIncludeComponent demonstrates including multiple checkboxes
// using data-flux-include to submit all selected IDs at once.
type MultiIncludeComponent struct {
	liveflux.Base

	Items         []MultiItem
	SelectedIDs   map[string]bool
	LastSubmitted []string
}

type MultiItem struct {
	ID   string
	Name string
}

func (c *MultiIncludeComponent) GetKind() string {
	return "include.multi"
}

func (c *MultiIncludeComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Items = []MultiItem{
		{ID: "1", Name: "Email notifications"},
		{ID: "2", Name: "SMS alerts"},
		{ID: "3", Name: "Product updates"},
		{ID: "4", Name: "Weekly summary"},
	}
	c.SelectedIDs = map[string]bool{}
	c.LastSubmitted = nil
	return nil
}

func (c *MultiIncludeComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "apply-selection":
		ids := data["selected_ids"]

		// Reset and mark which IDs are currently selected.
		c.SelectedIDs = map[string]bool{}
		for _, id := range ids {
			c.SelectedIDs[id] = true
		}

		// Store what was submitted last for display.
		c.LastSubmitted = append([]string(nil), ids...)
	}
	return nil
}

func (c *MultiIncludeComponent) Render(ctx context.Context) hb.TagInterface {
	return c.Root(
		hb.Div().Children([]hb.TagInterface{
			hb.Div().Children([]hb.TagInterface{
				hb.Span().Class("pill").Text("Collects multiple checkboxes with data-flux-include"),
			}),
			c.renderList(),
			c.renderFooter(),
		}),
	)
}

func (c *MultiIncludeComponent) renderList() hb.TagInterface {
	ul := hb.UL()
	for _, item := range c.Items {
		checked := ""
		if c.SelectedIDs[item.ID] {
			checked = "checked"
		}

		ul.Child(
			hb.LI().Children([]hb.TagInterface{
				hb.Input().
					Attr("type", "checkbox").
					Attr("name", "selected_ids").
					Attr("value", item.ID).
					AttrIf(checked != "", "checked", "checked"),
				hb.Span().Text(item.Name),
			}),
		)
	}

	return hb.Div().Class("items").Child(ul)
}

func (c *MultiIncludeComponent) renderFooter() hb.TagInterface {
	// Build a human-readable description of the last submitted IDs.
	label := "Nothing submitted yet. Choose some options and click Apply selection."
	if len(c.LastSubmitted) > 0 {
		copyIDs := append([]string(nil), c.LastSubmitted...)
		sort.Strings(copyIDs)
		label = fmt.Sprintf("Last submitted IDs: %v", copyIDs)
	}

	return hb.Div().Children([]hb.TagInterface{
		hb.Div().Class("status").Text(label),
		hb.Div().Style("margin-top: 8px;").Children([]hb.TagInterface{
			// Note: this includes all checked checkboxes inside this component root.
			hb.Button().
				Class("btn secondary").
				Attr(liveflux.DataFluxAction, "apply-selection").
				Attr(liveflux.DataFluxInclude, "[name='selected_ids']:checked").
				Text("Apply selection"),
		}),
	})
}
