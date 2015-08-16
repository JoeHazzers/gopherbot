package event

import "sync"

// EventType distinguishes events from oneanother.
type Type int

// Event is an event passed fired through an EventBus and handled by
// EventHandlers./
type Event struct {
	Type    Type
	Payload interface{}
}

type HandleFunc func(Event)

// EventHandler is a func which consumes an Event.
type Handler struct {
	Name string
	Fun  HandleFunc
}

// EventBus dispatches fired Events to registered handlers for that EventType.
type Bus struct {
	sync.RWMutex
	Handlers map[Type][]Handler
}

// NewEventBus creates a new EventBus.
func NewBus() *Bus {
	bus := Bus{
		Handlers: make(map[Type][]Handler, 0),
	}

	return &bus
}

type ByName []Handler

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n ByName) Less(i, j int) bool { return n[i].Name < n[j].Name }

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
func (bus *Bus) DeleteHandler(t Type, name string) {
	bus.Lock()
	defer bus.Unlock()

	handlers, ok := bus.Handlers[t]

	if !ok {
		return
	}

	for i := 0; i < len(handlers); i++ {
		if handlers[i].Name == name {
			// Delete without preserving order
			handlers[i] = handlers[len(handlers)-1]
			handlers[len(handlers)-1] = Handler{}
			handlers = handlers[:len(handlers)-1]

			// we've removed an element, so we're closer to the end
			i++
		}
	}

	bus.Handlers[t] = handlers
}

// Fire sends an Event through an EventBus to all registered handlers for the
// EventType of the Event given.
func (bus *Bus) Fire(e Event) {
	handlers, ok := bus.Handlers[e.Type]
	if !ok {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(handlers))
	for _, handler := range handlers {
		go handler.Fun(e)
	}
	wg.Wait()
}
