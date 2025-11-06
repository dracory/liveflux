package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// CartItem represents a single line item in the cart.
type CartItem struct {
	ID    int
	Name  string
	Price float64
}

// CartComponent demonstrates targeted fragment updates across multiple scopes.
type CartComponent struct {
	liveflux.Base
	Total float64
	Items []CartItem
}

func (c *CartComponent) GetAlias() string {
	return "cart"
}

func (c *CartComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Total = 99.99
	c.Items = []CartItem{
		{ID: 1, Name: "Widget A", Price: 49.99},
		{ID: 2, Name: "Widget B", Price: 50.00},
	}

	// Ensure targeted fragments render during the initial mount
	c.MarkTargetDirty("#cart-total")
	c.MarkTargetDirty(".line-items")
	c.MarkTargetDirty("#global-cart-badge")
	return nil
}

func (c *CartComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "add-item":
		// Add a new item
		newID := len(c.Items) + 1
		newItem := CartItem{
			ID:    newID,
			Name:  fmt.Sprintf("Widget %c", 'A'+newID-1),
			Price: 25.00,
		}
		c.Items = append(c.Items, newItem)
		c.Total += newItem.Price

		// Mark only the changed targets as dirty
		c.MarkTargetDirty("#cart-total")
		c.MarkTargetDirty(".line-items")
		c.MarkTargetDirty("#global-cart-badge")

	case "remove-item":
		if len(c.Items) > 0 {
			removedItem := c.Items[len(c.Items)-1]
			c.Items = c.Items[:len(c.Items)-1]
			c.Total -= removedItem.Price

			// Mark targets as dirty
			c.MarkTargetDirty("#cart-total")
			c.MarkTargetDirty(".line-items")
			c.MarkTargetDirty("#global-cart-badge")
		}
	}
	return nil
}

func (c *CartComponent) Render(ctx context.Context) hb.TagInterface {
	return c.Root(
		hb.Div().Class("cart-container").Children([]hb.TagInterface{
			hb.H2().Text("Shopping Cart"),
			c.renderTotal(),
			c.renderItems(),
			c.renderActions(),
		}),
	)
}

// RenderTargets implements the TargetRenderer interface to send only changed fragments.
func (c *CartComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
	fragments := []liveflux.TargetFragment{}

	if c.IsDirty("#cart-total") {
		fragments = append(fragments, liveflux.TargetFragment{
			Selector: "#cart-total",
			Content:  c.renderTotal(),
			SwapMode: liveflux.SwapReplace,
		})
	}

	if c.IsDirty(".line-items") {
		fragments = append(fragments, liveflux.TargetFragment{
			Selector: ".line-items",
			Content:  c.renderItems(),
			SwapMode: liveflux.SwapReplace,
		})
	}

	if c.IsDirty("#global-cart-badge") {
		fragments = append(fragments, liveflux.TargetFragment{
			Selector:            "#global-cart-badge",
			Content:             c.renderGlobalBadge(),
			SwapMode:            liveflux.SwapInner,
			NoComponentMetadata: true,
		})
	}

	return fragments
}

func (c *CartComponent) renderTotal() hb.TagInterface {
	return hb.Div().
		ID("cart-total").
		Class("cart-total").
		Style("font-size: 1.5em; font-weight: bold; margin: 20px 0;").
		Text(fmt.Sprintf("Total: $%.2f", c.Total))
}

func (c *CartComponent) renderItems() hb.TagInterface {
	ul := hb.UL().Class("line-items").Style("list-style: none; padding: 0;")
	for _, item := range c.Items {
		ul.Child(
			hb.LI().
				Class("line-item").
				Attr("data-id", fmt.Sprintf("%d", item.ID)).
				Style("padding: 10px; border: 1px solid #ddd; margin: 5px 0;").
				Text(fmt.Sprintf("%s - $%.2f", item.Name, item.Price)),
		)
	}
	return ul
}

func (c *CartComponent) renderActions() hb.TagInterface {
	return hb.Div().Class("cart-actions").Style("margin-top: 20px;").Children([]hb.TagInterface{
		hb.Button().
			Attr(liveflux.DataFluxAction, "add-item").
			Style("padding: 10px 20px; margin-right: 10px; cursor: pointer;").
			Text("Add Item"),
		hb.Button().
			Attr(liveflux.DataFluxAction, "remove-item").
			Style("padding: 10px 20px; cursor: pointer;").
			Text("Remove Item"),
	})
}

func (c *CartComponent) renderGlobalBadge() hb.TagInterface {
	return hb.Span().
		ID("global-cart-badge").
		Class("cart-badge").
		Text(fmt.Sprintf("Cart: %d items • Total $%.2f", len(c.Items), c.Total))
}
