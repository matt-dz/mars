// Package role contains utilities for user roles.
package role

import (
	"math"
)

type Role int

const (
	RoleAdmin   Role = 200
	RoleUser    Role = 100
	RoleUnknown Role = math.MinInt
)

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleUser:
		return "user"
	default:
		return "unknown"
	}
}

func ToRole(role string) Role {
	switch role {
	case "admin":
		return RoleAdmin
	case "user":
		return RoleUser
	default:
		return RoleUnknown
	}
}
