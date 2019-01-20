package server

import (
	"context"

	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/json"
	"github.com/andrew-suprun/legion/tasks"

	"sync"
	"time"
)

const (
	ServerError    errors.ErrorCode = "server_error"
	InvalidCommand errors.ErrorCode = "invalid_command"
	DatabaseError  errors.ErrorCode = "database_error"
)

type Server struct {
	TimeService
	Persistence
	es.EntityFactory
}

type TimeService interface {
	Now() time.Time
}

type Persistence interface {
	PersistEvent(event es.Event)
	PersistEvents(events ...es.Event)
	FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error)
	FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error)
}

type Command interface {
	CommandType() es.CommandType
	Validate(helper CommandHelper) error
	Authorize(helper CommandHelper) error
	Handle(helper CommandHelper) error
}

type CommandHelper interface {
	Context() context.Context
	ConnectionId() es.EntityId
	CreateEntity(entity es.Entity)
	FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error)
	FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error)
	Reply(MessageType es.MessageType, info ...es.Info)
	AddDiagnostic(code errors.ErrorCode, desc string, info ...es.Info)

	// TODO: extract those two methods into separate services
	Now() time.Time
	SendMessage(message es.Message)
}

type ServiceResult struct {
	ConnectionId es.EntityId   `json:"connection_id"`
	Command      Command       `json:"command,omitempty"`
	Events       es.Events     `json:"events,omitempty"`
	Messages     es.Messages   `json:"messages,omitempty"`
	Diagnostics  errors.Errors `json:"diagnostics,omitempty"`
	Failure      error         `json:"failure,omitempty"`
	Panic        interface{}   `json:"panic,omitempty"`
}

func (r *ServiceResult) String() string {
	return json.Encode(r)
}

const (
	failure es.MessageType = "failure"
)

func New(
	ts TimeService,
	p Persistence,
	e es.EntityFactory,
) *Server {
	return &Server{
		TimeService:   ts,
		Persistence:   p,
		EntityFactory: e,
	}
}

func (s *Server) Shutdown() {
	// TODO:
}

func (s *Server) Serve(connId es.EntityId, cmd Command) (resultChan chan interface{}) {
	result := &ServiceResult{
		ConnectionId: connId,
		Command:      cmd,
	}

	activityResultChan := tasks.Start(
		func() interface{} {
			h := &commandHelper{
				timeService: s.TimeService,
				persistence: s.Persistence,
				entities:    map[es.EntityId]es.Entity{},
				result:      result,
			}
			h.result.Failure = cmd.Validate(h)
			if h.result.Failure != nil {
				return h.result
			}
			h.result.Failure = cmd.Authorize(h)
			if h.result.Failure != nil {
				return h.result
			}
			h.result.Failure = cmd.Handle(h)
			return h.result
		},
	)

	return tasks.Start(
		func() interface{} {
			value := <-activityResultChan
			switch v := value.(type) {
			case tasks.Panic:
				result.Panic = v.Value
			case error:
				result.Failure = v
			}

			return result
		},
	)
}

type commandHelper struct {
	lock        sync.Mutex
	ctx         context.Context
	timeService TimeService
	persistence Persistence
	result      *ServiceResult
	entities    map[es.EntityId]es.Entity
}

func (h *commandHelper) Now() time.Time {
	return h.timeService.Now()
}

func (h *commandHelper) Context() context.Context {
	return h.ctx
}

func (h *commandHelper) ConnectionId() es.EntityId {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.result.ConnectionId
}

func (h *commandHelper) CreateEntity(entity es.Entity) {
	h.lock.Lock()
	h.entities[entity.Id()] = entity
	h.lock.Unlock()
}

func (h *commandHelper) Reply(messageType es.MessageType, infos ...es.Info) {
	h.lock.Lock()
	h.result.Messages = append(h.result.Messages, es.Message{ConnectionId: h.result.ConnectionId, MessageType: messageType, Info: mergeInfos(infos...)})
	h.lock.Unlock()
}

func (h *commandHelper) SendMessage(message es.Message) {
	h.lock.Lock()
	h.result.Messages = append(h.result.Messages, message)
	h.lock.Unlock()
}

func (h *commandHelper) FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error) {
	h.lock.Lock()
	entity, ok := h.entities[id]
	h.lock.Unlock()
	if ok && entity.Type() == et {
		return entity, nil
	}
	entity, err := h.persistence.FetchEntity(et, id)
	if err != nil {
		return nil, err
	}
	h.lock.Lock()
	h.entities[id] = entity
	h.lock.Unlock()
	return entity, nil
}

func (h *commandHelper) FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error) {
	entity, err := h.persistence.FetchEntityAt(et, id, timestamp)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (h *commandHelper) AddDiagnostic(code errors.ErrorCode, desc string, info ...es.Info) {
	h.lock.Lock()
	h.result.Diagnostics = append(h.result.Diagnostics, errors.NewError(errors.Diagnostics, code, desc, info...))
	h.lock.Unlock()
}

func mergeInfo(this, other es.Info) {
	for k, v := range other {
		this[k] = v
	}
}

func mergeInfos(infos ...es.Info) es.Info {
	result := es.Info{}
	for _, v := range infos {
		mergeInfo(result, v)
	}
	return result
}
