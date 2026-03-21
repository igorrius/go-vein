## Why

`Subscription.OnC()` currently memoizes a single channel per subscription. That makes repeated `OnC()` calls share one stream instead of behaving like `On`, where each registration gets its own delivery target. For callers that expect fan-out semantics, this creates hidden coupling between consumers and prevents independent channel-based listeners on the same subscription.

## What Changes

- Change `Subscription[T].OnC()` so each call creates and registers a new buffered channel receiver on the subscription.
- Deliver each published event to every channel created by `OnC()`, using the same non-blocking drop-on-full behavior per channel.
- Update tests and documentation to describe per-call channel creation and fan-out delivery semantics.
- **BREAKING**: repeated `OnC()` calls on the same subscription will no longer return the same channel instance.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `subscription`: change `OnC()` requirements from shared-channel reuse to per-call channel registration with fan-out delivery.

## Impact

- Affected public behavior: `Subscription[T].OnC()` semantics change while keeping the same signature.
- Affected code: channel subscriber state management, publish fan-out logic, tests, and README examples.
- No new external dependencies or module changes.