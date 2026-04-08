package auth

import (
	"fmt"
	"strconv"

	godjangohttp "github.com/godjango/godjango/http"
)

func Login(req *godjangohttp.Request, user User) error {
	if req.Session == nil {
		return fmt.Errorf("session not found on request, ensure SessionMiddleware is active")
	}

	req.Session.Set("_auth_user_id", fmt.Sprintf("%d", user.ID()))
	req.Session.Set("_auth_user_name", user.Username())

	req.User = user

	return nil
}

func Logout(req *godjangohttp.Request) error {
	if req.Session == nil {
		return fmt.Errorf("session not found on request, ensure SessionMiddleware is active")
	}

	req.Session.Delete("_auth_user_id")
	req.Session.Delete("_auth_user_name")

	req.User = &AnonymousUser{}
	return nil
}

func GetUserFromRequest(req *godjangohttp.Request) User {
	if req.Session == nil {
		return &AnonymousUser{}
	}

	rawID, ok := req.Session.Get("_auth_user_id")
	if !ok {
		return &AnonymousUser{}
	}

	userIDStr, ok := rawID.(string)
	if !ok || userIDStr == "" {
		return &AnonymousUser{}
	}

	id, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return &AnonymousUser{}
	}

	user, err := GetUserByID(id)
	if err != nil {
		return &AnonymousUser{}
	}

	return user
}

func GetUserByID(id uint64) (User, error) {
	return defaultBackend.GetUser(id)
}
