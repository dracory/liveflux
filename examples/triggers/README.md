# Liveflux Triggers Example

This example demonstrates the `data-flux-trigger` feature for declarative event handling.

## Features Demonstrated

- **Live Search**: Search-as-you-type functionality with automatic debouncing
- **Event Modifiers**: Using `keyup changed delay:300ms` to control trigger behavior
- **No Custom JavaScript**: Everything is declarative using HTML attributes

## Running the Example

```bash
cd examples/triggers
go run .
```

Then open http://localhost:8082 in your browser.

## How It Works

The search input uses the `data-flux-trigger` attribute:

```html
<input 
  type="text" 
  name="query"
  data-flux-trigger="keyup changed delay:300ms"
  data-flux-action="search" />
```

### Trigger Breakdown

- **`keyup`** - Event to listen for (fires on every keystroke)
- **`changed`** - Filter that only triggers if the value actually changed
- **`delay:300ms`** - Debounce modifier that waits 300ms after the last keystroke

### Targeted Updates

The component implements `TargetRenderer` to use targeted updates:

```go
func (c *SearchComponent) RenderTargets(ctx context.Context) []liveflux.TargetFragment {
    if c.IsDirty("results") {
        return []liveflux.TargetFragment{
            {
                Selector: "#search-results",
                Content:  c.renderResults(),
            },
        }
    }
    return nil
}
```

This ensures only the results section is updated, **preserving input focus** during typing.

### Benefits

1. **Automatic Debouncing**: No need to write custom JavaScript timers
2. **Change Detection**: Built-in value caching prevents redundant requests
3. **Declarative**: Everything is configured via HTML attributes
4. **Progressive Enhancement**: Works with standard Liveflux components
5. **Maintains Focus**: Targeted updates keep the input focused while typing

## Code Structure

- **`search.go`** - SearchComponent with live search functionality
- **`main.go`** - HTTP server setup and page rendering

## Try It

1. Start typing in the search box
2. Notice the 300ms delay before the search executes
3. Try typing the same text again - the `changed` filter prevents duplicate requests
4. Search is performed server-side with full component state management
