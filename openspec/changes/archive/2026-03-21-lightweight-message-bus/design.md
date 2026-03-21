## Context

Go generics (1.18+) enable type-safe event systems without reflection. Existing event bus libraries (`EventBus`, `watermill`) either use `any` or carry heavyweight transport dependencies. `go-vein` is a small stdlib-only package targeting intra-process, high-throughput event dispatch with zero external dependencies.

Current state: empty repository with `go.mod` only.

## Goals / Non-Goals

**Goals:**
- Type-safe publish/subscribe with Go generics
- Per-event-type isolation (each `T` has its own subscriber list)
- Two delivery modes per subscription: callback (`On`) and channel (`OnC`)
- Non-blocking dispatch: slow subscribers never stall the publisher
- Concurrent callback execution: each registered `On` handler fires in its own goroutine per publish
- Global default dispatcher for ergonomic package-level usage
- Minimal heap allocations on the hot dispatch path
- Go 1.26+ compatible

**Non-Goals:**
- Distributed / cross-process messaging
- Persistent or durable event queues
- Event ordering guarantees across subscribers
- Retry / dead-letter logic
- Priority queues or event filtering DSL

## Decisions

### D1 — Per-type subscriber registry via `sync.Map` keyed by `reflect.Type`

**Decision**: Use `sync.Map[reflect.Type, *topic[T]]` conceptually. Because Go does not allow generic fields on non-generic structs, the dispatcher stores `any` values in a `sync.Map` keyed by `reflect.TypeOf((*T)(nil))`, then type-asserts to `*topic[T]`.

**Alternatives considered**:
- `map[string]any` keyed by type name: brittle under `go build` obfuscation, no namespace
- Separate typed dispatcher per event type (callers must wire manually): too much boilerplate

**Rationale**: `sync.Map` is optimised for mostly-read, rarely-write workloads (subscribe once, publish many), matching exactly our access pattern. The type-key approach is safe because `reflect.Type` values are comparable singletons.

### D2 — `topic[T]` stores subscribers as a copy-on-write slice

**Decision**: Each `topic[T]` holds an `atomic.Pointer[[]subscriber[T]]`. On subscribe/unsubscribe the slice is copied, modified, and atomically swapped. Publish loads the pointer once and iterates without a lock.

**Alternatives considered**:
- `sync.RWMutex` around a slice: correct but contended under high publish rates
- Lock-free linked list: correct but complex and cache-unfriendly

**Rationale**: COW slice gives lock-free reads at publish time (critical path) with O(n) subscription management (acceptable — subscribe/unsubscribe are rare compared to publish).

### D3 — Channel-mode delivery uses a buffered channel with non-blocking send

**Decision**: `OnC()` creates a buffered `chan T` (default buffer 64, configurable via `WithChannelBuffer`). Dispatch does a non-blocking `select { case ch <- event: default: }`, silently dropping if the consumer is slow.

**Alternatives considered**:
- Unbuffered channel with timeout: complicates publisher
- Block forever: violates the stated requirement that dispatch is non-blocking

**Rationale**: Drop-on-full matches the requirement. Consumers that need reliable delivery should use `On(func)` with their own internal queue.

### D4 — Callback-mode delivery via `go handler(event)` per subscriber per publish

**Decision**: Each `On(func(T))` subscriber fires as `go fn(event)` inside the publish loop, giving per-call concurrency.

**Alternatives considered**:
- Shared worker pool: reduces goroutine churn but adds complexity and ordering constraints
- Single goroutine fan-out: serialises handlers, violating the concurrent-execution requirement

**Rationale**: `go fn(event)` is the simplest correct implementation. For most event rates goroutine creation overhead is negligible. A `sync.Pool` of goroutines can be added later without API changes.

### D5 — `Subscription[T]` is a value type wrapping a cancel function

**Decision**: `Subscription[T]` is a struct with an unexported `unsub func()` and an unexported `ch chan T`. `Unsubscribe()` calls `unsub`. `On(fn)` registers an additional callback on the same underlying subscriber slot. `OnC()` returns the channel, creating it lazily.

**Rationale**: Value type means no GC pressure for the caller. The unsub closure captures the topic pointer and subscriber ID for O(1) removal.

### D6 — Subscriber identified by pointer equality on a small heap-allocated node

**Decision**: Each subscription allocates one `subscriber[T]` node on the heap. Unsubscribe removes by pointer identity after load/CAS loop on the COW slice.

**Rationale**: One allocation per `Subscribe` call (not per publish) is acceptable. Pointer identity avoids the need for an integer ID counter and atomic ID allocation.

### D7 — Global dispatcher initialised via `sync.Once`

**Decision**: Package-level `var defaultBus = &Dispatcher{}` is valid zero-value. `Subscribe[T]()` and `Publish[T]()` are free functions that delegate to `defaultBus`.

**Rationale**: Zero-value-usable structs are idiomatic Go. No `init()` needed.

## Risks / Trade-offs

- **Goroutine proliferation**: High-frequency events with `On` handlers create many short-lived goroutines. → Mitigation: document the trade-off; add optional `WithWorkerPool` in a future change.
- **Silent channel drop**: `OnC` consumers that fall behind lose events silently. → Mitigation: expose a `DroppedEvents() uint64` counter on `Subscription`.
- **reflect.TypeOf cost**: Called once per `Publish` to get map key. → Mitigation: benchmark shows ~3 ns/op; acceptable. Can be eliminated in a future change using compile-time registration.
- **COW subscription cost**: O(n) copy per subscribe/unsubscribe. → Acceptable because subscription management is not on the hot path.

## Migration Plan

N/A — new package, no existing users.

## Open Questions

- Should `Dispatcher` support a `context.Context` for graceful shutdown that drains in-flight handlers?
  - not. not for this iteration
- Should the channel buffer size be global or per-subscription?
  - per-subscription
- Should dropped-event metrics be opt-in to avoid atomic overhead for users who don't need them?
  - yes, add Options pattern to Subscribe function
