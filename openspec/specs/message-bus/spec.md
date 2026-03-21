# message-bus Specification

## Purpose
TBD - created by archiving change lightweight-message-bus. Update Purpose after archive.
## Requirements
### Requirement: Dispatcher stores per-type topic registries
The dispatcher SHALL maintain an isolated subscriber registry for each distinct event type `T`, ensuring that publishing event type `A` never invokes subscribers registered for event type `B`.

#### Scenario: Isolated dispatch per type
- **WHEN** two subscribers are registered — one for `EventA` and one for `EventB` — and `EventA` is published
- **THEN** only the `EventA` subscriber is invoked; the `EventB` subscriber receives no call

### Requirement: Publisher publishes a typed event with a single call
The dispatcher SHALL provide a `Publish[T](d *Dispatcher, event T)` function (and a package-level `Publish[T](event T)` shorthand) that delivers the event to all currently registered subscribers of type `T`.

#### Scenario: Publish reaches all registered subscribers
- **WHEN** three subscribers for `MyEvent` are registered and `Publish(MyEvent{Value: 42})` is called
- **THEN** all three subscribers receive an event with `Value == 42`

#### Scenario: Publish with no subscribers is a no-op
- **WHEN** no subscribers are registered for `MyEvent` and `Publish(MyEvent{})` is called
- **THEN** the call returns immediately without error or panic

### Requirement: Dispatch is non-blocking from the publisher's perspective
A `Publish` call SHALL return without waiting for any subscriber handler to complete, regardless of how long handlers take.

#### Scenario: Slow subscriber does not stall publisher
- **WHEN** a subscriber's `On` handler blocks for 1 second and `Publish` is called
- **THEN** `Publish` returns before the handler finishes

### Requirement: Lock-free read path on publish
The dispatcher SHALL use an `atomic.Pointer` over a copy-on-write subscriber slice so that `Publish` does not acquire any mutex.

#### Scenario: Concurrent publishes do not block each other
- **WHEN** 1000 goroutines concurrently call `Publish` for the same event type
- **THEN** all calls complete without deadlock or data race (verified by `-race`)

### Requirement: Subscribe and Unsubscribe are safe under concurrency
The dispatcher SHALL allow goroutines to call `Subscribe` and `Unsubscribe` concurrently with `Publish` without data races.

#### Scenario: Subscribe during active publishing
- **WHEN** a goroutine calls `Subscribe` while another goroutine is concurrently publishing the same event type
- **THEN** no panic or data race occurs; the new subscriber may or may not receive the in-flight event

### Requirement: Zero-value Dispatcher is usable
A `Dispatcher` value SHALL be usable immediately after declaration (`var d Dispatcher`) without calling any constructor.

#### Scenario: No-init usage
- **WHEN** a caller declares `var d vein.Dispatcher` and immediately calls `Publish` or `Subscribe`
- **THEN** no panic occurs and the call behaves correctly

