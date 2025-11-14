package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// WebSocketCounter demonstrates a counter component with WebSocket support.
type WebSocketCounter struct {
	liveflux.Base
	Count int
}

// GetKind returns the component's kind.
func (c *WebSocketCounter) GetKind() string {
	return "websocket-counter"
}

// Mount initializes the component's state.
func (c *WebSocketCounter) Mount(ctx context.Context, params map[string]string) error {
	c.Count = 0
	return nil
}

// Handle processes actions from the client.
func (c *WebSocketCounter) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "increment":
		c.Count++
	case "decrement":
		c.Count--
	case "set":
		if val := data.Get("value"); val != "" {
			if num, err := strconv.Atoi(val); err == nil {
				c.Count = num
			}
		}
	}
	return nil
}

// HandleWS handles WebSocket messages.
func (c *WebSocketCounter) HandleWS(ctx context.Context, msg *liveflux.WebSocketMessage) (interface{}, error) {
	switch msg.Type {
	case "action":
		switch msg.Action {
		case "increment":
			c.Count++
		case "decrement":
			c.Count--
		case "set":
			// Accept either numeric or string value from client
			var anymap map[string]any
			if err := json.Unmarshal(msg.Data, &anymap); err == nil {
				if raw, ok := anymap["value"]; ok {
					switch v := raw.(type) {
					case float64:
						c.Count = int(v)
					case string:
						if n, err := strconv.Atoi(v); err == nil {
							c.Count = n
						}
					}
				}
			}
		}

		// Return the updated state
		return map[string]interface{}{
			"type":        "update",
			"componentID": c.GetID(),
			"data": map[string]interface{}{
				// Return only the inner container so the client replaces the correct node
				"html": c.inner(ctx).ToHTML(),
			},
		}, nil
	}

	return nil, nil
}

// inner builds the inner container that carries data-flux-component-id and WS hints.
func (c *WebSocketCounter) inner(ctx context.Context) hb.TagInterface {
	div := hb.Div().
		Attr(liveflux.DataFluxComponentID, c.GetID()).
		Attr(liveflux.DataFluxWS, "1").
		Attr(liveflux.DataFluxWSURL, "/liveflux").
		Style("padding: 20px; border: 1px solid #ddd; border-radius: 4px; max-width: 300px; margin: 20px auto;")

	title := hb.H2().Text("WebSocket Counter")
	display := hb.Div().
		Style("font-size: 2rem; margin: 20px 0; text-align: center;")

	// Display the current count
	display = display.Text(strconv.Itoa(c.Count))

	// Buttons with WebSocket actions
	buttons := hb.Div().Style("display: flex; justify-content: space-between;")

	decrBtn := hb.Button().
		Text("-").
		Type("button").
		Attr(liveflux.DataFluxAction, "decrement").
		Style("padding: 8px 16px; font-size: 1.2rem;")

	if c.Count <= 0 {
		decrBtn = decrBtn.Attr("disabled", "disabled")
	}

	buttons = buttons.Children([]hb.TagInterface{
		decrBtn,
		hb.Button().
			Attr(liveflux.DataFluxAction, "increment").
			Style("padding: 8px 16px; font-size: 1.2rem;").
			Text("+"),
	})

	// Form for setting a specific value
	form := hb.Form().
		Attr(liveflux.DataFluxAction, "set").
		Style("margin-top: 20px; display: flex;")

	input := hb.Input().
		Type("number").
		Name("value").
		Value(strconv.Itoa(c.Count)).
		Style("flex-grow: 1; padding: 8px; margin-right: 8px;")

	submitBtn := hb.Button().
		Text("Set").
		Type("submit").
		Style("padding: 8px 16px;")

	form = form.Children([]hb.TagInterface{
		input,
		submitBtn,
	})

	// Status indicator
	status := hb.Div().
		Style("margin-top: 10px; font-size: 0.8em; color: #666;")

	// JavaScript to update the status (wrapped in an IIFE to avoid redeclaration across re-renders)
	script := fmt.Sprintf(`(function(){
        var el = document.querySelector('[data-flux-component-id="%s"]');
        if (!el) return;
        el.addEventListener('flux-ws-open', function(){
            var status = el.querySelector('.status');
            if (status) status.textContent = 'Connected';
        });
        el.addEventListener('flux-ws-close', function(){
            var status = el.querySelector('.status');
            if (status) status.textContent = 'Disconnected';
        });
        el.addEventListener('flux-ws-error', function(e){
            var status = el.querySelector('.status');
            if (status) status.textContent = 'Error: ' + (e.error || 'Unknown error');
        });
    })();`, c.GetID())

	return div.Children([]hb.TagInterface{
		title,
		display,
		buttons,
		form,
		status.Class("status").Text("Connecting..."),
		hb.Script(script),
	})
}

// Render wraps the inner container with the Liveflux root.
func (c *WebSocketCounter) Render(ctx context.Context) hb.TagInterface {
	return c.Root(c.inner(ctx))
}

func init() {
	// Register the component so liveflux.New(&WebSocketCounter{}) works
	if err := liveflux.Register(new(WebSocketCounter)); err != nil {
		log.Fatal(err)
	}
}
