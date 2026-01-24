package openapi

import (
	"mars/internal/env"
)

type Server struct {
	Env *env.Env
}

var _ StrictServerInterface = (*Server)(nil)

func NewServer(env *env.Env) Server {
	return Server{
		Env: env,
	}
}
