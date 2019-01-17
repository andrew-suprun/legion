package commands

import (
	"github.com/andrew-suprun/legion/connections"
	"github.com/andrew-suprun/legion/es"
	"github.com/andrew-suprun/legion/users"
)

// TODO: Refactor
var factories = map[es.CommandType]func() es.Command{
	connections.ConfigureConnection: func() es.Command { return &connections.ConfigureConnectionCommand{} },
	users.CreateUserCredentials:     func() es.Command { return &users.CreateUserCredentialsCommand{} },
}

func Factory(cmdType es.CommandType) es.Command {
	factory := factories[cmdType]
	if factory == nil {
		return nil
	}
	return factory()
}
