## 1. Subscription State

- [x] 1.1 Refactor subscriber channel state so each `OnC()` call allocates and registers a new buffered channel instead of reusing a singleton channel.
- [x] 1.2 Update publish fan-out logic to deliver each event to every registered `OnC()` channel with independent non-blocking drop handling.
- [x] 1.3 Preserve `Unsubscribe()` and `DroppedEvents()` behavior under concurrent `On`, `OnC`, `Publish`, and `Unsubscribe` usage.

## 2. Tests And Docs

- [x] 2.1 Replace the repeated-`OnC()` identity test with assertions that separate `OnC()` calls return distinct channels.
- [x] 2.2 Add coverage proving two channels created from one subscription both receive the same published event and that a full channel still does not block publish.
- [x] 2.3 Update README and Go doc comments to describe per-call channel creation and fan-out delivery semantics for `OnC()`.

## 3. Validation

- [x] 3.1 Run the relevant Go test suite, including concurrency-sensitive subscription tests, and fix any regressions caused by the new fan-out behavior.