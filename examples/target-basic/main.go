package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// CartComponent demonstrates targeted fragment updates
type CartComponent struct {
	liveflux.Base
	Total float64
	Items []CartItem
}

type CartItem struct {
	ID    int
	Name  string
	Price float64
}

func (c *CartComponent) GetKind() string {
	return "cart"
}

func (c *CartComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Total = 99.99
	c.Items = []CartItem{
		{ID: 1, Name: "Widget A", Price: 49.99},
		{ID: 2, Name: "Widget B", Price: 50.00},
	}
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

	case "remove-item":
		if len(c.Items) > 0 {
			removedItem := c.Items[len(c.Items)-1]
			c.Items = c.Items[:len(c.Items)-1]
			c.Total -= removedItem.Price

			// Mark targets as dirty
			c.MarkTargetDirty("#cart-total")
			c.MarkTargetDirty(".line-items")
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

// RenderTargets implements the TargetRenderer interface
func (c *CartComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
	fragments := []liveflux.TargetFragment{}

	// Only render fragments for dirty targets
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

func main() {
	// Register the component
	if err := liveflux.Register(&CartComponent{}); err != nil {
		log.Fatal(err)
	}

	// Create handler
	handler := liveflux.NewHandler(nil)

	// Serve the example page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := hb.Webpage().
			SetTitle("Liveflux Target Example - Shopping Cart").
			SetCharset("utf-8").
			Style(`
				body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
				.cart-container { border: 2px solid #333; padding: 20px; border-radius: 8px; }
			`).
			Child(
				hb.Div().Class("container").Children([]hb.TagInterface{
					hb.H1().Text("Liveflux Targeted Updates Example"),
					hb.P().Text("This example demonstrates targeted fragment updates. When you click the buttons, only the cart total and item list are updated, not the entire component."),
					liveflux.PlaceholderByKind("cart"),
					liveflux.Script(),
				}),
			)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page.ToHTML()))
	})

	// Mount the Liveflux handler
	http.Handle("/liveflux", handler)

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Open your browser and check the Network tab to see targeted updates in action!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
