package vein

// Subscription represents a typed registration on a [Dispatcher].
//
// Delivery modes (may both be used simultaneously):
//   - [Subscription.On]  — each event fires fn in a new goroutine.
//   - [Subscription.OnC] — each event is sent to a buffered channel.
//
// Call [Subscription.Unsubscribe] to stop all delivery.
// Subscription is a value type and safe to copy.
type Subscription[T any] struct {
	node *subscriber[T]
	t    *topic[T]
}

// On registers fn as a handler for published events.
// Every time an event is published, fn is invoked in a new goroutine,
// independently and concurrently with all other registered handlers.
// Multiple On calls on the same Subscription register multiple handlers.
func (s Subscription[T]) On(fn func(T)) {
	s.node.addHandler(fn)
}

// OnC returns a buffered channel that receives published events.
// The buffer holds up to 64 events by default. When the buffer is full at
// publish time the event is silently dropped and counted by [Subscription.DroppedEvents].
//
// Repeated calls return the same channel.
//
// Note: the channel is not closed when Unsubscribe is called.
// For clean shutdown, use a select with a done channel or context:
//
//	ch := sub.OnC()
//	for {
//	    select {
//	    case e := <-ch:
//	        handle(e)
//	    case <-ctx.Done():
//	        sub.Unsubscribe()
//	        return
//	    }
//	}
func (s Subscription[T]) OnC() <-chan T {
	return s.node.getOrCreateChan(defaultChanBuf)
}

// Unsubscribe removes this subscription from the dispatcher.
// After Unsubscribe returns, no new events will be delivered via On or OnC.
// An event published concurrently with Unsubscribe may still be delivered.
// Unsubscribe is safe to call multiple times (idempotent).
func (s Subscription[T]) Unsubscribe() {
	s.t.remove(s.node)
}

// DroppedEvents returns the number of events dropped because the OnC
// channel buffer was full at the time of publishing.
func (s Subscription[T]) DroppedEvents() uint64 {
	return s.node.dropped.Load()
}
