package in_memory

import (
	"time"

	"github.com/andrew-suprun/legion/aggregates"
	"github.com/andrew-suprun/legion/errors"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/server"

	"github.com/reillywatson/goloose"
)

type persistence struct {
	entityFactory es.EntityFactory
	events        map[es.EntityType]map[es.EntityId]es.Events
}

func NewPersistence(entityFactory es.EntityFactory) server.Persistence {
	return &persistence{
		entityFactory: entityFactory,
		events:        map[es.EntityType]map[es.EntityId]es.Events{},
	}
}

func (p *persistence) PersistEvents(events ...es.Event) {
	for _, event := range events {
		p.PersistEvent(event)
	}
}

func (p *persistence) PersistEvent(event es.Event) {
	typeEvents, ok := p.events[event.EntityType]
	if !ok {
		typeEvents = map[es.EntityId]es.Events{}
		p.events[event.EntityType] = typeEvents
	}
	entityEvents := typeEvents[event.EntityId]
	entityEvents = append(entityEvents, event)
}

func (p *persistence) FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error) {
	return p.fetchEntity(et, id, func(es.Event) bool { return true })
}

func (p *persistence) FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error) {
	return p.fetchEntity(et, id, func(event es.Event) bool { return event.Timestamp.Before(timestamp) })
}

func (p *persistence) fetchEntity(et es.EntityType, id es.EntityId, filter func(es.Event) bool) (es.Entity, error) {
	entity := p.entityFactory(et)
	if entity == nil {
		panic(errors.NewError(errors.Alert, server.ServerError, "Unknown entity type.", es.Info{"entity_type": et}))
	}

	events := p.fetchEntityEvents(et, id)
	if len(events) == 0 {
		return nil, nil
	}
	aggr := es.Info{}
	for _, event := range events {
		if filter(event) {
			aggregates.Aggregate(aggr, event.Info)
		}
	}

	goloose.ToStruct(aggr, entity)

	return entity, nil
}

func (p *persistence) fetchEntityEvents(et es.EntityType, id es.EntityId) es.Events {
	if typeEvents, ok := p.events[et]; ok {
		return typeEvents[id]
	}
	return nil
}
