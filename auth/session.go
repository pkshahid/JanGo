package auth

import (
	"fmt"
	"strconv"

	godjangohttp "github.com/godjango/godjango/http"
)

// Login persists the authenticated user in the current session.
func Login(req *godjangohttp.Request, user User) error {
	if req.Session == nil {
		return fmt.Errorf("session not found on request, ensure SessionMiddleware is active")
	}

	req.Session.Set("_auth_user_id", fmt.Sprintf("%d", user.ID()))
	req.Session.Set("_auth_user_name", user.Username())

	// Update request user
	req.User = user

	return nil
}

// Logout clears the user from the current session.
func Logout(req *godjangohttp.Request) error {
	if req.Session == nil {
		return fmt.Errorf("session not found on request, ensure SessionMiddleware is active")
	}

	req.Session.Delete("_auth_user_id")
	req.Session.Delete("_auth_user_name")

	req.User = &AnonymousUser{}
	return nil
}

// GetUser retrieves the currently logged in user from the request session.
func GetUser(req *godjangohttp.Request) User {
	if req.Session == nil {
		return &AnonymousUser{}
	}

	userIDStr, ok := req.Session.Get("_auth_user_id").(string)
	if !ok || userIDStr == "" {
		return &AnonymousUser{}
	}

	id, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return &AnonymousUser{}
	}

	user, err := GetUserByID(id) // Call to backends implementation
	if err != nil {
		return &AnonymousUser{}
	}

	return user
}

// GetUserByID wraps backend calls dynamically
func GetUserByID(id uint64) (User, error) {
	return defaultBackend.GetUser(id)
}
