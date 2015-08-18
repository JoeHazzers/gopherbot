package event

import (
	"fmt"
	"reflect"
	"sync"
)

type Bus struct {
	sync.RWMutex
	callbacks map[string][]reflect.Value
}

func NewBus() *Bus {
	bus := Bus{
		callbacks: make(map[string][]reflect.Value),
	}

	return &bus
}

func (bus *Bus) Add(t string, f interface{}) error {
	err := validateCallback(f)
	if err != nil {
		return err
	}

	bus.Lock()
	defer bus.Unlock()

	callbacks, ok := bus.callbacks[t]
	if !ok {
		callbacks = make([]reflect.Value, 0)
	}

	bus.callbacks[t] = append(callbacks, reflect.ValueOf(f))

	return nil
}

func (bus *Bus) Delete(topic string, callback interface{}) (bool, error) {
	bus.Lock()
	defer bus.Unlock()
	return bus.del(topic, callback, false)
}

func (bus *Bus) DeleteAll(topic string, callback interface{}) (bool, error) {
	bus.Lock()
	defer bus.Unlock()
	return bus.del(topic, callback, true)
}

func (bus *Bus) Fire(t string, args ...interface{}) {
	bus.RLock()
	defer bus.RUnlock()

	if _, ok := bus.callbacks[t]; !ok {
		return
	}

	argVals := make([]reflect.Value, len(args))

	for i := 0; i < len(args); i++ {
		argVals[i] = reflect.ValueOf(args[i])
	}

	var wg sync.WaitGroup

	for _, c := range bus.callbacks[t] {
		wg.Add(1)
		go func(c reflect.Value) {
			defer wg.Done()
			c.Call(argVals)
		}(c)
	}

	wg.Wait()
}

func (bus *Bus) Purge(callback interface{}) (bool, error) {
	bus.Lock()
	defer bus.Unlock()

	found := false
	for topic := range bus.callbacks {
		del, err := bus.del(topic, callback, true)
		if del {
			found = true
		}
		if err != nil {
			return found, err
		}
	}
	return found, nil
}

func (bus *Bus) Reset() {
	bus.callbacks = make(map[string][]reflect.Value)
}

func (bus *Bus) del(t string, f interface{}, all bool) (bool, error) {
	err := validateCallback(f)
	if err != nil {
		return false, err
	}

	if _, ok := bus.callbacks[t]; !ok {
		return false, nil
	}

	v := reflect.ValueOf(f)

	found := false

	for i := 0; i < len(bus.callbacks[t]); i++ {
		if bus.callbacks[t][i] == v {
			bus.callbacks[t] = append(bus.callbacks[t][:i], bus.callbacks[t][i+1:]...)
			found = true
			if !all {
				break
			}
		}
	}

	if len(bus.callbacks[t]) == 0 {
		delete(bus.callbacks, t)
	}

	return found, nil
}

func validateCallback(f interface{}) error {
	if reflect.TypeOf(f).Kind() != reflect.Func {
		return fmt.Errorf("Provided callback is not a func")
	}

	return nil
}
