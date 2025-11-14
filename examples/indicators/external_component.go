package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// ExternalIndicatorDemo shows how a component can toggle an indicator outside its root.
type ExternalIndicatorDemo struct {
	liveflux.Base
	Message string
}

func (c *ExternalIndicatorDemo) GetKind() string {
	return "indicators-external"
}

func (c *ExternalIndicatorDemo) Mount(ctx context.Context, params map[string]string) error {
	c.Message = "Global indicator is idle"
	return nil
}

func (c *ExternalIndicatorDemo) Handle(ctx context.Context, action string, data url.Values) error {
	if action != "ping" {
		return nil
	}

	time.Sleep(700 * time.Millisecond)
	c.Message = fmt.Sprintf("Ping acknowledged at %s", time.Now().Format("15:04:05"))
	return nil
}

func (c *ExternalIndicatorDemo) Render(ctx context.Context) hb.TagInterface {
	status := hb.Div().
		Attr("id", "external-status").
		Class("mb-3 text-secondary fw-semibold").
		Text(c.Message)

	button := hb.Button().
		Class("btn btn-outline-primary").
		Attr(liveflux.DataFluxAction, "ping").
		Attr(liveflux.DataFluxSelect, "#external-status").
		Attr(liveflux.DataFluxIndicator, "#global-indicator").
		Text("Ping server")

	content := hb.Div().
		Class("card p-4 shadow-sm").
		Child(hb.H2().Text("Component with External Indicator")).
		Child(hb.Paragraph().Text("This component toggles a page-level indicator when the button is clicked.")).
		Child(status).
		Child(button)

	return c.Root(content)
}

func init() {
	if err := liveflux.Register(new(ExternalIndicatorDemo)); err != nil {
		log.Fatal(err)
	}
}
