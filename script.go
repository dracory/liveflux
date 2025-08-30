package liveflux

import (
	_ "embed"

	"github.com/gouniverse/hb"
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

// JS returns the standard Liveflux client script used to mount and act on components.
// Include once per page (e.g., in your base layout) either via hb.Script(liveflux.JS())
// or using liveflux.Script().
func JS() string {
	return clientUtilJS + "\n" +
		clientNetworkJS + "\n" +
		clientMountJS + "\n" +
		clientHandlersJS + "\n" +
		clientBootstrapJS
}

// Script returns an hb.Script tag containing the standard client JS.
func Script() hb.TagInterface { return hb.Script(JS()) }

// JSWithEndpoint returns the client script prefixed with a small configuration
// snippet that sets the transport endpoint (defaults to "/liveflux").
// Example: hb.Script(liveflux.JSWithEndpoint("/api/liveflux"))
func JSWithEndpoint(endpoint string) string {
	cfg := "(function(){window.__lw=window.__lw||{};window.__lw.endpoint='" + endpoint + "';})();\n"
	return cfg + JS()
}

// ScriptWithEndpoint returns an hb.Script tag that sets the endpoint and embeds the client JS.
func ScriptWithEndpoint(endpoint string) hb.TagInterface { return hb.Script(JSWithEndpoint(endpoint)) }
