# WebSocket Support Proposal for LiveFlux

## Overview
This document outlines a proposal for adding WebSocket support to LiveFlux, enabling real-time, bidirectional communication between the server and clients. This will allow for more interactive and responsive user experiences without full page reloads.

## Goals
1. Enable real-time server-to-client updates
2. Maintain backward compatibility with existing HTTP-based components
3. Provide a simple API for developers to work with WebSockets
4. Support automatic reconnection and state synchronization
5. Minimize boilerplate code for common WebSocket patterns

## Design

### 1. WebSocket Handler
Create a new `WebSocketHandler` that extends the existing `Handler` to manage WebSocket connections:

```go
type WebSocketHandler struct {
    *Handler
    upgrader websocket.Upgrader
    conns    map[string]*websocket.Conn // Component ID to WebSocket mapping
    mu       sync.RWMutex
}
```

### 2. Connection Lifecycle
- **Establishment**: Client initiates WebSocket connection after initial component load
- **Authentication**: Reuse existing session/auth mechanisms
- **Heartbeat**: Implement ping/pong for connection health monitoring
- **Cleanup**: Properly close and clean up connections on component unmount

### 3. Message Protocol
Define a simple JSON-based protocol for WebSocket messages:

```typescript
interface WSMessage {
    type: 'action' | 'update' | 'error' | 'sync';
    componentID: string;
    data: any;
    action?: string;  // For action messages
}
```

### 4. Component Integration
Add WebSocket support to components:

```go
type WebSocketComponent interface {
    Component
    // HandleWS handles WebSocket messages
    HandleWS(ctx context.Context, message []byte) ([]byte, error)
    // PushUpdate sends updates to the client
    PushUpdate(data interface{}) error
}
```

## Implementation Plan

### Phase 1: Core WebSocket Support
1. Add WebSocket dependency (`github.com/gorilla/websocket`)
2. Implement `WebSocketHandler` with basic connection management
3. Add WebSocket upgrade endpoint
4. Implement message routing between components and WebSocket connections

### Phase 2: Client-Side Integration
1. Create JavaScript client for WebSocket communication
2. Implement automatic reconnection logic
3. Add support for server-sent updates
4. Handle component initialization over WebSocket

### Phase 3: Advanced Features
1. Support for broadcasting to multiple clients
2. Subscription-based updates
3. Compression for WebSocket messages
4. Metrics and monitoring

## API Changes

### Server-Side
```go
// Enable WebSocket support on a route
mux.Handle("/ws", liveflux.NewWebSocketHandler(store))

// In component implementation
func (c *MyComponent) HandleWS(ctx context.Context, message []byte) ([]byte, error) {
    // Handle WebSocket message
    return json.Marshal(response)
}
```

### Client-Side
```javascript
// Initialize WebSocket connection
const ws = new LiveFluxWS('/ws', {
    onOpen: () => console.log('Connected'),
    onMessage: (msg) => handleUpdate(msg),
    onClose: () => console.log('Disconnected')
});

// Send action to server
ws.sendAction('component-id', 'increment', {value: 1});
```

## Security Considerations
- **Done** Implement CSRF protection for WebSocket upgrade (`WithWebSocketCSRFCheck`)
- Rate limiting for WebSocket connections
- **Done** Input validation for all WebSocket messages (`WithWebSocketMessageValidator`)
- **Done** Secure WebSocket (WSS) support (`WithWebSocketRequireTLS`)
- **Done** Origin validation (`WithWebSocketAllowedOrigins`)

## Performance Considerations
1. Connection pooling and management
2. Message batching for multiple updates
3. Efficient diffing for state updates
4. Compression for large payloads

## Backward Compatibility
- Existing HTTP-based components will continue to work without changes
- WebSocket support will be opt-in for components
- Fallback to HTTP for clients that don't support WebSockets

## Future Enhancements
1. Support for server-sent events (SSE) as a fallback
2. Offline support with optimistic updates
3. Message queuing for disconnected clients
4. Integration with existing state management solutions

## Testing Strategy
1. Unit tests for WebSocket handler and components
2. Integration tests for WebSocket communication
3. Load testing for WebSocket server
4. Browser compatibility testing

## Documentation
1. API reference for WebSocket support
2. Migration guide from HTTP to WebSocket
3. Example implementations
4. Best practices for real-time applications

## Timeline
- Phase 1: 2-3 weeks
- Phase 2: 2 weeks
- Phase 3: 2-3 weeks
- Testing and Documentation: 1-2 weeks

## Dependencies
- github.com/gorilla/websocket
- (Optional) Additional libraries for compression, message serialization

## Alternatives Considered
1. Server-Sent Events (SSE): Simpler but lacks bidirectional communication
2. Long polling: Higher latency and overhead
3. gRPC-Web: More complex setup and larger bundle size

## Conclusion
Adding WebSocket support to LiveFlux will significantly enhance its capabilities for building real-time, interactive web applications while maintaining the simplicity and developer experience that makes LiveFlux great.
