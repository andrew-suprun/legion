package es

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"legion/json"
	"time"
)

type CommandType string
type Command interface {
	Validate(helper CommandHelper) error
	Authorize(helper CommandHelper) error
	Handle(helper CommandHelper) error
}
type CommandFactory func(CommandType) Command
type CommandHelper interface {
	Context() context.Context
	ConnectionId() EntityId
	CreateEntity(entity Entity)
	FetchEntity(et EntityType, id EntityId) (Entity, error)
	FetchEntityAt(et EntityType, id EntityId, timestamp time.Time) (Entity, error)
	Reply(MessageType MessageType, info ...Info)
	AddDiagnostic(code ErrorCode, desc string, info ...Info)

	// TODO: extract those two methods into separate services
	Now() time.Time
	SendMessage(message Message)
}

type EventId string
type Event struct {
	EventId     EventId     `json:"event_id" bson:"event_id"`
	CommandType CommandType `json:"command_type" bson:"command_type"`
	CommandId   EntityId    `json:"command_id" bson:"command_id"`
	EntityType  EntityType  `json:"entity_type" bson:"entity_type"`
	EntityId    EntityId    `json:"entity_id" bson:"entity_id"`
	Timestamp   time.Time   `json:"timestamp" bson:"timestamp"`
	Info        Info        `json:"info,omitempty" bson:"info,omitempty"`
}

func (e Event) String() string {
	return json.Encode(e)
}

type Events []Event

func (e Events) String() string {
	return json.Encode(e)
}

type EntityType string
type EntityId string
type Entity interface {
	Id() EntityId
	Type() EntityType
}
type Entities []Entity
type EntityFactory func(et EntityType) Entity

func NewEntityId() EntityId {
	var buf [15]byte
	rand.Read(buf[:])
	return EntityId(base64.URLEncoding.EncodeToString(buf[:]))
}

type MessageType string
type Message struct {
	ConnectionId EntityId    `json:"connection_id"`
	MessageType  MessageType `json:"message_type"`
	Info         Info        `json:"info,omitempty"`
}

func (m Message) String() string {
	return json.Encode(m)
}

type Messages []Message

func (m Messages) String() string {
	return json.Encode(m)
}

type ErrorCode string

type Info = map[string]interface{}
type Infos []Info
