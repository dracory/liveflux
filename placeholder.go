package liveflux

import (
	"fmt"

	"github.com/dracory/hb"
	"github.com/samber/lo"
)

// PlaceholderByKind returns a generic mount placeholder for a component by kind.
// The inline JS client should look for elements with data-flux-mount="1"
// and use the data-flux-component-kind value to POST an initial mount.
func PlaceholderByKind(kind string, params ...map[string]string) hb.TagInterface {
	p := lo.FirstOr(params, map[string]string{})

	label := fmt.Sprintf("Loading %s...", kind)
	div := hb.Div().
		Attr(DataFluxMount, "1").
		Attr(DataFluxComponentKind, kind).
		Text(label)

	for k, v := range p {
		if k == "" {
			continue
		}
		div = div.Attr(DataFluxParam+"-"+k, v)
	}

	return div
}

// PlaceholderFor uses a Component's kind to build a placeholder.
// Note: the instance ID is not available until after the first mount.
func Placeholder(c ComponentInterface, params ...map[string]string) hb.TagInterface {
	if c == nil {
		return hb.Text("component missing")
	}

	kind := KindOf(c)

	if kind == "" {
		return hb.Text("component has no kind")
	}

	return PlaceholderByKind(kind, params...)
}
