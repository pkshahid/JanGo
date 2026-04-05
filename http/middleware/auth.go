package middleware

import (
	godjangohttp "github.com/godjango/godjango/http"
)

type AnonymousUser struct{}

func (u *AnonymousUser) IsAuthenticated() bool { return false }
func (u *AnonymousUser) GetUsername() string   { return "" }

type AuthenticatedUser struct {
	Username string
}

func (u *AuthenticatedUser) IsAuthenticated() bool { return true }
func (u *AuthenticatedUser) GetUsername() string   { return u.Username }

// AuthenticationMiddleware attaches a User to the request.
// It depends on SessionMiddleware being executed first.
func AuthenticationMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.Session != nil {
			userID := req.Session.Get("_auth_user_id")
			if userID != nil {
				// In a full implementation, this would fetch the user from the DB.
				// For prototype purposes, we attach a generic AuthenticatedUser
				username, _ := req.Session.Get("_auth_user_name").(string)
				req.User = &AuthenticatedUser{Username: username}
			} else {
				req.User = &AnonymousUser{}
			}
		} else {
			req.User = &AnonymousUser{}
		}

		return next(req)
	}
}
