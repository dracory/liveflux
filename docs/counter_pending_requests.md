# Investigation: Counter Pending Requests

_Date_: 2025-11-04

## Context
Rapidly clicking the increment button in the Counter example caused multiple `fetch` calls to appear stuck as "pending" in the browser.

## Findings
1. **Server request flow** – The default Go `http.Server` processes requests sequentially per connection. While the Liveflux handler is still mutating state and writing a response, follow-up POSTs queue at the TCP level, so each request waits for the previous one to finish, appearing as pending to the client. The handler path (`handle`) performs the mutation and storage synchronously without spawning a goroutine. @handler.go#160-194
2. **Client behaviour** – The browser runtime fires a `fetch` for every click instantly and has no throttling or cancellation. Each call serializes form data and posts immediately, so rapid clicks pile up while the DOM is still being replaced with the previous response. @js/liveflux_handlers.js#15-92
3. **State storage** – Components are persisted in an in-memory store protected by an RWMutex, so requests are serialized safely today. If the HTTP handler is adapted to allow concurrent processing, per-component locking or another concurrency strategy will be necessary to avoid state races. @state.go#12-48 @handler.go#178-185

## Recommendations
1. ~~Switch to a request model that allows concurrency (e.g., wrap handler work in goroutines or use a server implementation that spawns them) if lower latency is desired. Ensure component state updates remain synchronized.~~ **NOT NEEDED** - Go's `http.Server` already spawns goroutines per connection
2. ~~Add client-side throttling/debouncing or abort in-flight requests before dispatching another action to improve perceived responsiveness.~~ **IMPLEMENTED** @js/liveflux_handlers.js#15-92
3. ~~If enabling concurrent actions, introduce per-component locking or optimistic concurrency controls so simultaneous requests do not corrupt component state.~~ **IMPLEMENTED** @handler.go#168-174 @state.go#52-67

## Implementation

### Client-Side Request Throttling (2025-11-04)
Implemented per-component request throttling in `liveflux_handlers.js` to prevent multiple simultaneous requests:

- **Tracking mechanism**: Added a `Map` to track in-flight requests by component ID
- **Request gating**: Before dispatching a new action, check if a request is already pending for that component
- **Automatic cleanup**: Use `.finally()` to ensure the pending flag is cleared regardless of success or failure
- **User feedback**: Log skipped actions to console for debugging

**Benefits**:
- Eliminates "pending" request buildup from rapid clicks
- Prevents potential race conditions on component state
- Maintains request ordering (first click wins until completion)
- No server-side changes required

**Code location**: @js/liveflux_handlers.js#15-75

### Server-Side Per-Component Locking (2025-11-04)
Implemented per-component mutex locking in the handler to prevent race conditions when multiple requests target the same component:

**Problem**: Go's `http.Server` spawns goroutines per connection, so concurrent requests for the same component ID could:
1. Both retrieve the same component instance from the store
2. Both modify it simultaneously (race condition)
3. Both save it back, potentially losing updates
4. Cause deadlocks or stuck requests

**Solution**: Added per-component locking using `sync.Map` to track mutexes by component ID:

```go
// In MemoryStore
locks sync.Map // map[string]*sync.Mutex

func (s *MemoryStore) LockComponent(id string) *sync.Mutex {
    actual, _ := s.locks.LoadOrStore(id, &sync.Mutex{})
    mu := actual.(*sync.Mutex)
    mu.Lock()
    return mu
}
```

```go
// In Handler.handle()
if memStore, ok := h.Store.(*MemoryStore); ok {
    componentLock := memStore.LockComponent(id)
    defer memStore.UnlockComponent(componentLock)
}
```

**Benefits**:
- Prevents concurrent modifications of the same component
- Serializes requests per component (different components can still process concurrently)
- Eliminates race conditions and stuck requests
- Automatic cleanup via `defer`

**Code locations**: 
- @state.go#18 (locks field)
- @state.go#52-67 (LockComponent/UnlockComponent methods)
- @handler.go#168-174 (lock acquisition in handle)

**Testing**: Open multiple browser tabs, rapidly click buttons in both tabs simultaneously. Requests should process sequentially per component without hanging or corrupting state.
