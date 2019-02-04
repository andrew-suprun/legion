package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/andrew-suprun/legion/es"
)

func TestInfoWithString(t *testing.T) {
	result := ValidateInfo(
		es.Info{
			"foo": "bar",
		}, es.Info{
			"foo": "bar",
		})
	if !result.Succeeded() {
		fmt.Printf("### logs\n%s\n", result)
		t.FailNow()
	}
}

func TestNestedInfo(t *testing.T) {
	result := ValidateInfo(es.Info{
		"foo": es.Info{
			"bar": time.Now(),
		},
	}, es.Info{
		"foo": es.Info{
			"bar": time.Now().Format(time.RFC3339Nano),
		},
	})
	if !result.Succeeded() {
		fmt.Printf("### logs\n%s\n", result)
		t.FailNow()
	}
}
