package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// IndicatorForm demonstrates indicators on form submissions.
type IndicatorForm struct {
	liveflux.Base
	Name    string
	Message string
}

func (c *IndicatorForm) GetKind() string {
	return "indicators-form"
}

func (c *IndicatorForm) Mount(ctx context.Context, params map[string]string) error {
	c.Message = "Enter a name and submit"
	return nil
}

func (c *IndicatorForm) Handle(ctx context.Context, action string, data url.Values) error {
	if action != "submit" {
		return nil
	}

	name := strings.TrimSpace(data.Get("name"))
	if name == "" {
		name = "Anonymous"
	}

	time.Sleep(850 * time.Millisecond)
	c.Message = fmt.Sprintf("Saved profile for %s", name)
	return nil
}

func (c *IndicatorForm) Render(ctx context.Context) hb.TagInterface {
	status := hb.Div().
		Attr("id", "form-status").
		Class("alert alert-secondary py-2 px-3 mb-3").
		Text(c.Message)

	spinner := hb.Span().
		Attr("id", "form-spinner").
		Class("spinner-border spinner-border-sm align-middle ms-2 text-light").
		Style("display: none;").
		Attr("role", "status").
		Child(hb.Span().Class("visually-hidden").Text("Saving"))

	submit := hb.Button().
		Attr("type", "submit").
		Class("btn btn-success").
		Attr(liveflux.DataFluxAction, "submit").
		Attr(liveflux.DataFluxIndicator, "this, #form-spinner").
		Attr(liveflux.DataFluxSelect, "#form-status").
		Text("Save Profile").
		Child(spinner)

	form := hb.Form().
		Attr("autocomplete", "off").
		Child(
			hb.Div().Class("mb-3").
				Child(hb.Label().Attr("for", "name-input").Class("form-label").Text("Name")).
				Child(hb.Input().Attr("id", "name-input").Attr("name", "name").Class("form-control").Attr("placeholder", "Jane Doe")),
		).
		Child(submit)

	content := hb.Div().
		Class("card p-4 shadow-sm").
		Child(hb.H2().Text("Form Component with Indicator")).
		Child(hb.Paragraph().Text("Submit the form to see the spinner appear while the server processes.")).
		Child(status).
		Child(form)

	return c.Root(content)
}

func init() {
	if err := liveflux.Register(new(IndicatorForm)); err != nil {
		log.Fatal(err)
	}
}
