package errors

import (
	"bytes"
	"runtime/debug"
	"strings"

	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/json"
)

// TODO: Ability to chain errors
type Error struct {
	Severity    ErrorSeverity `json:"error_severity" bson:"error_severity"`
	Code        ErrorCode     `json:"error_code" bson:"error_code"`
	Description string        `json:"description" bson:"description"`
	Info        es.Info       `json:"info,omitempty" bson:"info,omitempty"`
	Trace       []string      `json:"stack_trace,omitempty" bson:"stack_trace,omitempty"`
}
type Errors []Error

type ErrorSeverity string

const (
	Diagnostics ErrorSeverity = "diagnostics" // errors that do not prevent service to be completed
	Failure     ErrorSeverity = "failure"     // service cannot be completed
	Alert       ErrorSeverity = "alert"       // authorities need to be alerted
)

type ErrorCode string

const (
	InvalidRequest ErrorCode = "invalid_request"
	Unauthorized   ErrorCode = "unauthorized"
)

func NewError(severity ErrorSeverity, code ErrorCode, desc string, info ...es.Info) Error {
	merged := es.Info{}
	for _, d := range info {
		for k, v := range d {
			merged[k] = v
		}
	}

	result := Error{
		Severity:    severity,
		Code:        code,
		Description: desc,
		Info:        merged,
	}

	if severity != Diagnostics {
		result.Trace = StackTrace()
	}

	return result
}

func (err Error) Error() string {
	return json.Encode(err)
}

func StackTrace() []string {
	var names []string
	var addresses []string
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	for i := 5; i < len(lines)-1; i++ {
		line := lines[i]
		name := strings.Split(line, "(0x")[0]
		address := strings.Split(strings.TrimSpace(lines[i+1]), " ")[0]

		names = append(names, name)
		addresses = append(addresses, address)

		i++
	}

	addressLen := 0
	for _, address := range addresses {
		if addressLen < len(address) {
			addressLen = len(address)
		}
	}

	result := make([]string, len(names))
	for i := range names {
		writer := &bytes.Buffer{}
		writer.WriteString(addresses[i])
		for j := len(addresses[i]); j <= addressLen; j++ {
			writer.WriteString(" ")
		}
		writer.WriteString(names[i])
		result[i] = writer.String()

	}

	return result
}
