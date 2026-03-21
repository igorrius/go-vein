## Why

Go lacks a built-in, type-safe message bus. Existing solutions rely on `any` or `reflect`, introducing runtime overhead and losing compile-time safety. A generic, allocation-minimal event bus enables decoupled, high-throughput communication between components with full type safety.

## What Changes

- Introduce a new `vein` package: a lightweight, generic message bus for Go
- Typed event subscriptions via generics — no `any` or reflection
- Global default dispatcher for ergonomic usage without explicit initialization
- `Subscription[T]` with `On(func(T))`, `OnC() <-chan T`, and `Unsubscribe()` methods
- Non-blocking channel dispatch (drop or buffered, configurable)
- Concurrent handler execution: each `On(func)` subscriber runs in its own goroutine per event
- Simple publish API: `Publish(event)`
- Minimal allocations on hot path via sync.Pool and pre-allocated structures

## Capabilities

### New Capabilities

- `message-bus`: Core dispatcher — type-safe event publish/subscribe engine with generic support
- `subscription`: Subscription lifecycle — `On`, `OnC`, `Unsubscribe`, concurrent/channel delivery modes
- `global-dispatcher`: Package-level default dispatcher with `Subscribe[T]()` and `Publish[T]()` helpers

### Modified Capabilities

## Impact

- New top-level Go package (module: `github.com/igorrius/go-vein` or similar per `go.mod`)
- No external runtime dependencies — stdlib only (`sync`, `sync/atomic`, `context`)
- Go 1.21+ required for generic type inference improvements
