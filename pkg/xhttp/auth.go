package xhttp

import (
	"davidterranova/jurigen/pkg/auth"
	"fmt"
	"net/http"

	"davidterranova/jurigen/pkg/user"
)

type AuthFn func(r *http.Request) (user.User, error)

func AuthMiddleware(authFn AuthFn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, err := authFn(r)
			if err != nil {
				WriteError(ctx, w, http.StatusUnauthorized, "unauthorized", err)
				return
			}

			reqWithCtx := r.WithContext(auth.ContextWithUser(ctx, user))
			next.ServeHTTP(w, reqWithCtx)
		})
	}
}

func BasicAuthFn(username string, password string) AuthFn {
	return func(r *http.Request) (user.User, error) {
		user, err := auth.BasicAuth(username, password)(r.Header.Get("Authorization"))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", auth.ErrUnauthorized, err.Error())
		}

		return user, nil
	}
}

func GrantAnyFn() AuthFn {
	return func(r *http.Request) (user.User, error) {
		user, err := auth.GrantAnyAccess()(r.Header.Get("Authorization"))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", auth.ErrUnauthorized, err.Error())
		}

		return user, nil
	}
}
