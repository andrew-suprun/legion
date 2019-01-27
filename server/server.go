package server

import (
	"context"

	"github.com/andrew-suprun/legion/aggregates"

	"github.com/reillywatson/goloose"

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
	EntityFactory
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

type EntityFactory func(et es.EntityType, id es.EntityId) es.Entity

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
	CommandId    es.EntityId   `json:"command_id"`
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
	e EntityFactory,
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
		CommandId:    es.NewEntityId(),
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
			h.createEventsFromEntities()
			h.persistence.PersistEvents(h.result.Events...)
			return h.result
		},
	)

	return tasks.Start(
		func() interface{} {
			value := <-activityResultChan
			switch v := value.(type) {
			case tasks.Panic:
				result.Panic = v.Value
				result.Failure = v.Err
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
	entityData  map[es.EntityId]es.Info
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
	h.entities[entity.EntityId()] = entity
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
	if ok && entity.EntityType() == et {
		return entity, nil
	}
	entity, err := h.persistence.FetchEntity(et, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	h.lock.Lock()
	h.entities[id] = entity
	var data es.Info
	goloose.ToStruct(entity, &data)
	h.entityData[id] = data
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

func (h *commandHelper) createEventsFromEntities() {
	for _, entity := range h.entities {
		originalData := h.entityData[entity.EntityId()]
		var updatedData es.Info
		goloose.ToStruct(entity, &updatedData)
		diff := aggregates.Diff(originalData, updatedData)
		h.result.Events = append(h.result.Events, es.Event{
			EventId:     es.NewEventId(),
			CommandType: h.result.Command.CommandType(),
			CommandId:   h.result.CommandId,
			EntityType:  entity.EntityType(),
			EntityId:    entity.EntityId(),
			Timestamp:   h.timeService.Now(),
			Info:        diff,
		})
	}
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
