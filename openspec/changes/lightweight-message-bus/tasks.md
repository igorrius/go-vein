## 1. Project Bootstrap

- [x] 1.1 Verify `go.mod` module path and ensure Go 1.21+ version directive
- [x] 1.2 Create package directory structure: `vein/` (or root package)

## 2. Core Types & Interfaces

- [x] 2.1 Define `subscriber[T any]` internal node struct (holds `on []func(T)`, `ch chan T`, `unsub func()`)
- [x] 2.2 Define `topic[T any]` struct with `atomic.Pointer[[]subscriber[T]]` field
- [x] 2.3 Implement `topic[T].add(s *subscriber[T])` — COW append, atomic swap
- [x] 2.4 Implement `topic[T].remove(s *subscriber[T])` — COW filter by pointer, atomic swap
- [x] 2.5 Implement `topic[T].publish(event T)` — load pointer, iterate, dispatch each subscriber

## 3. Dispatcher

- [x] 3.1 Define `Dispatcher` struct with `sync.Map` field for per-type topic storage
- [x] 3.2 Implement `getOrCreateTopic[T](d *Dispatcher) *topic[T]` helper using `sync.Map.LoadOrStore`
- [x] 3.3 Implement `Subscribe[T any](d *Dispatcher) Subscription[T]` — allocates subscriber node, registers via topic, returns Subscription
- [x] 3.4 Implement `Publish[T any](d *Dispatcher, event T)` — resolves topic and calls `topic.publish`
- [x] 3.5 Ensure zero-value `Dispatcher` works without any constructor (no `sync.Map` init needed)

## 4. Subscription

- [x] 4.1 Define `Subscription[T any]` struct with unexported `node *subscriber[T]` and `topic *topic[T]`
- [x] 4.2 Implement `(s Subscription[T]) On(fn func(T))` — appends fn to `node.on` under a mutex on the node
- [x] 4.3 Implement `(s Subscription[T]) OnC() <-chan T` — lazily creates buffered `chan T` (buffer 64) on `node.ch`, returns it
- [x] 4.4 Implement `(s Subscription[T]) Unsubscribe()` — calls `topic.remove(node)`, closes `node.ch` if non-nil
- [x] 4.5 Add `dropped atomic.Uint64` counter to `subscriber[T]`; increment on non-blocking channel drop
- [x] 4.6 Expose `(s Subscription[T]) DroppedEvents() uint64` reading the atomic counter

## 5. Delivery Logic

- [x] 5.1 In `topic[T].publish`: for each subscriber, launch `go fn(event)` for every handler in `node.on`
- [x] 5.2 In `topic[T].publish`: for each subscriber with a non-nil `node.ch`, do non-blocking send (`select { case ch <- event: default: dropped++ }`)

## 6. Global Dispatcher

- [x] 6.1 Declare `var defaultDispatcher Dispatcher` at package level
- [x] 6.2 Implement package-level `Subscribe[T any]() Subscription[T]` delegating to `defaultDispatcher`
- [x] 6.3 Implement package-level `Publish[T any](event T)` delegating to `defaultDispatcher`

## 7. Tests

- [x] 7.1 Unit test: isolated dispatch per type (publish A does not fire B subscriber)
- [x] 7.2 Unit test: `On` callback receives published event
- [x] 7.3 Unit test: `OnC` channel receives published event
- [x] 7.4 Unit test: `Unsubscribe` stops delivery
- [x] 7.5 Unit test: `Unsubscribe` is idempotent (no panic on double call)
- [x] 7.6 Unit test: `OnC` full buffer drops event, publisher does not block
- [x] 7.7 Unit test: `DroppedEvents` counter increments on drop
- [x] 7.8 Race test: concurrent Publish + Subscribe + Unsubscribe (`go test -race`)
- [x] 7.9 Unit test: zero-value `Dispatcher` usable without constructor
- [x] 7.10 Unit test: package-level `Subscribe` / `Publish` work correctly

## 8. Benchmarks

- [x] 8.1 `BenchmarkPublish_OnHandler` — single On subscriber, measure allocs/op
- [x] 8.2 `BenchmarkPublish_OnCChannel` — single OnC subscriber, measure allocs/op
- [x] 8.3 `BenchmarkPublish_NoSubscribers` — baseline, empty topic
- [x] 8.4 `BenchmarkSubscribeUnsubscribe` — subscribe/unsubscribe cycle

## 9. Documentation

- [x] 9.1 Write package-level doc comment in `doc.go` explaining the core concepts and quick-start example
- [x] 9.2 Add godoc comments to exported types: `Dispatcher`, `Subscription`, `Subscribe`, `Publish`
- [x] 9.3 Update `README.md` with usage examples covering `On`, `OnC`, `Unsubscribe`, and global dispatcher
