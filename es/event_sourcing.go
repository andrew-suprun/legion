package es

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/andrew-suprun/legion/json"
)

type EventId string
type CommandId string
type CommandType string

type Event struct {
	EventId     EventId     `json:"event_id" bson:"event_id"`
	CommandType CommandType `json:"command_type" bson:"command_type"`
	CommandId   CommandId   `json:"command_id" bson:"command_id"`
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

type Info = map[string]interface{}
type Infos []Info
