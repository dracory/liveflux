package liveflux

import (
	"fmt"

	"github.com/gouniverse/hb"
	"github.com/samber/lo"
)

// PlaceholderByAlias returns a generic mount placeholder for a component by alias.
// The inline JS client should look for elements with data-flux-mount="1"
// and use the data-flux-component value to POST an initial mount.
func PlaceholderByAlias(alias string, params ...map[string]string) hb.TagInterface {
	p := lo.FirstOr(params, map[string]string{})

	label := fmt.Sprintf("Loading %s...", alias)
	div := hb.Div().
		Attr("data-flux-mount", "1").
		Attr("data-flux-component", alias).
		Text(label)

	for k, v := range p {
		if k == "" {
			continue
		}
		div = div.Attr("data-flux-param-"+k, v)
	}

	return div
}

// PlaceholderFor uses a Component's alias to build a placeholder.
// Note: the instance ID is not available until after the first mount.
func Placeholder(c ComponentInterface, params ...map[string]string) hb.TagInterface {
	if c == nil {
		return hb.Text("component missing")
	}

	alias := AliasOf(c)

	if alias == "" {
		return hb.Text("component has no alias")
	}

	return PlaceholderByAlias(alias, params...)
}
