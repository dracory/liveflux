package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// IndicatorDemo demonstrates request indicators reacting to server latency.
type fetchDataComponent struct {
	liveflux.Base
	Status string
}

func (c *fetchDataComponent) GetKind() string {
	return "fetch-data"
}

func (c *fetchDataComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Status = "Ready to fetch"
	return nil
}

func (c *fetchDataComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "fetch":
		fmt.Printf("[%s] starting fetch\n", c.GetID())
		time.Sleep(1 * time.Second)
		c.Status = fmt.Sprintf("Fetched at %s", time.Now().Format("15:04:05"))
		fmt.Printf("[%s] finished fetch\n", c.GetID())
	}
	return nil
}

func (c *fetchDataComponent) Render(ctx context.Context) hb.TagInterface {
	title := hb.H2().Text("Fetch Data Demo Component")

	status := hb.Paragraph().
		Attr("id", "status-text").
		Class("mb-3").
		Text(c.Status)

	spinner := hb.Span().
		Class("demo-spinner spinner-border spinner-border-sm align-middle ms-2").
		Style("display: none;").
		Attr("role", "status").
		Child(hb.Span().Class("visually-hidden").Text("Loading"))

	button := hb.Button().
		Class("btn btn-primary").
		Attr(liveflux.DataFluxAction, "fetch").
		Attr(liveflux.DataFluxSelect, "#status-text").
		Attr(liveflux.DataFluxIndicator, "this, .demo-spinner").
		Text("Fetch data").
		Child(spinner)

	content := hb.Div().
		Class("card p-4 shadow-sm").
		Child(title).
		Child(hb.Paragraph().Text("Click the button to simulate a slow server call.")).
		Child(status).
		Child(button)

	return c.Root(content)
}

func init() {
	liveflux.Register(new(fetchDataComponent))
}
