package tests

import (
	"fmt"

	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/json"
	"github.com/andrew-suprun/legion/server"
)

type Reply interface {
	Events() es.Events
	Messages() es.Messages
	Diagnostics() errors.Errors
	CheckSucceeded() Reply
	CheckFailed() Reply
	ValidateEvents(events ...es.Event) Reply
	ValidateMessages(expectedMsgs ...es.Message) Reply
	ValidateDiagnostics(errs ...errors.Error) Reply
	ValidateFailure(errs errors.Error) Reply
}

type reply struct {
	test       *Test
	resultChan chan interface{}
}

func (test *Test) newReply(resultChan chan interface{}) *reply {
	return &reply{
		test:       test,
		resultChan: resultChan,
	}
}

func (r *reply) GetResult() *server.ServiceResult {
	result := <-r.resultChan
	r.resultChan <- result
	return result.(*server.ServiceResult)
}

func (r *reply) Events() es.Events {
	result := r.GetResult()
	return result.Events
}

func (r *reply) Messages() es.Messages {
	result := r.GetResult()
	return result.Messages
}

func (r *reply) Diagnostics() errors.Errors {
	result := r.GetResult()
	return result.Diagnostics
}

func (r *reply) CheckSucceeded() Reply {
	result := r.GetResult()
	if result.Failure != nil {
		if failure, ok := result.Failure.(errors.Error); ok {
			r.test.FailWithResult(failure.Description, result)
		} else {
			r.test.FailWithResult("Command is expected to succeed.", result)
		}
	} else if result.Panic != nil {
		r.test.FailWithResult("Command is expected to succeed.", result)
	}
	return r
}

func (r *reply) CheckFailed() Reply {
	result := r.GetResult()
	if result.Failure == nil {
		r.test.FailWithResult("Command is expected to fail.", result)
	}
	return r
}

func (r *reply) ValidateEvents(events ...es.Event) Reply {
	return r
}

func (r *reply) ValidateMessages(expectedMsgs ...es.Message) Reply {
	receivedMessages := r.GetResult().Messages
	for _, expectedMsg := range expectedMsgs {
		r.validateMessage(expectedMsg, receivedMessages)
	}
	return r
}

func (r *reply) validateMessage(expectedMsg es.Message, receivedMessages es.Messages) Reply {
	fmt.Printf("### validateMessage:\n expected %s\n", json.Encode(expectedMsg))
	fmt.Printf("### validateMessage:\n received %s\n", json.Encode(receivedMessages))
	succeeded := false
	for _, receivedMsg := range receivedMessages {
		if receivedMsg.ConnectionId == expectedMsg.ConnectionId && receivedMsg.MessageType == expectedMsg.MessageType {
			validation := ValidateInfo(expectedMsg.Info, receivedMsg.Info)
			if validation.Succeeded() {
				succeeded = true
				break
			}
		}
	}
	if !succeeded {
		r.test.FailWithResult("Failed.", r.GetResult())
	}
	return r
}

func (r *reply) ValidateDiagnostics(errs ...errors.Error) Reply {
	return r
}

func (r *reply) ValidateFailure(failure errors.Error) Reply {
	return r
}

func (r *reply) String() string {
	return json.Encode(r)
}
