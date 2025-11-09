package main

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// DealsComponent showcases a second component using targeted updates.
type DealsComponent struct {
	liveflux.Base
	Deals []string
	Index int
}

func (d *DealsComponent) GetKind() string {
	return "deals"
}

func (d *DealsComponent) Mount(ctx context.Context, params map[string]string) error {
	d.Deals = []string{
		"Free shipping over $100",
		"Buy one, get one 20% off",
		"Members save extra 10%",
	}
	d.Index = 0
	return nil
}

func (d *DealsComponent) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "next-deal":
		if len(d.Deals) == 0 {
			return nil
		}
		d.Index = (d.Index + 1) % len(d.Deals)
		d.MarkTargetDirty("#current-deal")
	case "shuffle":
		if len(d.Deals) == 0 {
			return nil
		}
		first := d.Deals[0]
		d.Deals = append(d.Deals[1:], first)
		d.Index = 0
		d.MarkTargetDirty("#current-deal")
	}
	return nil
}

func (d *DealsComponent) Render(ctx context.Context) hb.TagInterface {
	return d.Root(
		hb.Div().Class("deals-container").Children([]hb.TagInterface{
			hb.H2().Text("Daily Deals"),
			d.renderCurrentDeal(),
			d.renderActions(),
		}),
	)
}

func (d *DealsComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
	if !d.IsDirty("#current-deal") {
		return nil
	}
	return []liveflux.TargetFragment{
		{
			Selector: "#current-deal",
			Content:  d.renderCurrentDeal(),
			SwapMode: liveflux.SwapReplace,
		},
	}
}

func (d *DealsComponent) renderCurrentDeal() hb.TagInterface {
	message := "No deals available"
	if len(d.Deals) > 0 {
		message = d.Deals[d.Index]
	}
	return hb.Div().
		ID("current-deal").
		Class("deal-banner").
		Text(message)
}

func (d *DealsComponent) renderActions() hb.TagInterface {
	return hb.Div().Class("deals-actions").Children([]hb.TagInterface{
		hb.Button().
			Attr(liveflux.DataFluxAction, "next-deal").
			Class("btn").
			Style("margin-right: 10px;").
			Text("Next Deal"),
		hb.Button().
			Attr(liveflux.DataFluxAction, "shuffle").
			Class("btn").
			Text("Shuffle Deals"),
	})
}
