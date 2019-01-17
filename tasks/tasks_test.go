package tasks

import (
	"errors"
	"fmt"
	"testing"
)

func TestRunNormal(t *testing.T) {
	future := Start(func() interface{} { return 7 })
	v := <-future
	if v != 7 {
		fmt.Println(v)
		t.FailNow()
	}
}

func TestRunWithError(t *testing.T) {
	fubar := errors.New("FUBAR")
	future := Start(func() interface{} { return fubar })
	v := <-future
	if v != fubar {
		fmt.Println(v)
		t.FailNow()
	}
}

func TestRunPanic(t *testing.T) {
	future := Start(func() interface{} { panic("FUBAR"); return 7 })
	v := <-future
	if _, ok := v.(Panic); !ok {
		fmt.Println(v)
		t.FailNow()
	}
}

func TestPanicInResultHandler(t *testing.T) {
	future := Start(func() interface{} {
		panic(42)
		return 7
	})

	Start(func() interface{} {
		v := <-future
		if v.(Panic).Value != 42 {
			fmt.Println(v)
			t.FailNow()
		}
		return 7
	})
}
