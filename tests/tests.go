package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/json"
	"github.com/andrew-suprun/legion/persistence/in_memory"
	"github.com/andrew-suprun/legion/persistence/mongo"
	"github.com/andrew-suprun/legion/server"
)

type Test struct {
	*testing.T
	server.TimeService
	server.Persistence
	*server.Server
}

type ValueValidator interface {
	Valid(value interface{}) bool
}

type TestFailure struct {
	Message    string                `json:"message,omitempty"`
	Result     *server.ServiceResult `json:"result,omitempty"`
	Info       es.Info               `json:"info,omitempty"`
	StackTrace []string              `json:"stack_trace,omitempty"`
}

func (f TestFailure) Error() string {
	return json.Encode(f)
}

func NewTest(
	t *testing.T,
	commandFactory server.CommandFactory,
	entityFactory server.EntityFactory,
) *Test {
	mongoConnectString := os.Getenv("LEGION_MONGO")
	ts := &testTimeService{}
	p := in_memory.NewPersistence(entityFactory)
	if mongoConnectString != "" {
		p = mongo.NewPersistence(mongoConnectString, entityFactory)
	}
	serv := server.New(
		ts,
		p,
		commandFactory,
	)

	return &Test{
		T:           t,
		TimeService: ts,
		Persistence: p,
		Server:      serv,
	}
}

func (t *Test) Send(connId es.EntityId, cmdType es.CommandType, cmdInfo es.Info) Reply {
	resultChan := t.Server.Serve(connId, cmdType, cmdInfo)
	return t.newReply(resultChan)
}

type testTimeService struct {
	timeshift time.Duration
}

func (t *testTimeService) SetNow(timestamp time.Time) {
	t.timeshift = time.Until(timestamp)
}

func (t *testTimeService) Now() time.Time {
	return time.Now().UTC().Add(t.timeshift)
}

func (t *Test) FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error) {
	return t.Persistence.FetchEntity(et, id)
}

func (t *Test) FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error) {
	return t.Persistence.FetchEntityAt(et, id, timestamp)
}

func (t *Test) Fail(message string, info es.Info) {
	fmt.Printf("FAIL: %s\n", message)
	fmt.Printf("info: %s\n", json.Encode(info))
	trace := errors.StackTrace()
	fmt.Println("stack trace:")
	for _, line := range trace {
		fmt.Println(line)
	}
	t.T.FailNow()
}

func (t *Test) FailWithResult(message string, result *server.ServiceResult) {
	fmt.Println("================================================================================")
	fmt.Printf("FAIL: %s\n", message)
	if result.Command != nil {
		fmt.Printf("\ncommand type: %q\n", result.Command.CommandType())
		fmt.Printf("\ncommand: %s\n", json.Encode(result.Command))
	}
	fmt.Printf("\nconnection: %q\n", result.ConnectionId)
	if len(result.Events) > 0 {
		fmt.Printf("\nevents: %s\n", json.Encode(result.Events))
	}
	if len(result.Messages) > 0 {
		fmt.Printf("\nmessages: %s\n", json.Encode(result.Messages))
	}
	if result.Panic != nil {
		fmt.Printf("\npanic: %s\n", json.Encode(result.Panic))
	}
	if result.Failure != nil {
		fmt.Printf("\nfailure: %s\n", result.Failure.Error())
	}

	trace := errors.StackTrace()
	fmt.Println("\nstack trace:")
	for _, line := range trace {
		fmt.Println(line)
	}
	fmt.Println("================================================================================")
	t.T.FailNow()
}
