## Context

`Subscription[T].OnC()` currently lazily creates a single channel and returns that same instance on every call. This differs from `On`, where each call adds an independent delivery target. The change is small at the API surface but cross-cuts subscriber state, publish fan-out behavior, tests, and documentation because channel delivery is currently modeled as a single optional field on the subscriber.

## Goals / Non-Goals

**Goals:**
- Make each `OnC()` call register an independent buffered channel for the same subscription.
- Preserve non-blocking publish behavior by treating each channel send independently with drop-on-full semantics.
- Keep `Subscription[T]` safe under concurrent `On`, `OnC`, `Publish`, and `Unsubscribe` calls.
- Keep the public method signature unchanged.

**Non-Goals:**
- Changing channel buffer sizing or adding per-channel configuration.
- Closing channels on `Unsubscribe`.
- Introducing delivery ordering or reliability guarantees between channel consumers.

## Decisions

### D1 - Represent channel listeners as a collection, not a singleton

**Decision:** Replace the single lazily initialized channel field on the subscriber with a collection of registered channels so each `OnC()` call appends a new channel.

**Alternatives considered:**
- Keep one shared channel and document multiplexed consumption more clearly: rejected because it preserves the current surprising behavior.
- Allocate a new subscription per `OnC()` call internally: rejected because it would complicate `Unsubscribe()` semantics and distort `DroppedEvents()` accounting.

**Rationale:** `On` already models repeated registration as fan-out. `OnC` should match that mental model while preserving one `Subscription` handle for lifecycle control.

### D2 - Snapshot handlers and channels together during publish

**Decision:** The subscriber should expose a concurrency-safe snapshot of both callback handlers and channel listeners so publish can iterate without holding locks while it dispatches.

**Alternatives considered:**
- Hold a subscriber mutex while sending to channels: rejected because even non-blocking sends would serialize concurrent publishers and registration.
- Store channels in a separate atomic structure from handlers: possible, but rejected unless needed because it increases state coordination complexity.

**Rationale:** The current design already favors lock-free or short critical sections around publish. A snapshot-based approach keeps publish behavior aligned with the rest of the package.

### D3 - Count drops per failed channel send

**Decision:** If an event cannot be enqueued because one registered channel is full, increment `DroppedEvents()` once for that failed send, even if sibling channels receive the same event successfully.

**Alternatives considered:**
- Count drops per published event regardless of how many channels were full: rejected because it hides which delivery targets are overloaded.

**Rationale:** The counter should reflect actual dropped deliveries. With multiple `OnC()` channels, delivery pressure is now per channel, not per subscription-wide singleton.

## Risks / Trade-offs

- More memory per subscription when callers register many channels -> Mitigation: each channel remains opt-in and lazily allocated only when `OnC()` is called.
- More work per publish because each event fans out to every registered channel -> Mitigation: this is the intended semantic change; sends remain non-blocking and bounded per channel.
- Existing callers relying on shared-channel identity will break semantically -> Mitigation: document the breaking behavior in README and release notes, and add regression tests for distinct-channel fan-out.

## Migration Plan

1. Update the `subscription` delta spec to require distinct channels per `OnC()` call.
2. Refactor subscriber state and publish logic to support multiple channels.
3. Replace identity-based tests with fan-out tests that assert distinct channels and shared delivery.
4. Update README examples to clarify that every `OnC()` call creates a new listener.

Rollback is straightforward: restore the previous singleton-channel state model and revert the spec change.

## Open Questions

None.