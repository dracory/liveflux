package main

import (
	"context"
	"net/url"
	"strconv"

	"github.com/dracory/liveflux"
	"github.com/gouniverse/hb"
)

// Counter is a minimal Liveflux counter example.
type Counter struct {
	liveflux.Base
	Count int
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
	content := hb.Div().
		Child(hb.H2().
			Text("Counter")).
		Child(hb.Div().
			Style("font-size: 2rem; margin: 10px 0;").
			Text(strconv.Itoa(c.Count))).
		Child(
			hb.Div().
				Child(hb.Button().
					Class("btn btn-primary me-2").
					Attr("data-lw-action", "inc").
					Text("+1")).
				Child(hb.Button().
					Class("btn btn-secondary me-2").
					Attr("data-lw-action", "dec").
					Text("-1")).
				Child(hb.Button().
					Class("btn btn-outline-danger").
					Attr("data-lw-action", "reset").
					Text("Reset")),
		)

	return c.Root(content)
}

func init() {
	// Register using default alias derived from type ("counter")
	liveflux.Register(func() liveflux.ComponentInterface { return &Counter{} })
}
