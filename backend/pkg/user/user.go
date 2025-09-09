package user

import "github.com/google/uuid"

type UserType string

const (
	UserTypeSystem          UserType = "system"
	UserTypeAuthenticated   UserType = "authenticated"
	UserTypeUnauthenticated UserType = "unauthenticated"
)

type User interface {
	Id() uuid.UUID
	Type() UserType
}

func New(id uuid.UUID, userType UserType) User {
	switch userType {
	case UserTypeAuthenticated:
		return &UserAuthenticated{id: id}
	case UserTypeSystem:
		return &UserSystem{id: id}
	default:
		return &UserUnauthenticated{}
	}
}

func NewUnauthenticated() *UserAuthenticated {
	return &UserAuthenticated{id: uuid.Nil}
}

type UserAuthenticated struct {
	id uuid.UUID
}

func (u UserAuthenticated) Id() uuid.UUID {
	return u.id
}

func (u UserAuthenticated) Type() UserType {
	return UserTypeAuthenticated
}

type UserSystem struct {
	id uuid.UUID
}

func (u UserSystem) Id() uuid.UUID {
	return u.id
}

func (u UserSystem) Type() UserType {
	return UserTypeSystem
}

type UserUnauthenticated struct{}

func (u UserUnauthenticated) Id() uuid.UUID {
	return uuid.Nil
}

func (u UserUnauthenticated) Type() UserType {
	return UserTypeUnauthenticated
}
