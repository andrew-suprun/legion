package mongo

import (
	"time"

	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/server"
)

type persistence struct {
}

func NewPersistence(connectString string, entityFactory server.EntityFactory) *persistence {
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
