# Events Example

This example demonstrates Liveflux's event system, inspired by Laravel Livewire's event architecture.

## Features Demonstrated

- **Event Dispatching**: The `PostCreator` component dispatches a `post-created` event when a post is created
- **Event Listening**: The `PostList` component listens for `post-created` events using the `OnPostCreated` method
- **Automatic Registration**: Event listeners are automatically registered based on method naming convention (`On{EventName}`)
- **Event Data**: Events can carry data payloads (title, timestamp)
- **Multiple Listeners**: Multiple components can listen to the same event
- **JavaScript Integration**: Components can listen to events from JavaScript using `$wire.on()`

## How It Works

### Server-Side Events

1. **Dispatching Events**:
   ```go
   pc.Dispatch("post-created", map[string]interface{}{
       "title":     pc.Title,
       "timestamp": time.Now().Format("15:04:05"),
   })
   ```

2. **Listening to Events**:
   ```go
   // Method naming convention: On{EventName} -> listens to "event-name"
   func (pl *PostList) OnPostCreated(ctx context.Context, event liveflux.Event) error {
       title, _ := event.Data["title"].(string)
       pl.Posts = append(pl.Posts, Post{Title: title})
       return nil
   }
   ```

3. **Registering Listeners**:
   ```go
   liveflux.RegisterEventListeners(pl, pl.GetEventDispatcher())
   ```

### Client-Side Events

Components can listen to events from JavaScript:

```javascript
root.$wire.on('post-created', function(event){
    console.log('Event received:', event.data);
});
```

### Event Flow

1. User creates a post in `PostCreator`
2. `PostCreator.Handle()` dispatches `post-created` event
3. Event is sent to client via `X-Liveflux-Events` header
4. Client processes event and dispatches it as a browser event
5. `PostList` receives the event (if it has registered a listener)
6. Other components can also listen via JavaScript

## Running the Example

From the repository root:

```bash
go run ./examples/events
```

Or with Task:

```bash
task examples:events:run
```

Then open http://localhost:8084 in your browser.

## Event System Features

### Dispatch Methods

- `Dispatch(name, data)` - Dispatch to all listeners
- `DispatchTo(componentKind, name, data)` - Dispatch to specific component type
- `DispatchSelf(name, data)` - Dispatch only to current component

### Listening Methods

- **Go**: Use `On{EventName}` method naming convention
- **JavaScript**: Use `$wire.on(eventName, callback)` in component scripts
- **Global**: Use `liveflux.on(eventName, callback)` or `document.addEventListener(eventName, callback)`

### Alpine.js Integration

Events work seamlessly with Alpine.js:

```html
<div x-on:post-created.window="console.log('Post created!', $event.detail)">
```

## Comparison with Livewire

This implementation closely mirrors Livewire's event system:

- ✅ `$this->dispatch()` → `component.Dispatch()`
- ✅ `#[On('event')]` → `OnEventName()` method
- ✅ `$wire.on()` → `$wire.on()`
- ✅ `Livewire.on()` → `liveflux.on()`
- ✅ `dispatch()->to()` → `DispatchTo()`
- ✅ `dispatchSelf()` → `DispatchSelf()`
- ✅ Alpine integration via browser events
