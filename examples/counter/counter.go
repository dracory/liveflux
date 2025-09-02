package main

import (
	"context"
	"net/url"
	"strconv"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// Counter is a minimal Liveflux counter example.
type Counter struct {
	liveflux.Base
	Count int
}

func (c *Counter) GetAlias() string {
	return "counter"
}

// Mount initializes the component's state.
func (c *Counter) Mount(ctx context.Context, params map[string]string) error {
	c.Count = 0
	return nil
}

// Handle processes actions from the client.
func (c *Counter) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "inc":
		c.Count++
	case "dec":
		c.Count--
	case "reset":
		c.Count = 0
	}
	return nil
}

// Render outputs the HTML for the component.
func (c *Counter) Render(ctx context.Context) hb.TagInterface {
	title := hb.H2().Text("Counter")

	display := hb.Div().
		Style("font-size: 2rem; margin: 10px 0;").
		Text(strconv.Itoa(c.Count))

	buttonIncrement := hb.Button().
		Data("flux-action", "inc").
		Text("+1")

	buttonDecrement := hb.Button().
		Data("flux-action", "dec").
		Text("-1")

	buttonReset := hb.Button().
		Data("flux-action", "reset").
		Text("Reset")

	content := hb.Div().
		Child(title).
		Child(display).
		Child(
			hb.Div().
				Child(buttonIncrement).
				Child(buttonDecrement).
				Child(buttonReset),
		)

	return c.Root(content)
}

func init() {
	// Register using default alias derived from type ("counter")
	liveflux.Register(new(Counter))
}
