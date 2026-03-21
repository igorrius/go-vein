## ADDED Requirements

### Requirement: Package-level default Dispatcher exists without initialisation
The `vein` package SHALL expose a package-level default `*Dispatcher` that is ready for use when the package is imported, requiring no explicit initialisation by the caller.

#### Scenario: Default dispatcher ready on import
- **WHEN** a caller imports `vein` and does not call any constructor
- **THEN** `vein.Publish` and `vein.Subscribe` work correctly

### Requirement: Subscribe free function delegates to the default Dispatcher
The package SHALL provide a `Subscribe[T]() Subscription[T]` free function that registers a new subscription on the default dispatcher.

#### Scenario: Package-level Subscribe works
- **WHEN** `sub := vein.Subscribe[MyEvent]()` is called
- **THEN** a valid `Subscription[MyEvent]` is returned and subsequent publishes reach it

### Requirement: Publish free function delegates to the default Dispatcher
The package SHALL provide a `Publish[T](event T)` free function that publishes to the default dispatcher.

#### Scenario: Package-level Publish reaches package-level subscriber
- **WHEN** `vein.Subscribe[MyEvent]().On(fn)` is registered and `vein.Publish(MyEvent{})` is called
- **THEN** `fn` is invoked with the published event

### Requirement: Default Dispatcher is safe for concurrent use
The default dispatcher SHALL be safe to use from multiple goroutines simultaneously without external synchronisation.

#### Scenario: Concurrent package-level Subscribe and Publish
- **WHEN** 100 goroutines concurrently call `vein.Subscribe[MyEvent]()` and 100 goroutines concurrently call `vein.Publish(MyEvent{})` 
- **THEN** no data race is detected by the Go race detector
