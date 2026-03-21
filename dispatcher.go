package vein

import (
	"reflect"
	"sync"
)

// Dispatcher is a type-safe event dispatcher.
// Its zero value is immediately usable — no constructor needed.
//
// For most use-cases the package-level [Subscribe] and [Publish] functions,
// which operate on a shared default dispatcher, are sufficient.
// Use an explicit Dispatcher when isolation is required (e.g. in tests).
type Dispatcher struct {
	// topics maps reflect.Type(*T) → *topic[T].
	// sync.Map is optimal here: written rarely (per unique event type),
	// read heavily (every Publish call).
	topics sync.Map
}

// getOrCreateTopic returns the topic for event type T, creating it on first use.
// The reflect.Type key encodes the full concrete type, ensuring isolation.
func getOrCreateTopic[T any](d *Dispatcher) *topic[T] {
	key := reflect.TypeOf((*T)(nil))
	if v, ok := d.topics.Load(key); ok {
		return v.(*topic[T]) // fast path: no allocation
	}
	t := &topic[T]{}
	actual, _ := d.topics.LoadOrStore(key, t)
	return actual.(*topic[T])
}

// SubscribeTo registers a new subscription on d for events of type T.
// Returns a [Subscription] that configures delivery and cancels the registration.
//
//	sub := vein.SubscribeTo[MyEvent](&d)
//	sub.On(func(e MyEvent) { ... })
func SubscribeTo[T any](d *Dispatcher) Subscription[T] {
	t := getOrCreateTopic[T](d)
	s := &subscriber[T]{}
	t.add(s)
	return Subscription[T]{node: s, t: t}
}

// PublishTo delivers event to all subscribers of type T registered on d.
// It returns immediately; handlers run asynchronously.
//
//	vein.PublishTo(&d, MyEvent{ID: 1})
func PublishTo[T any](d *Dispatcher, event T) {
	key := reflect.TypeOf((*T)(nil))
	v, ok := d.topics.Load(key)
	if !ok {
		return // no subscribers for T — fast exit, no allocation
	}
	v.(*topic[T]).publish(event)
}
