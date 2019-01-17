package errors

import (
	"bytes"
	"legion/es"
	"legion/json"
	"os"
	"runtime/debug"
	"strings"
)

// TODO: Ability to chain errors
type Error struct {
	Severity    ErrorSeverity `json:"error_severity" bson:"error_severity"`
	Code        es.ErrorCode  `json:"error_code" bson:"error_code"`
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

const (
	InvalidRequest es.ErrorCode = "invalid_request"
	Unauthorized   es.ErrorCode = "unauthorized"
)

func NewError(severity ErrorSeverity, code es.ErrorCode, desc string, info ...es.Info) Error {
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
	projectPath, _ := os.Getwd()
	for len(projectPath) > 0 {
		if _, err := os.Stat(projectPath + "/go.mod"); !os.IsNotExist(err) {
			break
		}
		projectPath = projectPath[:strings.LastIndex(projectPath, "/")]
	}

	projectPathLen := len(projectPath) + 1
	var names []string
	var addresses []string
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	for i := 0; i < len(lines)-1; i++ {
		line := lines[i]
		if strings.HasPrefix(line, "legion/") || strings.HasPrefix(line, "main.") {
			name := strings.Split(line, "(0")[0]
			address := strings.Split(strings.TrimSpace(lines[i+1]), " ")[0]

			if strings.HasPrefix(name, "legion/errors.") || strings.HasPrefix(name, "legion/server.") {
				// skip
			} else {
				names = append(names, name)
				addresses = append(addresses, address)
			}

			i++
		}
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

	for i := range result {
		result[i] = result[i][projectPathLen:]
	}

	return result
}
