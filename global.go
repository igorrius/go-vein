package vein

// defaultDispatcher is the package-level dispatcher shared by Subscribe and Publish.
// Its zero value is valid — no initialisation needed.
var defaultDispatcher Dispatcher

// Subscribe registers a new subscription on the default [Dispatcher] for events of type T.
// It is shorthand for [SubscribeTo][T](&defaultDispatcher).
//
//	sub := vein.Subscribe[OrderPlaced]()
//	sub.On(func(e OrderPlaced) { fmt.Println(e.ID) })
func Subscribe[T any]() Subscription[T] {
	return SubscribeTo[T](&defaultDispatcher)
}

// Publish delivers event to all subscribers of type T on the default [Dispatcher].
// The event type is inferred from the argument — no explicit type parameter needed.
//
//	vein.Publish(OrderPlaced{ID: 42})
func Publish[T any](event T) {
	PublishTo[T](&defaultDispatcher, event)
}
