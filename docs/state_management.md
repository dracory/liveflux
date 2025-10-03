# State Management

State in Liveflux is stored server-side and keyed by component ID. The `Store` interface allows you to plug in different persistence backends depending on your deployment needs.

## Store Interface

Defined in `state.go`:

```go
type Store interface {
    Get(id string) (Component, bool)
    Set(c Component)
    Delete(id string)
}
```

- `Component` is an alias for `ComponentInterface`.
- `Set` should persist the component using its `GetID()` value.
- `Get` must return the same instance type that was stored to allow `Render` to run correctly.

## Default MemoryStore

`state.go` ships with `MemoryStore`, a `sync.RWMutex` guarded map. It is ideal for development or single-instance deployments:

```go
store := liveflux.NewMemoryStore()
mux.Handle("/liveflux", liveflux.NewHandler(store))
```

### Characteristics

- Process-local, reset on restart.
- Thread-safe for concurrent HTTP requests and WebSocket messages.
- Minimal configuration required.

## Custom Store Examples

### Session-based Store

Persist component state in user sessions (cookie store or server-side session cache). The example below wraps [`gorilla/sessions`](https://github.com/gorilla/sessions) and stores component state encoded with `encoding/gob`:

```go
type SessionStore struct {
    Sessions sessions.Store
    Name     string // session cookie name
}

// helper pulled from request context in Set/Get wrappers
type sessionCtx struct {
    Request  *http.Request
    Response http.ResponseWriter
}

func (s *SessionStore) Get(id string) (liveflux.Component, bool) {
    ctx := getSessionCtx() // application helper retrieving sessionCtx from context
    if ctx == nil {
        return nil, false
    }

    sess, err := s.Sessions.Get(ctx.Request, s.Name)
    if err != nil {
        return nil, false
    }

    raw, ok := sess.Values[id]
    if !ok {
        return nil, false
    }

    comp, ok := raw.(liveflux.Component)
    return comp, ok
}

func (s *SessionStore) Set(c liveflux.Component) {
    ctx := getSessionCtx()
    if ctx == nil {
        return
    }

    sess, err := s.Sessions.Get(ctx.Request, s.Name)
    if err != nil {
        return
    }

    sess.Values[c.GetID()] = c
    _ = sess.Save(ctx.Request, ctx.Response)
}

func (s *SessionStore) Delete(id string) {
    ctx := getSessionCtx()
    if ctx == nil {
        return
    }

    sess, err := s.Sessions.Get(ctx.Request, s.Name)
    if err != nil {
        return
    }

    delete(sess.Values, id)
    _ = sess.Save(ctx.Request, ctx.Response)
}
```

Because the `Store` interface does not receive `*http.Request`, wrap the Liveflux handler to attach request/response objects to the context consumed by `getSessionCtx()`:

```go
store := &SessionStore{Sessions: sessions.NewCookieStore([]byte("secret")), Name: "lfx"}
handler := liveflux.NewHandler(store)

mux.Handle("/liveflux", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx := context.WithValue(r.Context(), sessionContextKey{}, &sessionCtx{Request: r, Response: w})
    handler.ServeHTTP(w, r.WithContext(ctx))
}))
```

Register component concrete types with `gob.Register` during init so session serialization round-trips correctly.

### Distributed Cache (Redis)

Use Redis to share state between processes. Serialize component state with `encoding/gob` or JSON. The sample below stores a custom `Snapshot` payload per component:

```go
type snapshot struct {
    Alias string
    State map[string]any
}

type RedisStore struct {
    Client *redis.Client
    TTL    time.Duration
}

func (s *RedisStore) key(id string) string { return "liveflux:" + id }

func (s *RedisStore) Get(id string) (liveflux.Component, bool) {
    data, err := s.Client.Get(context.Background(), s.key(id)).Bytes()
    if err != nil {
        return nil, false
    }

    var snap snapshot
    if err := json.Unmarshal(data, &snap); err != nil {
        return nil, false
    }

    proto, err := liveflux.NewByAlias(snap.Alias)
    if err != nil {
        return nil, false
    }

    hydrator, ok := proto.(interface{ Hydrate(map[string]any) error })
    if !ok {
        return nil, false
    }

    if err := hydrator.Hydrate(snap.State); err != nil {
        return nil, false
    }
    return proto, true
}

func (s *RedisStore) Set(c liveflux.Component) {
    snapshotter, ok := c.(interface{ Snapshot() (map[string]any, error) })
    if !ok {
        return
    }

    state, err := snapshotter.Snapshot()
    if err != nil {
        return
    }

    data, err := json.Marshal(snapshot{Alias: c.GetAlias(), State: state})
    if err != nil {
        return
    }

    ttl := s.TTL
    if ttl <= 0 {
        ttl = time.Hour
    }
    _ = s.Client.Set(context.Background(), s.key(c.GetID()), data, ttl).Err()
}

func (s *RedisStore) Delete(id string) {
    _ = s.Client.Del(context.Background(), s.key(id)).Err()
}
```

This approach pushes serialization responsibility to components through `Snapshot()`/`Hydrate()`. Alternatively, maintain a registry of custom marshaling functions per alias. Apply expirations to avoid orphaned keys and call `Delete` when user sessions end.

## Lifecycle Considerations

- **Mount**: After `Mount`, the handler immediately stores the component. Ensure `Mount` populates all required state before returning.
- **Handle**: The handler re-stores the component after `Handle`. If your store performs partial updates, make sure state mutations persist atomically.
- **Cleanup**: Call `Store.Delete` when a component is no longer needed (e.g., after a redirect to a new page). The default handler does not currently auto-delete; implement cleanup in a custom handler wrapper if required.

## Concurrency

Components should be safe for concurrent access when using WebSockets or multiple HTTP requests. Guard mutable collections or counters inside your component if actions may overlap. The default store ensures serialized access per request, but WebSocket messages processed concurrently might need additional synchronization if they mutate shared state.

## Testing Stores

Use `component_test.go` and `state_test.go` as references. For custom stores, write tests that mimic handler interactions:

1. Mount a component; ensure `Set` persists state.
2. Retrieve via `Get` and verify the component mutates correctly after `Handle`.
3. Delete an ID and confirm subsequent `Get` fails.
