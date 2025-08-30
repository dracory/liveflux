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

// JS returns the standard Livewire client script used to mount and act on components.
// Include once per page (e.g., in your base layout) either via hb.Script(livewire.JS())
// or using livewire.Script().
func JS() string {
	return clientUtilJS + "\n" +
		clientNetworkJS + "\n" +
		clientMountJS + "\n" +
		clientHandlersJS + "\n" +
		clientBootstrapJS
}

// Script returns an hb.Script tag containing the standard client JS.
func Script() hb.TagInterface { return hb.Script(JS()) }
