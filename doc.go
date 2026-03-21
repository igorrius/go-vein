// Package vein is a lightweight, type-safe, allocation-minimal message bus for Go.
//
// Events are plain Go structs. Subscriptions are typed via generics — no
// interface{} magic, no reflection visible to callers.
//
// # Quick start (global dispatcher)
//
//	type OrderPlaced struct{ ID int }
//
//	// Subscribe
//	sub := vein.Subscribe[OrderPlaced]()
//	sub.On(func(e OrderPlaced) {
//	    fmt.Println("order:", e.ID)
//	})
//
//	// Or receive via channel
//	ch := sub.OnC()
//
//	// Publish — type is inferred
//	vein.Publish(OrderPlaced{ID: 42})
//
//	// Stop receiving
//	sub.Unsubscribe()
//
// # Isolated dispatcher (e.g. in tests)
//
//	var d vein.Dispatcher // zero value, no constructor needed
//	sub := vein.SubscribeTo[OrderPlaced](&d)
//	vein.PublishTo(&d, OrderPlaced{ID: 1})
//
// # Delivery guarantees
//
//   - [Subscription.On]: each publish fires the handler in a new goroutine.
//     Multiple On registrations on the same Subscription all fire concurrently.
//   - [Subscription.OnC]: each call creates a buffered channel listener (64 slots).
//     Publishes fan out to all registered channels. A full buffer causes that
//     channel delivery to be dropped; see [Subscription.DroppedEvents].
//   - [Dispatcher.PublishTo] / [Publish] return immediately — they never block.
//
// # Performance
//
// The subscriber list is stored behind an atomic pointer (copy-on-write).
// Publish reads the list with a single atomic load — no mutex on the hot path.
// Subscription management (Subscribe/Unsubscribe) is O(n) but lock-free.
package vein
