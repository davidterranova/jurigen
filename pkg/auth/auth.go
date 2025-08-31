package auth

import (
	"context"
	"errors"

	"davidterranova/jurigen/pkg/user"
)

type RequestCtxKey string

const RequestCtxUserKey RequestCtxKey = "user"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUnauthorized = errors.New("unauthorized")
)

func UserFromContext(ctx context.Context) (user.User, error) {
	u, ok := ctx.Value(RequestCtxUserKey).(user.User)
	if !ok {
		return user.NewUnauthenticated(), ErrUserNotFound
	}

	return u, nil
}

func ContextWithUser(ctx context.Context, u user.User) context.Context {
	return context.WithValue(ctx, RequestCtxUserKey, u)
}
