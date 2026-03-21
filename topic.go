package vein

import "sync/atomic"

const defaultChanBuf = 64

// subscriberData is an immutable snapshot of a subscriber's delivery configuration.
// It is replaced atomically on every On/OnC call (copy-on-write).
type subscriberData[T any] struct {
	on []func(T)
	ch chan T
}

// subscriber is a single subscription node held inside a topic.
type subscriber[T any] struct {
	data    atomic.Pointer[subscriberData[T]]
	dropped atomic.Uint64
}

// addHandler appends fn to the handler list using a lock-free COW update.
func (s *subscriber[T]) addHandler(fn func(T)) {
	for {
		old := s.data.Load()
		var next *subscriberData[T]
		if old == nil {
			next = &subscriberData[T]{on: []func(T){fn}}
		} else {
			on := make([]func(T), len(old.on)+1)
			copy(on, old.on)
			on[len(old.on)] = fn
			next = &subscriberData[T]{on: on, ch: old.ch}
		}
		if s.data.CompareAndSwap(old, next) {
			return
		}
	}
}

// getOrCreateChan lazily initialises the delivery channel using a lock-free COW update.
// Concurrent calls are safe; only one channel is ever created.
func (s *subscriber[T]) getOrCreateChan(buf int) chan T {
	for {
		old := s.data.Load()
		if old != nil && old.ch != nil {
			return old.ch
		}
		ch := make(chan T, buf)
		var next *subscriberData[T]
		if old == nil {
			next = &subscriberData[T]{ch: ch}
		} else {
			next = &subscriberData[T]{on: old.on, ch: ch}
		}
		if s.data.CompareAndSwap(old, next) {
			return ch
		}
		// Another goroutine won the CAS; retry — next Load will see their channel.
	}
}

// topic manages the set of subscribers for a single event type T.
// The subscriber slice is stored behind an atomic pointer for lock-free reads.
type topic[T any] struct {
	subs atomic.Pointer[[]*subscriber[T]]
}

// add appends s to the subscriber list using a lock-free COW update.
func (t *topic[T]) add(s *subscriber[T]) {
	for {
		old := t.subs.Load()
		var next []*subscriber[T]
		if old == nil {
			next = []*subscriber[T]{s}
		} else {
			next = make([]*subscriber[T], len(*old)+1)
			copy(next, *old)
			next[len(*old)] = s
		}
		if t.subs.CompareAndSwap(old, &next) {
			return
		}
	}
}

// remove deletes s from the subscriber list by pointer identity using a lock-free COW update.
func (t *topic[T]) remove(s *subscriber[T]) {
	for {
		old := t.subs.Load()
		if old == nil || len(*old) == 0 {
			return
		}
		next := make([]*subscriber[T], 0, len(*old))
		for _, sub := range *old {
			if sub != s {
				next = append(next, sub)
			}
		}
		if t.subs.CompareAndSwap(old, &next) {
			return
		}
	}
}

// publish delivers event to all current subscribers.
//
// The subscriber slice is loaded with a single atomic read (lock-free hot path).
// - Each On handler is dispatched in its own goroutine for concurrent execution.
// - Each OnC channel receives a non-blocking send; overflows are counted and dropped.
func (t *topic[T]) publish(event T) {
	p := t.subs.Load()
	if p == nil || len(*p) == 0 {
		return
	}
	for _, s := range *p {
		d := s.data.Load()
		if d == nil {
			continue
		}
		// Launch each registered callback in its own goroutine.
		for _, fn := range d.on {
			fn := fn // per-iteration capture (safe in all Go versions)
			go fn(event)
		}
		// Non-blocking channel send; drop and count if buffer is full.
		if d.ch != nil {
			select {
			case d.ch <- event:
			default:
				s.dropped.Add(1)
			}
		}
	}
}
