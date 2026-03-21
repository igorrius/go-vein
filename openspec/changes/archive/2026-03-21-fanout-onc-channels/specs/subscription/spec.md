## MODIFIED Requirements

### Requirement: OnC returns a non-blocking receive channel
`Subscription[T].OnC()` SHALL return a buffered `<-chan T` that receives published events. Each call to `OnC()` on the same subscription SHALL create and register a new channel listener. Every registered channel listener SHALL receive future published events independently. If any channel buffer is full at publish time, delivery to that channel SHALL be silently dropped rather than blocking the publisher.

#### Scenario: Channel receives published event
- **WHEN** `ch := sub.OnC()` and `MyEvent{ID: 7}` is published
- **THEN** `<-ch` returns `MyEvent{ID: 7}`

#### Scenario: Full channel does not block publisher
- **WHEN** a channel returned by `OnC()` is full and `Publish` is called
- **THEN** `Publish` returns immediately without blocking; delivery to that channel is dropped

#### Scenario: OnC called multiple times returns distinct channels
- **WHEN** `sub.OnC()` is called twice on the same `Subscription`
- **THEN** the two calls return different channel instances

#### Scenario: Multiple OnC channels all receive the same event
- **WHEN** two channels are created via `ch1 := sub.OnC()` and `ch2 := sub.OnC()` and an event is published
- **THEN** both `ch1` and `ch2` receive that event independently