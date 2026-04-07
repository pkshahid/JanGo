package auth

import (
	"fmt"

	"github.com/godjango/godjango/auth/hashers"
	"github.com/godjango/godjango/orm/queryset"
)

// AuthBackend represents the interface for authentication mechanisms.
type AuthBackend interface {
	Authenticate(username, password string) (User, error)
	GetUser(id uint64) (User, error)
	HasPerm(user User, perm string) bool
}

// ModelBackend authenticates against the default User model via ORM.
type ModelBackend struct{}

// Authenticate verifies the username and password against the AbstractUser model.
func (b *ModelBackend) Authenticate(username, password string) (User, error) {
	qs := queryset.NewQuerySet[AbstractUser]()

	// Django authenticates active users by default
	user, err := qs.Get(queryset.Lookup{"UsernameStr__exact": username})
	if err != nil {
		return nil, fmt.Errorf("user not found or inactive: %w", err)
	}

	if !user.IsActiveVal {
		return nil, fmt.Errorf("user is inactive")
	}

	if hashers.CheckPassword(password, user.Password) {
		// Needs to return the pointer so it implements User interface
		return &user, nil
	}

	return nil, fmt.Errorf("invalid password")
}

// GetUser retrieves a user by their ID.
func (b *ModelBackend) GetUser(id uint64) (User, error) {
	qs := queryset.NewQuerySet[AbstractUser]()
	// In a real framework, we'd use select_related / prefetch_related for groups and perms.
	// Since we're using a simplified query executor, we just fetch the core model.
	user, err := qs.Get(queryset.Lookup{"ID__exact": id})
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// HasPerm delegates to the user model's HasPerm method.
func (b *ModelBackend) HasPerm(user User, perm string) bool {
	return user.HasPerm(perm)
}

// Global authentication logic

var defaultBackend = &ModelBackend{}

// Authenticate iterates over configured backends (only ModelBackend in this prototype).
func Authenticate(username, password string) (User, error) {
	return defaultBackend.Authenticate(username, password)
}

// GetUser retrieves a user from the configured backends.
func GetUser(id uint64) (User, error) {
	return defaultBackend.GetUser(id)
}
