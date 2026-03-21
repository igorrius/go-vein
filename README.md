# go-vein

A lightweight, type-safe, allocation-minimal message bus for Go.

- **Generics-first** — events are typed Go structs; no `interface{}` or reflection visible to callers
- **Lock-free hot path** — the subscriber list is read with a single atomic load; no mutex on `Publish`
- **Two delivery modes** — callback (`On`) and channel (`OnC`) per subscription
- **Non-blocking dispatch** — `Publish` never blocks regardless of slow or absent subscribers
- **Global default dispatcher** — `Subscribe[T]()` / `Publish(event)` work right out of the box
- **Zero-value ready** — `var d Dispatcher` needs no initialisation

## Requirements

Go 1.21+

## Install

```sh
go get github.com/igorrius/go-vein
```

## Quick start

```go
package main

import (
    "fmt"
    vein "github.com/igorrius/go-vein"
)

type OrderPlaced struct{ ID int }

func main() {
    // Subscribe on the default dispatcher
    sub := vein.Subscribe[OrderPlaced]()
    sub.On(func(e OrderPlaced) {
        fmt.Println("order:", e.ID)
    })

    // Publish — type is inferred from the argument
    vein.Publish(OrderPlaced{ID: 42})

    // Unsubscribe when done
    sub.Unsubscribe()
}
```

## Delivery modes

### On — concurrent callbacks

Each `On` registration fires in its own goroutine per publish.
Multiple registrations on the same Subscription all fire concurrently.

```go
sub := vein.Subscribe[OrderPlaced]()
sub.On(func(e OrderPlaced) { /* runs in new goroutine */ })
sub.On(func(e OrderPlaced) { /* also runs concurrently */ })
```

### OnC — channel delivery

`OnC` returns a buffered `<-chan T` (64 slots). Overflow events are dropped silently.
Use `select` + a done channel for clean shutdown; the channel is not closed on `Unsubscribe`.

```go
sub := vein.Subscribe[OrderPlaced]()
ch := sub.OnC()

go func() {
    for {
        select {
        case e := <-ch:
            fmt.Println("channel:", e.ID)
        case <-ctx.Done():
            sub.Unsubscribe()
            return
        }
    }
}()
```

Check how many events were silently dropped:

```go
fmt.Println("dropped:", sub.DroppedEvents())
```

## Isolated dispatcher

Use an explicit `Dispatcher` when isolation is required (e.g. in tests):

```go
var d vein.Dispatcher // zero value, no constructor needed

sub := vein.SubscribeTo[OrderPlaced](&d)
sub.On(func(e OrderPlaced) { fmt.Println(e.ID) })

vein.PublishTo(&d, OrderPlaced{ID: 1})
```

## Performance

Benchmarks on Intel i7-1260P (amd64):

| Scenario | ns/op | allocs/op |
|---|---|---|
| Publish, no subscribers | 11 | 0 |
| Publish → OnC channel | 50 | 0 |
| Publish → On callback | 268 | 1 |
| Subscribe + Unsubscribe | 133 | 5 |

The `On` allocation (1/publish) is the goroutine stack. The `OnC` path is zero-alloc.

## Design

- Per-type topic registry via `sync.Map` keyed by `reflect.Type`
- Subscriber list stored behind `atomic.Pointer` (copy-on-write); `Publish` reads it lock-free
- `Subscription[T]` is a value type; `subscriber[T]` state updated via lock-free CAS
- Non-blocking channel send with `select { case ch <- e: default: dropped++ }`
