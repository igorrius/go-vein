package vein_test

import (
"testing"

vein "github.com/igorrius/go-vein"
)

type benchEvent struct{ Payload int64 }

func BenchmarkPublish_OnHandler(b *testing.B) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[benchEvent](&d)
	sub.On(func(benchEvent) {})
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		vein.PublishTo(&d, benchEvent{Payload: int64(i)})
	}
}

func BenchmarkPublish_OnCChannel(b *testing.B) {
	var d vein.Dispatcher
	sub := vein.SubscribeTo[benchEvent](&d)
	ch := sub.OnC()
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-ch:
			case <-stop:
				return
			}
		}
	}()
	b.Cleanup(func() {
		sub.Unsubscribe()
		close(stop)
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		vein.PublishTo(&d, benchEvent{Payload: int64(i)})
	}
}

func BenchmarkPublish_NoSubscribers(b *testing.B) {
	var d vein.Dispatcher
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		vein.PublishTo(&d, benchEvent{Payload: int64(i)})
	}
}

func BenchmarkSubscribeUnsubscribe(b *testing.B) {
	var d vein.Dispatcher
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sub := vein.SubscribeTo[benchEvent](&d)
		sub.Unsubscribe()
	}
}
