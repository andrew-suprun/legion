package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
)

func TestInvalidCommand(t *testing.T) {
	s := New(testTimeService{}, &testPersistence{}, testCommandFactory)
	resultChan := s.Serve("conn", "invalid", nil)
	result := (<-resultChan).(*ServiceResult)
	if result.Failure == nil {
		fmt.Println("Unexpectedly succeeded.")
		t.Fatalf("Unexpectedly succeeded.")
	}
}

func TestValidCommand(t *testing.T) {
	s := New(testTimeService{}, &testPersistence{}, testCommandFactory)
	resultChan := s.Serve("conn", "valid", nil)
	result := (<-resultChan).(*ServiceResult)
	if result.Failure != nil {
		t.Fatalf("Unexpectedly failed.")
	}
}

type testTimeService struct{}

func (ts testTimeService) Now() time.Time {
	return time.Now()
}

type testPersistence struct{}

func (p *testPersistence) PersistEvent(event es.Event) {}

func (p *testPersistence) PersistEvents(events ...es.Event) {}

func (p *testPersistence) FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error) {
	return nil, nil
}

func (p *testPersistence) FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error) {
	return nil, nil
}

func testCommandFactory(cmdType es.CommandType, info es.Info) (Command, error) {
	if cmdType == "valid" {
		return testCommand{}, nil
	}
	return nil, errors.NewError(errors.Alert, InvalidCommand, "invalid")
}

type testCommand struct{}

func (testCommand) CommandType() es.CommandType {
	return "valid"
}

func (testCommand) Validate(helper CommandHelper) error {
	return nil
}

func (testCommand) Authorize(helper CommandHelper) error {
	return nil
}

func (testCommand) Handle(helper CommandHelper) error {
	return nil
}
