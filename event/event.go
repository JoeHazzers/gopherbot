package event

import "sync"

// Type distinguishes Events from oneanother.
type Type int

// Event is an event passed fired through a Bus and handled by
// Handlers.
type Event struct {
	Type    Type
	Payload interface{}
}

// HandleFunc is a func which receives Events.
type HandleFunc func(Event)

// Handler is a func which consumes an Event.
type Handler struct {
	Id       string
	Callback HandleFunc
}

// Bus dispatches fired Events to registered handlers for that Type.
type Bus struct {
	sync.RWMutex
	Handlers map[Type][]Handler
}

// NewBus creates a new Bus.
func NewBus() *Bus {
	bus := Bus{
		Handlers: make(map[Type][]Handler, 0),
	}

	return &bus
}

// AddHandler registers an EventHandler with the EventBus to handle Events of
// EventType t.
func (bus *Bus) AddHandler(t Type, h Handler) {
	bus.Lock()
	defer bus.Unlock()

	handlers, ok := bus.Handlers[t]
	if !ok {
		handlers = make([]Handler, 0)
	}

	bus.Handlers[t] = append(handlers, h)
}

// DeleteHandler unregisters an EventHandler from the EventBus so that it
// no longer handles Events of EventType t
func (bus *Bus) DeleteHandler(t Type, id string) {
	bus.Lock()
	defer bus.Unlock()

	handlers, ok := bus.Handlers[t]

	if !ok {
		return
	}

	for i := 0; i < len(handlers); i++ {
		if handlers[i].Id == id {
			// Delete without preserving order
			handlers[i] = handlers[len(handlers)-1]
			handlers[len(handlers)-1] = Handler{}
			handlers = handlers[:len(handlers)-1]
		}
	}

	bus.Handlers[t] = handlers
}

// Fire sends an Event through an EventBus to all registered handlers for the
// EventType of the Event given.
func (bus *Bus) Fire(e Event) {
	bus.RLock()
	defer bus.RUnlock()

	handlers, ok := bus.Handlers[e.Type]
	if !ok {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(handlers))
	for _, handler := range handlers {
		go func(h Handler, e Event) {
			defer wg.Done()
			h.Callback(e)
		}(handler, e)
	}
	wg.Wait()
}
