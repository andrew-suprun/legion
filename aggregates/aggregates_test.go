package aggregates

import (
	"legion/es"
	"legion/json"

	"log"
	"reflect"
	"testing"
)

func TestAggregate(t *testing.T) {
	entity := es.Info{
		"a":   13,
		"b":   14,
		"d":   15.0,
		"e":   16,
		"foo": "bar",
		"x": es.Info{
			"x1": "aaa",
			"x2": "bbb",
			"x4": "eee",
		},
		"y": "strinG",
		"z": es.Info{},
		"g": es.Info{},
	}
	event := es.Info{
		"a":   nil,
		"c":   42,
		"e":   16.5,
		"foo": "baz",
		"x": es.Info{
			"x2": "ccc",
			"x3": "ddd",
		},
		"y": es.Info{
			"FOO": "BAR",
		},
		"g": nil,
		"h": es.Info{},
	}
	Aggregate(entity, event)
	expected := es.Info{
		"b":   14,
		"c":   42,
		"d":   15.0,
		"e":   16.5,
		"foo": "baz",
		"h":   es.Info{},
		"x": es.Info{
			"x1": "aaa",
			"x2": "ccc",
			"x3": "ddd",
			"x4": "eee",
		},
		"y": es.Info{
			"FOO": "BAR",
		},
		"z": es.Info{},
	}

	if !reflect.DeepEqual(expected, entity) {
		log.Printf("Expected %s\n Got %s\n", json.Encode(expected), json.Encode(entity))
		t.Fail()
	}
}

func TestDiff(t *testing.T) {
	oldInfo := es.Info{
		"a":   13,
		"b":   14,
		"d":   15,
		"e":   16,
		"f":   17,
		"foo": "bar",
		"x": es.Info{
			"x1": "aaa",
			"x2": "bbb",
			"x4": "eee",
		},
		"y": "strinG",
		"z": es.Info{},
		"g": es.Info{},
	}
	newInfo := es.Info{
		"c":   42,
		"d":   nil,
		"e":   "aaa",
		"f":   17.5,
		"foo": "baz",
		"x": es.Info{
			"x1": "aaa",
			"x2": "ccc",
			"x3": "ddd",
		},
		"y": es.Info{
			"FOO": "BAR",
		},
		"z": es.Info{},
		"h": es.Info{},
	}
	diff := Diff(oldInfo, newInfo)
	expected := es.Info{
		"a":   nil,
		"b":   nil,
		"c":   42,
		"d":   nil,
		"e":   "aaa",
		"f":   17.5,
		"foo": "baz",
		"x": es.Info{
			"x2": "ccc",
			"x3": "ddd",
		},
		"y": es.Info{
			"FOO": "BAR",
		},
		"g": nil,
		"h": es.Info{},
	}

	if !reflect.DeepEqual(expected, diff) {
		log.Printf("Expected %s\n Got %s\n", json.Encode(expected), json.Encode(diff))
		t.Fail()
	}
}
