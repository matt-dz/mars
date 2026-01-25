// Package env provides a way to access environmental dependencies
package env

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"mars/internal/database"
	marshttp "mars/internal/http"
	"mars/internal/log"
)

type envKeyType struct{}

var envKey envKeyType

// Env holds the dependencies for the environment.
type Env struct {
	Logger   *slog.Logger
	Database database.Querier
	HTTP     *marshttp.Client
	vars     map[string]string
}

func (e *Env) Get(key string) string {
	if e.vars == nil {
		e.vars = make(map[string]string)
	}
	return e.vars[key]
}

func (e *Env) Set(key, val string) {
	if e.vars == nil {
		e.vars = make(map[string]string)
	}
	e.vars[key] = val
}

func (e *Env) IsProd() bool {
	return strings.ToLower(os.Getenv("ENV")) == "production"
}

func New() *Env {
	return &Env{
		vars: make(map[string]string),
	}
}

// Null constructs a null instance.
func Null() *Env {
	return &Env{
		Logger:   log.NullLogger(),
		Database: nil,
		vars:     make(map[string]string),
	}
}

func WithContext(ctx context.Context, env *Env) context.Context {
	return context.WithValue(ctx, envKey, env)
}

func FromContext(ctx context.Context) *Env {
	env, ok := ctx.Value(envKey).(*Env)
	if !ok {
		return Null()
	}
	return env
}
