package auth

import (
	"davidterranova/jurigen/backend/pkg/user"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

func GrantAnyAccess() func(authToken string) (user.User, error) {
	return func(authToken string) (user.User, error) {
		reqUsername, _, ok := parseBasicAuth(authToken)
		if !ok {
			return *user.NewUnauthenticated(), ErrUnauthorized
		}

		id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(reqUsername))
		return user.New(id, user.UserTypeAuthenticated), nil
	}
}

func BasicAuth(username string, password string) func(authToken string) (user.User, error) {
	return func(authToken string) (user.User, error) {
		reqUsername, reqPassword, ok := parseBasicAuth(authToken)
		if !ok {
			return user.NewUnauthenticated(), ErrUnauthorized
		}

		if reqUsername != username || reqPassword != password {
			return user.NewUnauthenticated(), ErrUnauthorized
		}

		id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(username))
		return user.New(id, user.UserTypeAuthenticated), nil
	}
}

func parseBasicAuth(auth string) (username string, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !equalFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
