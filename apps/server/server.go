package main

import (
	"fmt"
	"runtime/debug"

	"github.com/andrew-suprun/legion/json"

	"github.com/andrew-suprun/legion/errors"
)

func main() {

	fmt.Println(string(debug.Stack()))
	fmt.Println("----")
	fmt.Println(json.Encode(errors.StackTrace()))
}
