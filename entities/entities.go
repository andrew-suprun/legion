package entities

import (
	"github.com/andrew-suprun/legion/connections"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/users"
)

var factories = map[es.EntityType]func() es.Entity{
	users.UserCredentialsEntityType:  func() es.Entity { return &users.UserCredentials{} },
	connections.ConnectionEntityType: func() es.Entity { return &connections.Connection{} },
}

func Factory(et es.EntityType) es.Entity {
	factory := factories[et]
	if factory == nil {
		return nil
	}
	return factory()
}
