package event

import (
	"reflect"
	"strconv"
	"testing"
)

const (
	testNumCallbacks int = 10
	testNumTopics    int = 10
)

func testCallbackEmpty() {
	return
}

func testCallbackEmptyTwo() {
	return
}

func testNewAckCallback(c chan<- []reflect.Value) func(...interface{}) {
	return func(args ...interface{}) {
		argVals := make([]reflect.Value, len(args))
		for i := range argVals {
			argVals[i] = reflect.ValueOf(args[i])
		}

		c <- argVals
	}
}

func TestAddInvalid(t *testing.T) {
	bus := NewBus()

	err := bus.Add("test", "invalid")
	if err == nil {
		t.Errorf("Did not get error adding invalid callback")
	}
}

func TestAdd(t *testing.T) {
	bus := NewBus()

	topics := make([]string, testNumTopics)
	for i := 0; i < len(topics); i++ {
		topics[i] = strconv.Itoa(i)
		for j := 0; j < testNumCallbacks; j++ {
			err := bus.Add(topics[i], testCallbackEmpty)
			if err != nil {
				t.Errorf("Encountered error adding good callback: %+v", err)
			}
		}
	}

	v := reflect.ValueOf(testCallbackEmpty)

	for _, topic := range topics {
		callbacks, ok := bus.callbacks[topic]
		if !ok {
			t.Errorf("Topic %s not initialised")
		}

		if len(callbacks) != testNumCallbacks {
			t.Errorf("Expected %d callbacks, got %d", testNumCallbacks, len(callbacks))
		}

		for _, callback := range callbacks {
			if callback != v {
				t.Errorf("Expected callback %+v, got %+v", v, callback)
			}
		}
	}
}

func TestDelete(t *testing.T) {
	bus := NewBus()

	topics := make([]string, testNumTopics)
	for i := 0; i < len(topics); i++ {
		topics[i] = strconv.Itoa(i)
		for j := 0; j < testNumCallbacks; j++ {
			var callback interface{}
			if j%2 == 0 {
				callback = testCallbackEmpty
			} else {
				callback = testCallbackEmptyTwo
			}

			err := bus.Add(topics[i], callback)
			if err != nil {
				t.Errorf("Encountered error adding good callback: %+v", err)
			}
		}
	}

	v := reflect.ValueOf(testCallbackEmpty)

	for _, topic := range topics {
		for i := 0; i < testNumCallbacks; i++ {
			if i%2 == 0 {
				ok, err := bus.Delete(topic, testCallbackEmptyTwo)
				if !ok {
					t.Errorf("Callback reported as not deleted for topic %s", topic)
				}
				if err != nil {
					t.Errorf("Error deleting callback for topic %s: %+v", topic, err)
				}
			}
		}
	}

	for _, topic := range topics {
		ok, err := bus.Delete(topic, testCallbackEmptyTwo)
		if ok {
			t.Errorf("Callback reported as deleted for topic %s when none should exist", topic)
		}
		if err != nil {
			t.Errorf("Error deleting callback for topic %s: %+v", topic, err)
		}

		for i, callback := range bus.callbacks[topic] {
			if callback != v {
				t.Errorf("Wrong callback found for topic %s at index %d", topic, i)
			}
		}
	}
}

func TestFire(t *testing.T) {
	bus := NewBus()
	chans := make([]chan []reflect.Value, testNumCallbacks)

	for i := range chans {
		chans[i] = make(chan []reflect.Value)
		callback := testNewAckCallback(chans[i])
		bus.Add("test", callback)
	}

	args := []interface{}{1, "test", false}
	argVals := make([]reflect.Value, len(args))
	for i := range argVals {
		argVals[i] = reflect.ValueOf(args[i])
	}

	done := make(chan struct{})
	go func() {
		bus.Fire("test", args...)
		close(done)
	}()

	for _, c := range chans {
		res := <-c

		if len(res) != len(argVals) {
			t.Errorf("Expected %d args, got %d.", len(argVals), len(res))
		}

		for i := range argVals {
			if argVals[i] != res[i] {
				t.Errorf("Arg %d does match. Expected %+v, got %+v", i, argVals[i], res[i])
			}
		}
	}

	<-done
}
