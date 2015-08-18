package event

import (
	"strconv"
	"testing"
)

// testNumHandlers is the number of handlers to use during tests.
const testNumHandlers = 100

// testEventType is the Event Type used during tests.
const testEventType = Type(2)

// testMakeAckHandler creates a HandleFunc which outputs the received Event
// on the returned channel.
func testMakeAckHandler() (HandleFunc, <-chan Event) {
	c := make(chan Event)
	f := func(e Event) {
		c <- e
	}
	return f, c
}

// testEmptyHandler is an empty HandleFunc which does nothing. Useful when
// testing other components of the Bus.
func testEmptyHandleFunc(e Event) {
	return
}

// testMakeHandlers creates a slice of Handlers wherein each Handler's name is
// the string representation of the original index of the Handler within the
// returned slice, and the func contained within the handler is the provided
// HandleFunc.
func testMakeHandlers(n int, f HandleFunc) []Handler {
	handlers := make([]Handler, n)
	for i := range handlers {
		handlers[i] = Handler{
			Id:       strconv.Itoa(i),
			Callback: f,
		}
	}

	return handlers
}

// testHandlerDiff returns a slice of strings which represent any Handler names
// which were found in the first provided slice but not the second.
func testHandlerDiff(a []Handler, b []Handler) []string {
	names := make(map[string]struct{})

	for _, handler := range a {
		names[handler.Id] = struct{}{}
	}

	for _, handler := range b {
		delete(names, handler.Id)
	}

	if len(names) == 0 {
		return nil
	}

	unmatched := make([]string, 0, len(names))
	for name := range names {
		unmatched = append(unmatched, name)
	}

	return unmatched
}

// TestNewBus checks that a Bus returned by NewBus has been correctly
// initialised.
func TestNewBus(t *testing.T) {
	bus := NewBus()
	if bus == nil {
		t.Error("Got nil instead of EventBus")
	}

	if bus.Handlers == nil {
		t.Error("EventBus handlers not initialised")
	}
}

// TestAddHandler checks that all handlers added to a Bus via AddHandler()
// are stored in the correct slice within the Bus.
func TestAddHandler(t *testing.T) {
	bus := NewBus()
	handlers := testMakeHandlers(testNumHandlers, testEmptyHandleFunc)

	for _, handler := range handlers {
		bus.AddHandler(testEventType, handler)
	}

	busHandlers := bus.Handlers[testEventType]

	if len(busHandlers) != len(handlers) {
		t.Errorf("Expected %d handlers, got %d", len(handlers), len(busHandlers))
	}

	diff := testHandlerDiff(handlers, busHandlers)
	if diff != nil {
		t.Errorf("Lists do not match. Unmatched names: %+v", diff)
	}
}

// TestDeleteHandler checks that a handler added to a Bus via DeleteHandler()
// is removed correctly from the Bus.
func TestDeleteHandler(t *testing.T) {
	bus := NewBus()
	handlers := testMakeHandlers(testNumHandlers, testEmptyHandleFunc)

	// Add all generated handlers to the bus
	for _, handler := range handlers {
		bus.AddHandler(testEventType, handler)
	}

	expectedHandlers := make([]Handler, 0, len(handlers))

	// Delete every 5th handler
	for i, h := range handlers {
		if i%5 == 0 {
			bus.DeleteHandler(testEventType, h.Id)
		} else {
			// If it's not a 5th handler, we're expecting it back at the end
			expectedHandlers = append(expectedHandlers, h)
		}
	}

	busHandlers := bus.Handlers[testEventType]

	// Bug out early if the lengths don't match
	if len(busHandlers) != len(expectedHandlers) {
		t.Errorf("Expected %d handlers, got %d", len(busHandlers), len(expectedHandlers))
	}

	diff := testHandlerDiff(expectedHandlers, busHandlers)
	if diff != nil {
		t.Errorf("Lists do not match. Unmatched names: %+v", diff)
	}
}

// TestFire checks that events Fired on a Bus are sent to the registered
// Handlers for the Event's Type.
func TestFire(t *testing.T) {
	bus := NewBus()

	// All channels we're expecting messages on
	chans := make([]<-chan Event, testNumHandlers)

	// Create handler and add to bus for every possible chan
	for i := range chans {
		fun, c := testMakeAckHandler()
		chans[i] = c

		handler := Handler{
			Id:       strconv.Itoa(i),
			Callback: fun,
		}
		bus.AddHandler(testEventType, handler)
	}

	// Create an event and fire it
	event := Event{
		Type:    testEventType,
		Payload: struct{}{},
	}

	// Fire() blocks until all Handlers have returned. Fire the event
	// concurrently so we can check return values.
	go bus.Fire(event)

	// Check that the event was sent to all registered Handlers by checking for
	// the same Event on each chan.
	for i, c := range chans {
		res := <-c
		if res != event {
			t.Errorf("Incorrect event received on chan %d. Expected %+v, got %+v", i, event, res)
		}
	}
}

func BenchmarkFire(b *testing.B) {
	bus := NewBus()
	handler := Handler{
		Id:       "test",
		Callback: testEmptyHandleFunc,
	}
	bus.AddHandler(testEventType, handler)

	event := Event{
		Type:    testEventType,
		Payload: nil,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Fire(event)
	}
}
