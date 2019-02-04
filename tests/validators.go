package tests

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
)

type ValidationResult interface {
	Succeeded() bool
}

func ValidateInfo(expected, received es.Info) ValidationResult {
	var infoLogs logs
	validateInfo(0, "", expected, received, &infoLogs)
	return infoLogs
}

type log struct {
	level  int
	key    string
	failed bool
	value  string
}

type logs []log

func (l logs) Succeeded() bool {
	for _, log := range l {
		if log.failed {
			return false
		}
	}
	return true
}

func (l logs) String() string {
	buf := &bytes.Buffer{}

	for _, log := range l {
		if log.failed {
			fmt.Fprint(buf, "ERR ")
			for i := 0; i < log.level; i++ {
				fmt.Fprint(buf, "    ")
			}
		} else {
			fmt.Fprint(buf, "    ")
			for i := 0; i < log.level; i++ {
				fmt.Fprint(buf, "    ")
			}
		}
		if log.key != "" {
			fmt.Fprintf(buf, "%q: ", log.key)
		}
		fmt.Fprintln(buf, log.value)
	}
	return buf.String()
}

func validateInfo(level int, key string, expected es.Info, received interface{}, logs *logs) {
	openLog := log{
		level: level,
		key:   key,
		value: "{",
	}
	*logs = append(*logs, openLog)

	rInfo, ok := received.(es.Info)
	if !ok {
		openLog.failed = true
	}

	for k, eValue := range expected {
		rValue := rInfo[k]
		switch typedEValue := eValue.(type) {
		case string:
			validateString(level+1, k, typedEValue, rValue, logs)
		case int:
			validateInt(level+1, k, typedEValue, rValue, logs)
		case float64:
			validateFloat(level+1, k, typedEValue, rValue, logs)
		case time.Time:
			validateTime(level+1, k, typedEValue, rValue, logs)
		case es.Info:
			validateInfo(level+1, k, typedEValue, rValue, logs)
		case []interface{}:
			panic(errors.NewError(errors.Alert, "PANIC", "Slices are not implemented yet."))
		case ValueValidator:
			validateValue(level+1, k, typedEValue, rValue, logs)
		default:
			panic(errors.NewError(errors.Alert, "PANIC", "cannot validate value of unsupported type", es.Info{
				"type": fmt.Sprintf("%T", eValue),
			}))
		}
	}
	*logs = append(*logs, log{
		level: level,
		value: "}",
	})
}

func validateString(level int, key string, expected string, received interface{}, logs *logs) {
	logRecord := log{
		level: level,
		key:   key,
		value: fmt.Sprintf("%q", expected),
	}
	rString, ok := received.(string)
	if !ok || rString != expected {
		logRecord.failed = true
	}
	*logs = append(*logs, logRecord)
}

func validateInt(level int, key string, expected int, received interface{}, logs *logs) {
	logRecord := log{
		level: level,
		key:   key,
		value: fmt.Sprintf("%v", expected),
	}
	rFloat, ok := received.(float64)
	if !ok || math.Abs(rFloat-float64(expected)) > 0.000001 {
		logRecord.failed = true
	}
	*logs = append(*logs, logRecord)
}

func validateFloat(level int, key string, expected float64, received interface{}, logs *logs) {
	logRecord := log{
		level: level,
		key:   key,
		value: fmt.Sprintf("%v", expected),
	}
	rFloat, ok := received.(float64)
	if !ok || math.Abs(rFloat-expected) > 0.000001 {
		logRecord.failed = true
	}
	*logs = append(*logs, logRecord)
}

func validateTime(level int, key string, expected time.Time, received interface{}, logs *logs) {
	logRecord := log{
		level: level,
		key:   key,
		value: fmt.Sprintf("%v", expected),
	}
	rTimeString, ok := received.(string)
	if ok {
		rTime, err := time.Parse(time.RFC3339, rTimeString)
		if err != nil {
			logRecord.failed = true
		}
		timeDiff := rTime.Sub(expected)
		if timeDiff > 5*time.Second || timeDiff < -5*time.Second {
			logRecord.failed = true
		}
	} else {
		logRecord.failed = true
	}
	*logs = append(*logs, logRecord)
}

func validateValue(level int, key string, expected ValueValidator, received interface{}, logs *logs) {
	logRecord := log{
		level:  level,
		key:    key,
		value:  fmt.Sprintf("%v", expected),
		failed: !expected.Valid(received),
	}
	*logs = append(*logs, logRecord)
}
