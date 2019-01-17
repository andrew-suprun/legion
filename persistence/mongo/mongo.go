package mongo

import (
	"legion/es"
	"time"
)

type persistence struct {
}

func NewPersistence(connectString string, entityFactory es.EntityFactory) *persistence {
	return &persistence{}
}

func (env *persistence) PersistEvents(events ...es.Event) {
}

func (env *persistence) PersistEvent(event es.Event) {
}

func (env *persistence) FetchEntity(et es.EntityType, id es.EntityId) (es.Entity, error) {
	return nil, nil
}

func (env *persistence) FetchEntityAt(et es.EntityType, id es.EntityId, timestamp time.Time) (es.Entity, error) {
	return nil, nil
}
