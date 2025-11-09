package liveflux

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/hb"
)

// TestComponent implements TargetRenderer for testing
type TestTargetComponent struct {
	Base
	Value string
}

func (c *TestTargetComponent) GetKind() string {
	return "test-target"
}

func (c *TestTargetComponent) Mount(ctx context.Context, params map[string]string) error {
	c.Value = "initial"
	return nil
}

func (c *TestTargetComponent) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "update" {
		c.Value = "updated"
		c.MarkTargetDirty("#value")
	}
	return nil
}

func (c *TestTargetComponent) Render(ctx context.Context) hb.TagInterface {
	return c.Root(
		hb.Div().ID("value").Text(c.Value),
	)
}

func (c *TestTargetComponent) RenderTargets(ctx context.Context) []TargetFragment {
	fragments := []TargetFragment{}

	if c.IsDirty("#value") {
		fragments = append(fragments, TargetFragment{
			Selector: "#value",
			Content:  hb.Div().ID("value").Text(c.Value),
			SwapMode: SwapReplace,
		})
	}

	return fragments
}

func TestTargetFragment(t *testing.T) {
	t.Run("basic fragment creation", func(t *testing.T) {
		frag := TargetFragment{
			Selector: "#test",
			Content:  hb.Div().Text("content"),
			SwapMode: SwapReplace,
		}

		if frag.Selector != "#test" {
			t.Errorf("expected selector #test, got %s", frag.Selector)
		}
		if frag.SwapMode != SwapReplace {
			t.Errorf("expected swap mode replace, got %s", frag.SwapMode)
		}
	})
}

func TestBuildTargetResponse(t *testing.T) {
	comp := &TestTargetComponent{}
	comp.SetKind("test-target")
	comp.SetID("abc123")

	fragments := []TargetFragment{
		{
			Selector: "#value",
			Content:  hb.Div().ID("value").Text("updated"),
			SwapMode: SwapReplace,
		},
	}

	fullRender := `<div data-flux-root="1"><div id="value">updated</div></div>`

	response := BuildTargetResponse(fragments, fullRender, comp)

	// Check that response contains template elements
	if !strings.Contains(response, "<template") {
		t.Error("response should contain template elements")
	}

	// Check selector
	if !strings.Contains(response, `data-flux-target="#value"`) {
		t.Error("response should contain target selector")
	}

	// Check swap mode
	if !strings.Contains(response, `data-flux-swap="replace"`) {
		t.Error("response should contain swap mode")
	}

	// Check component metadata
	if !strings.Contains(response, `data-flux-component-kind="test-target"`) {
		t.Error("response should contain component kind")
	}
	if !strings.Contains(response, `data-flux-component-id="abc123"`) {
		t.Error("response should contain component ID")
	}

	// Check full render fallback
	if !strings.Contains(response, fullRender) {
		t.Error("response should contain full render fallback")
	}
}

func TestBuildTargetResponseWithMultipleFragments(t *testing.T) {
	comp := &TestTargetComponent{}
	comp.SetKind("test")
	comp.SetID("123")

	fragments := []TargetFragment{
		{
			Selector: "#first",
			Content:  hb.Div().Text("first"),
			SwapMode: SwapReplace,
		},
		{
			Selector: "#second",
			Content:  hb.Div().Text("second"),
			SwapMode: SwapInner,
		},
	}

	response := BuildTargetResponse(fragments, "", comp)

	// Should contain both fragments
	if !strings.Contains(response, `data-flux-target="#first"`) {
		t.Error("response should contain first fragment")
	}
	if !strings.Contains(response, `data-flux-target="#second"`) {
		t.Error("response should contain second fragment")
	}

	// Should have different swap modes
	firstIdx := strings.Index(response, `data-flux-target="#first"`)
	secondIdx := strings.Index(response, `data-flux-target="#second"`)

	if firstIdx == -1 || secondIdx == -1 {
		t.Fatal("fragments not found in response")
	}

	// Check swap modes appear in correct positions
	firstPart := response[:secondIdx]
	if !strings.Contains(firstPart, `data-flux-swap="replace"`) {
		t.Error("first fragment should have replace swap mode")
	}

	secondPart := response[secondIdx:]
	if !strings.Contains(secondPart, `data-flux-swap="inner"`) {
		t.Error("second fragment should have inner swap mode")
	}
}

func TestTargetHelpers(t *testing.T) {
	t.Run("TargetID", func(t *testing.T) {
		result := TargetID("my-element")
		if result != "#my-element" {
			t.Errorf("expected #my-element, got %s", result)
		}
	})

	t.Run("TargetClass", func(t *testing.T) {
		result := TargetClass("my-class")
		if result != ".my-class" {
			t.Errorf("expected .my-class, got %s", result)
		}
	})

	t.Run("TargetAttr", func(t *testing.T) {
		result := TargetAttr("data-id", "42")
		if result != "[data-id='42']" {
			t.Errorf("expected [data-id='42'], got %s", result)
		}
	})

	t.Run("TargetSelector", func(t *testing.T) {
		result := TargetSelector("#custom")
		if result != "#custom" {
			t.Errorf("expected #custom, got %s", result)
		}
	})
}

func TestBaseDirtyTracking(t *testing.T) {
	base := &Base{}

	t.Run("MarkTargetDirty", func(t *testing.T) {
		base.MarkTargetDirty("#test")

		if !base.IsDirty("#test") {
			t.Error("target should be marked dirty")
		}
	})

	t.Run("IsDirty returns false for unmarked targets", func(t *testing.T) {
		if base.IsDirty("#other") {
			t.Error("unmarked target should not be dirty")
		}
	})

	t.Run("GetDirtyTargets", func(t *testing.T) {
		base.MarkTargetDirty("#first")
		base.MarkTargetDirty("#second")

		dirty := base.GetDirtyTargets()

		if len(dirty) != 3 { // #test, #first, #second
			t.Errorf("expected 3 dirty targets, got %d", len(dirty))
		}

		if !dirty["#first"] || !dirty["#second"] || !dirty["#test"] {
			t.Error("all marked targets should be in dirty map")
		}
	})

	t.Run("ClearDirtyTargets", func(t *testing.T) {
		base.ClearDirtyTargets()

		if base.IsDirty("#test") {
			t.Error("targets should be cleared")
		}

		dirty := base.GetDirtyTargets()
		if len(dirty) != 0 {
			t.Errorf("expected 0 dirty targets after clear, got %d", len(dirty))
		}
	})

	t.Run("GetDirtyTargets returns copy", func(t *testing.T) {
		base.MarkTargetDirty("#test")
		dirty := base.GetDirtyTargets()

		// Modify the copy
		dirty["#new"] = true

		// Original should not be affected
		if base.IsDirty("#new") {
			t.Error("modifying copy should not affect original")
		}
	})
}

func TestTargetRendererInterface(t *testing.T) {
	comp := &TestTargetComponent{}
	comp.SetKind("test-target")
	comp.SetID("test-id")

	// Mount component
	ctx := context.Background()
	if err := comp.Mount(ctx, nil); err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	// Initially no dirty targets
	fragments := comp.RenderTargets(ctx)
	if len(fragments) != 0 {
		t.Error("should have no fragments when nothing is dirty")
	}

	// Handle action that marks target dirty
	if err := comp.Handle(ctx, "update", url.Values{}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	// Now should have fragments
	fragments = comp.RenderTargets(ctx)
	if len(fragments) != 1 {
		t.Errorf("expected 1 fragment, got %d", len(fragments))
	}

	if fragments[0].Selector != "#value" {
		t.Errorf("expected selector #value, got %s", fragments[0].Selector)
	}

	if fragments[0].SwapMode != SwapReplace {
		t.Errorf("expected swap mode replace, got %s", fragments[0].SwapMode)
	}

	// Verify content
	html := fragments[0].Content.ToHTML()
	if !strings.Contains(html, "updated") {
		t.Error("fragment content should contain updated value")
	}
}

func TestSwapModeConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{SwapReplace, "replace"},
		{SwapInner, "inner"},
		{SwapBeforeBegin, "beforebegin"},
		{SwapAfterBegin, "afterbegin"},
		{SwapBeforeEnd, "beforeend"},
		{SwapAfterEnd, "afterend"},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("constant mismatch: expected %s, got %s", tt.expected, tt.constant)
		}
	}
}
