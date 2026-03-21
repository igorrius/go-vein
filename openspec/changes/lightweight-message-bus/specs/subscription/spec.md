## ADDED Requirements

### Requirement: Subscribe returns a typed Subscription value
`Subscribe[T](d *Dispatcher)` SHALL return a `Subscription[T]` value that represents the caller's registration and allows them to configure delivery mode and cancel the subscription.

#### Scenario: Subscribe returns non-zero Subscription
- **WHEN** a caller invokes `Subscribe[MyEvent](&d)`
- **THEN** a `Subscription[MyEvent]` is returned with a valid `Unsubscribe` capability

### Requirement: On registers a callback invoked concurrently per publish
`Subscription[T].On(fn func(T))` SHALL register `fn` as a handler. Each time the event is published, `fn` SHALL be invoked in a new goroutine, independently and concurrently with all other registered handlers for the same subscription.

#### Scenario: Handler fires on publish
- **WHEN** `sub.On(func(e MyEvent) { received <- e })` is registered and `MyEvent{ID: 1}` is published
- **THEN** `fn` is called with `MyEvent{ID: 1}`

#### Scenario: Multiple On registrations on the same Subscription all fire
- **WHEN** two handlers are registered via `sub.On(fn1)` and `sub.On(fn2)` and an event is published
- **THEN** both `fn1` and `fn2` are invoked

#### Scenario: On handlers execute concurrently
- **WHEN** two slow On handlers are registered and an event is published
- **THEN** both handlers start before either finishes (total wall time < sum of individual times)

### Requirement: OnC returns a non-blocking receive channel
`Subscription[T].OnC()` SHALL return a `<-chan T` that receives published events. The channel SHALL be buffered (default buffer ≥ 1). If the buffer is full at publish time, the event SHALL be silently dropped rather than blocking the publisher.

#### Scenario: Channel receives published event
- **WHEN** `ch := sub.OnC()` and `MyEvent{ID: 7}` is published
- **THEN** `<-ch` returns `MyEvent{ID: 7}`

#### Scenario: Full channel does not block publisher
- **WHEN** the channel returned by `OnC()` is full and `Publish` is called
- **THEN** `Publish` returns immediately without blocking; the event is dropped

#### Scenario: OnC called multiple times returns the same channel
- **WHEN** `sub.OnC()` is called twice on the same `Subscription`
- **THEN** both calls return the same channel instance

### Requirement: Unsubscribe removes the subscription from the dispatcher
`Subscription[T].Unsubscribe()` SHALL remove the subscription from the dispatcher so that future publishes do not reach its handlers or channel.

#### Scenario: No delivery after Unsubscribe
- **WHEN** `sub.Unsubscribe()` is called and then `Publish` is called
- **THEN** neither `On` handlers nor the `OnC` channel receive the event

#### Scenario: Unsubscribe is idempotent
- **WHEN** `sub.Unsubscribe()` is called twice
- **THEN** no panic or error occurs

### Requirement: Subscription is safe to use from multiple goroutines
All methods on `Subscription[T]` (`On`, `OnC`, `Unsubscribe`) SHALL be safe to call concurrently.

#### Scenario: Concurrent On and Unsubscribe
- **WHEN** one goroutine calls `sub.On(fn)` while another concurrently calls `sub.Unsubscribe()`
- **THEN** no panic or data race occurs
