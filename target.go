package liveflux

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/dracory/hb"
)

// TargetFragment represents a fragment of HTML to be applied to a specific DOM selector.
type TargetFragment struct {
	// Selector is the CSS selector identifying the target element
	Selector string
	// Content is the HTML content to apply
	Content hb.TagInterface
	// SwapMode determines how the content is merged with the target
	// Valid values: "replace", "inner", "beforebegin", "afterbegin", "beforeend", "afterend"
	SwapMode string
}

// TargetRenderer is an optional interface that components can implement
// to support targeted fragment updates instead of full component re-renders.
type TargetRenderer interface {
	// RenderTargets returns a list of fragments to update specific DOM regions.
	// The handler will check for this interface and, if the client supports targets,
	// will send only these fragments instead of the full component render.
	RenderTargets(ctx context.Context) []TargetFragment
}

// BuildTargetResponse constructs an HTML response containing template elements
// for each fragment, plus an optional full component render as fallback.
func BuildTargetResponse(fragments []TargetFragment, fullRender string, comp ComponentInterface) string {
	var sb strings.Builder

	// Write each fragment as a <template> element
	for _, frag := range fragments {
		swapMode := frag.SwapMode
		if swapMode == "" {
			swapMode = "replace"
		}

		// Escape attributes
		selector := html.EscapeString(frag.Selector)
		swap := html.EscapeString(swapMode)
		componentAlias := html.EscapeString(comp.GetAlias())
		componentID := html.EscapeString(comp.GetID())

		// Render content
		content := ""
		if frag.Content != nil {
			content = frag.Content.ToHTML()
		}

		sb.WriteString(fmt.Sprintf(
			`<template data-flux-target="%s" data-flux-swap="%s" data-flux-component="%s" data-flux-component-id="%s">%s</template>`,
			selector,
			swap,
			componentAlias,
			componentID,
			content,
		))
	}

	// Include full render as fallback
	if fullRender != "" {
		componentAlias := html.EscapeString(comp.GetAlias())
		componentID := html.EscapeString(comp.GetID())
		sb.WriteString(fmt.Sprintf(
			`<template data-flux-component="%s" data-flux-component-id="%s">%s</template>`,
			componentAlias,
			componentID,
			fullRender,
		))
	}

	return sb.String()
}

// TargetSelector is a helper to build common CSS selectors.
// Examples:
//   - TargetSelector("#cart-total") -> "#cart-total"
//   - TargetSelector(".line-items") -> ".line-items"
//   - TargetSelector("[data-id='42']") -> "[data-id='42']"
func TargetSelector(selector string) string {
	return selector
}

// TargetID builds a selector for an element with a specific ID.
func TargetID(id string) string {
	return "#" + id
}

// TargetClass builds a selector for elements with a specific class.
func TargetClass(class string) string {
	return "." + class
}

// TargetAttr builds a selector for elements with a specific attribute value.
func TargetAttr(attr, value string) string {
	return fmt.Sprintf("[%s='%s']", attr, value)
}

// SwapMode constants for common swap operations
const (
	SwapReplace     = "replace"     // Replace the target element entirely
	SwapInner       = "inner"       // Replace the target's innerHTML
	SwapBeforeBegin = "beforebegin" // Insert before the target element
	SwapAfterBegin  = "afterbegin"  // Insert as first child of target
	SwapBeforeEnd   = "beforeend"   // Insert as last child of target (append)
	SwapAfterEnd    = "afterend"    // Insert after the target element
)
