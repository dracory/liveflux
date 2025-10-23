# Events

Liveflux provides a robust event system inspired by Laravel Livewire that enables communication between components, JavaScript code, and even Alpine.js. Events use browser custom events under the hood, making them compatible with any JavaScript framework.

## Overview

The event system allows you to:

- Dispatch events from Go components to notify other components
- Listen for events in Go using method naming conventions
- Dispatch and listen for events from JavaScript
- Integrate with Alpine.js and vanilla JavaScript
- Target specific components or broadcast globally

## Dispatching Events

### From Go Components

Use the `Dispatch` method from any component that embeds `liveflux.Base`:

```go
type PostCreator struct {
    liveflux.Base
    Title string
}

func (pc *PostCreator) Handle(ctx context.Context, action string, data url.Values) error {
    if action == "create" {
        // Dispatch event with data
        pc.Dispatch("post-created", map[string]interface{}{
            "title": pc.Title,
            "id":    123,
        })
    }
    return nil
}
```

### Dispatch Methods

**`Dispatch(name, data)`** - Broadcast to all listeners:
```go
pc.Dispatch("post-created", map[string]interface{}{"title": "Hello"})
```

**`DispatchTo(componentAlias, name, data)`** - Send to specific component type:
```go
pc.DispatchTo("dashboard", "refresh", map[string]interface{}{"count": 5})
```

**`DispatchSelf(name, data)`** - Send only to the current component:
```go
pc.DispatchSelf("internal-update", map[string]interface{}{"status": "done"})
```

## Listening for Events

### In Go Components

Use method naming convention: `On{EventName}` automatically listens to `event-name`:

```go
type PostList struct {
    liveflux.Base
    Posts []Post
}

func (pl *PostList) Mount(ctx context.Context, params map[string]string) error {
    // Register event listeners based on method names
    liveflux.RegisterEventListeners(pl, pl.GetEventDispatcher())
    return nil
}

// OnPostCreated listens for "post-created" event
func (pl *PostList) OnPostCreated(ctx context.Context, event liveflux.Event) error {
    title, _ := event.Data["title"].(string)
    pl.Posts = append(pl.Posts, Post{Title: title})
    return nil
}

// OnUserUpdated listens for "user-updated" event
func (pl *PostList) OnUserUpdated(ctx context.Context, event liveflux.Event) error {
    // Handle user update
    return nil
}
```

**Method Signature:**
```go
func (c *Component) On{EventName}(ctx context.Context, event liveflux.Event) error
```

**Event Structure:**
```go
type Event struct {
    Name string                 // Event name
    Data map[string]interface{} // Event payload
}
```

### Dynamic Event Names

For dynamic event names (e.g., scoped to a model ID):

```go
// Dispatch with dynamic name
pc.Dispatch(fmt.Sprintf("post-updated.%d", post.ID), map[string]interface{}{
    "title": post.Title,
})
```

You can use manual registration for dynamic listeners:

```go
func (pl *PostList) Mount(ctx context.Context, params map[string]string) error {
    dispatcher := pl.GetEventDispatcher()
    
    // Register dynamic listener
    eventName := fmt.Sprintf("post-updated.%d", pl.PostID)
    dispatcher.On(eventName, func(ctx context.Context, event liveflux.Event) error {
        // Handle event
        return nil
    })
    
    return nil
}
```

## JavaScript Integration

### Listening in Component Scripts

Use `$wire.on()` within component scripts:

```go
func (c *Component) Render(ctx context.Context) hb.TagInterface {
    script := hb.Script(`
        (function(){
            var root = document.currentScript.closest('[data-flux-root]');
            if(!root || !root.$wire) return;
            
            root.$wire.on('post-created', function(event){
                console.log('Post created:', event.data.title);
                // Update UI or trigger actions
            });
        })();
    `)
    
    return c.Root(hb.Div().Child(script))
}
```

### Dispatching from JavaScript

**From component script:**
```javascript
root.$wire.dispatch('post-created', { title: 'My Post' });
root.$wire.dispatchSelf('internal-event');
root.$wire.dispatchTo('dashboard', 'refresh');
```

**Global dispatch:**
```javascript
Liveflux.dispatch('post-created', { title: 'My Post' });
```

### Global Listeners

Listen for events from any component:

```javascript
document.addEventListener('livewire:init', function(){
    Liveflux.on('post-created', function(event){
        console.log('Post created:', event.data);
    });
});
```

Remove listeners with the cleanup function:

```javascript
var cleanup = Liveflux.on('post-created', function(event){
    console.log('Post created');
});

// Later, remove the listener
cleanup();
```

## Alpine.js Integration

### Listening to Events

Use Alpine's `x-on` directive:

```html
<!-- Listen on element -->
<div x-on:post-created="console.log('Post created!', $event.detail)">
```

```html
<!-- Listen globally with .window modifier -->
<div x-on:post-created.window="handlePostCreated($event.detail)">
```

Access event data via `$event.detail`:

```html
<div x-data="{ message: '' }" 
     x-on:post-created.window="message = 'New post: ' + $event.detail.title">
    <span x-text="message"></span>
</div>
```

### Dispatching from Alpine

Use Alpine's `$dispatch`:

```html
<button @click="$dispatch('post-created', { title: 'My Post' })">
    Create Post
</button>
```

Liveflux components can listen for these events using the `On{EventName}` method.

## Event Flow

### Server to Client

1. Component calls `Dispatch()` during `Handle()` or `Mount()`
2. Events are queued in the component's `EventDispatcher`
3. Handler sends events via `X-Liveflux-Events` header
4. Client processes header and dispatches browser events
5. Listeners receive events (Go listeners on next request, JS listeners immediately)

### Client to Server

JavaScript events are client-side only. To trigger server-side actions:

```javascript
root.$wire.on('post-created', function(event){
    // Trigger a server action
    var form = root.querySelector('form');
    var submitEvent = new Event('submit', {bubbles: true, cancelable: true});
    form.dispatchEvent(submitEvent);
});
```

## Advanced Patterns

### Parent-Child Communication

Instead of events, you can use `$parent` for direct parent calls:

```html
<button wire:click="$parent.showCreateForm()">Create</button>
```

However, events are better for:
- Loose coupling between components
- Broadcasting to multiple listeners
- Integration with JavaScript/Alpine

### Listening on Child Components

Listen for events from specific child components in Blade templates:

```html
<div>
    <!-- Listen for 'saved' event from this child -->
    <div data-flux-mount="1" 
         data-flux-component="edit-post"
         x-on:saved="console.log('Post saved!')">
    </div>
</div>
```

### Event Scoping

**Global events** - All components receive:
```go
pc.Dispatch("post-created", data)
```

**Targeted events** - Only specific component type receives:
```go
pc.DispatchTo("dashboard", "refresh", data)
```

**Self events** - Only the dispatching component receives:
```go
pc.DispatchSelf("internal-update", data)
```

## Best Practices

1. **Use descriptive event names**: `post-created`, `user-updated`, not `event1`, `update`
2. **Keep event data simple**: Use JSON-serializable types (strings, numbers, booleans, maps, slices)
3. **Register listeners in Mount**: Call `RegisterEventListeners()` in your `Mount()` method
4. **Handle errors gracefully**: Event listeners should return errors for proper logging
5. **Avoid circular events**: Don't dispatch events from within event handlers that could create loops
6. **Use events for loose coupling**: Prefer events over direct component references
7. **Document your events**: Comment which events your component dispatches and listens to

## Comparison with Livewire

| Livewire | Liveflux | Notes |
|----------|----------|-------|
| `$this->dispatch('event')` | `component.Dispatch("event", data)` | Dispatch event |
| `#[On('event')]` | `OnEventName()` method | Listen to event |
| `$wire.on('event', fn)` | `$wire.on('event', fn)` | JS listener |
| `Livewire.on('event', fn)` | `Liveflux.on('event', fn)` | Global JS listener |
| `dispatch()->to(Component::class)` | `DispatchTo("alias", "event", data)` | Target specific component |
| `dispatchSelf('event')` | `DispatchSelf("event", data)` | Self-only event |
| `x-on:event` | `x-on:event` | Alpine integration |
| `$dispatch('event')` | `$dispatch('event')` | Alpine dispatch |

## Examples

See the full working example in `examples/events/` demonstrating:
- Event dispatching from a form submission
- Automatic event listener registration
- Multiple components listening to the same event
- JavaScript event handling
- Real-time UI updates via events

Run the example:
```bash
go run ./examples/events
```

## API Reference

### Go API

**Component Methods:**
- `Dispatch(name string, data ...map[string]interface{})` - Dispatch event globally
- `DispatchTo(alias, name string, data ...map[string]interface{})` - Dispatch to specific component
- `DispatchSelf(name string, data ...map[string]interface{})` - Dispatch to self only
- `GetEventDispatcher() *EventDispatcher` - Get component's event dispatcher

**Functions:**
- `RegisterEventListeners(component ComponentInterface, dispatcher *EventDispatcher)` - Auto-register listeners based on method names
- `ParseOnAttributes(component ComponentInterface) []OnAttribute` - Extract event listener methods

**Types:**
- `Event` - Event structure with Name and Data
- `EventDispatcher` - Manages event queue and listeners
- `EventListener` - Function type for event handlers

### JavaScript API

**$wire Object (component-scoped):**
- `$wire.on(eventName, callback)` - Listen for event
- `$wire.dispatch(eventName, data)` - Dispatch event
- `$wire.dispatchSelf(eventName, data)` - Dispatch self-only event
- `$wire.dispatchTo(alias, eventName, data)` - Dispatch to specific component

**Liveflux Object (global):**
- `Liveflux.on(eventName, callback)` - Listen globally
- `Liveflux.dispatch(eventName, data)` - Dispatch globally

**Browser Events:**
- `livewire:init` - Fired when Liveflux initializes
- Custom events dispatched with event name (e.g., `post-created`)
