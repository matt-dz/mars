// Package role contains utilities for user roles.
package role

import (
	"math"

	"mars/internal/database"
)

type Role int

const (
	RoleService Role = 200
	RoleAdmin   Role = 200
	RoleUser    Role = 100
	RoleUnknown Role = math.MinInt
)

func (r Role) String() string {
	if r == RoleAdmin || r == RoleService {
		return "admin"
	}
	if r == RoleUser {
		return "user"
	}
	return "unknown"
}

func ToRole(role string) Role {
	switch role {
	case "admin":
		return RoleAdmin
	case "user":
		return RoleUser
	case "service":
		return RoleService
	default:
		return RoleUnknown
	}
}

func DBToRole(role database.Role) Role {
	switch role {
	case database.RoleUser:
		return RoleUser
	case database.RoleAdmin:
		return RoleAdmin
	case database.RoleService:
		return RoleService
	default:
		return RoleUnknown
	}
}
