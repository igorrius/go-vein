package vein_test

import (
"sync"
"sync/atomic"
"testing"
"time"

vein "github.com/igorrius/go-vein"
)

type EventA struct{ Value int }
type EventB struct{ Value int }

// 7.1 Isolated dispatch per type
func TestIsolatedDispatchPerType(t *testing.T) {
	var d vein.Dispatcher
	var bReceived atomic.Bool
	subB := vein.SubscribeTo[EventB](&d)
	subB.On(func(EventB) { bReceived.Store(true) })

	vein.PublishTo(&d, EventA{Value: 1})
	time.Sleep(20 * time.Millisecond)

	if bReceived.Load() {
		t.Fatal("EventB subscriber must not receive an EventA publish")
	}
}

// 7.2 On callback receives published event
func TestOnCallbackReceivesEvent(t *testing.T) {
	var d vein.Dispatcher
	received := make(chan EventA, 1)
	sub := vein.SubscribeTo[EventA](&d)
	sub.On(func(e EventA) { received <- e })

	vein.PublishTo(&d, EventA{Value: 42})

	select {
	case e := <-received:
		if e.Value != 42 {
			t.Fatalf("expected Value=42, got %d", e.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout: On handler was not called")
	}
}

// 7.3 OnC channel receives published event
func TestOnCChannelReceivesEvent(t *testing.T) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[EventA](&d)
	ch := sub.OnC()

	vein.PublishTo(&d, EventA{Value: 7})

	select {
	case e := <-ch:
		if e.Value != 7 {
			t.Fatalf("expected Value=7, got %d", e.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout: OnC channel received nothing")
	}
}

func TestOnCReturnsSameChannel(t *testing.T) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[EventA](&d)
	ch1 := sub.OnC()
	ch2 := sub.OnC()
	if ch1 != ch2 {
		t.Fatal("OnC must return the same channel on repeated calls")
	}
}

func TestMultipleOnHandlersAllFire(t *testing.T) {
	var d vein.Dispatcher
	var wg sync.WaitGroup
	wg.Add(2)
	sub := vein.SubscribeTo[EventA](&d)
	sub.On(func(EventA) { wg.Done() })
	sub.On(func(EventA) { wg.Done() })

	vein.PublishTo(&d, EventA{})

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout: not all On handlers fired")
	}
}

// 7.4 Unsubscribe stops delivery
func TestUnsubscribeStopsDelivery(t *testing.T) {
	var d vein.Dispatcher
	var count atomic.Int32
	sub := vein.SubscribeTo[EventA](&d)
	sub.On(func(EventA) { count.Add(1) })

	vein.PublishTo(&d, EventA{})
	time.Sleep(20 * time.Millisecond)

	sub.Unsubscribe()

	vein.PublishTo(&d, EventA{})
	time.Sleep(20 * time.Millisecond)

	if n := count.Load(); n > 1 {
		t.Fatalf("expected at most 1 delivery after Unsubscribe, got %d", n)
	}
}

// 7.5 Unsubscribe is idempotent
func TestUnsubscribeIsIdempotent(t *testing.T) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[EventA](&d)
	sub.Unsubscribe()
	sub.Unsubscribe()
}

// 7.6 OnC full buffer drops event; publisher does not block
func TestOnCFullBufferDropsEvent(t *testing.T) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[EventA](&d)
	_ = sub.OnC()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			vein.PublishTo(&d, EventA{Value: i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on full OnC channel")
	}
}

// 7.7 DroppedEvents counter increments on drop
func TestDroppedEventsCounter(t *testing.T) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[EventA](&d)
	_ = sub.OnC()

	for i := 0; i < 500; i++ {
		vein.PublishTo(&d, EventA{})
	}

	if sub.DroppedEvents() == 0 {
		t.Fatal("expected DroppedEvents > 0 after flooding channel")
	}
}

// 7.8 Race test: concurrent Publish + Subscribe + Unsubscribe
func TestConcurrentPublishSubscribeUnsubscribe(t *testing.T) {
	var d vein.Dispatcher
	var wg sync.WaitGroup
	const goroutines = 100
	wg.Add(goroutines * 3)

	for range goroutines {
		go func() {
			defer wg.Done()
			sub := vein.SubscribeTo[EventA](&d)
			sub.On(func(EventA) {})
			sub.Unsubscribe()
		}()
		go func() {
			defer wg.Done()
			vein.PublishTo(&d, EventA{Value: 1})
		}()
		go func() {
			defer wg.Done()
			sub := vein.SubscribeTo[EventA](&d)
			_ = sub.OnC()
			sub.Unsubscribe()
		}()
	}

	wg.Wait()
}

// 7.9 Zero-value Dispatcher works without constructor
func TestZeroValueDispatcher(t *testing.T) {
	var d vein.Dispatcher
	received := make(chan struct{}, 1)
	sub := vein.SubscribeTo[EventA](&d)
	sub.On(func(EventA) { received <- struct{}{} })
	vein.PublishTo(&d, EventA{})

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("zero-value Dispatcher did not deliver event")
	}
}

// 7.10 Package-level Subscribe / Publish
func TestGlobalSubscribePublish(t *testing.T) {
	type globalTestEvent struct{ N int }

	received := make(chan globalTestEvent, 1)
	sub := vein.Subscribe[globalTestEvent]()
	defer sub.Unsubscribe()
	sub.On(func(e globalTestEvent) { received <- e })

	vein.Publish(globalTestEvent{N: 99})

	select {
	case e := <-received:
		if e.N != 99 {
			t.Fatalf("expected N=99, got N=%d", e.N)
		}
	case <-time.After(time.Second):
		t.Fatal("global dispatcher did not deliver event")
	}
}
