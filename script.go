package liveflux

import (
	_ "embed"
	"encoding/json"

	"github.com/dracory/hb"
	"github.com/samber/lo"
)

// Embed split client-side JS parts. Order matters for concatenation.
//
//go:embed js/util.js
var clientUtilJS string

//go:embed js/network.js
var clientNetworkJS string

//go:embed js/mount.js
var clientMountJS string

//go:embed js/handlers.js
var clientHandlersJS string

//go:embed js/bootstrap.js
var clientBootstrapJS string

//go:embed js/websocket.js
var clientWebSocketJS string

// baseJS concatenates embedded client modules.
func baseJS(includeWS bool) string {
	js := clientUtilJS + "\n" +
		clientNetworkJS + "\n" +
		clientMountJS + "\n" +
		clientHandlersJS + "\n" +
		clientBootstrapJS
	if includeWS {
		js += "\n" + clientWebSocketJS
	}
	return js
}

// JS returns the Liveflux client script. Optional ClientOptions configure the client
// (merged into window.__lw before the runtime). Include once per page.
func JS(opts ...ClientOptions) string {
	// Pick first options or zero value, then apply sensible defaults.
	o := lo.FirstOr(opts, ClientOptions{})

	if o.Endpoint == "" {
		o.Endpoint = "/liveflux"
	}

	// Note: WebSocketURL is optional. If not provided, the WS client will
    // fall back to using Endpoint on the client side.

	if o.Headers == nil {
		o.Headers = map[string]string{}
	}

	if o.RedirectHeader == "" {
		o.RedirectHeader = RedirectHeader
	}
	if o.RedirectAfterHeader == "" {
		o.RedirectAfterHeader = RedirectAfterHeader
	}

	b, _ := json.Marshal(o)

	cfg := "(function(){var o=" + string(b) + ";window.__lw=Object.assign({},window.__lw||{},o);})();\n"

	return cfg + baseJS(o.UseWebSocket)
}

// Script returns an hb.Script tag containing the client JS with optional configuration.
func Script(opts ...ClientOptions) hb.TagInterface {
	return hb.Script(JS(opts...))
}

// ClientOptions configures the embedded client.
// All fields are optional; zero values are ignored.
type ClientOptions struct {
	Endpoint    string            `json:"endpoint,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Credentials string            `json:"credentials,omitempty"` // e.g., "same-origin", "include"
	TimeoutMs   int               `json:"timeoutMs,omitempty"`   // request timeout; 0 = no timeout
	// Names of response headers used for client-side redirects
	RedirectHeader      string `json:"redirectHeader,omitempty"`
	RedirectAfterHeader string `json:"redirectAfterHeader,omitempty"`

	// WebSocket integration
	UseWebSocket bool   `json:"useWebSocket,omitempty"`
	WebSocketURL string `json:"wsEndpoint,omitempty"`
}
