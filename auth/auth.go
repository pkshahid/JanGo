package auth

// User represents an authenticated user in the system.
type User interface {
	ID() uint64
	Username() string
	Email() string
	IsAuthenticated() bool
	IsAnonymous() bool
	IsActive() bool
	IsStaff() bool
	IsSuperuser() bool
	HasPerm(perm string) bool
	HasPerms(perms []string) bool
	HasModulePerm(appLabel string) bool
}

// AnonymousUser represents an unauthenticated user.
type AnonymousUser struct{}

func (u *AnonymousUser) ID() uint64                         { return 0 }
func (u *AnonymousUser) Username() string                   { return "" }
func (u *AnonymousUser) Email() string                      { return "" }
func (u *AnonymousUser) IsAuthenticated() bool              { return false }
func (u *AnonymousUser) IsAnonymous() bool                  { return true }
func (u *AnonymousUser) IsActive() bool                     { return false }
func (u *AnonymousUser) IsStaff() bool                      { return false }
func (u *AnonymousUser) IsSuperuser() bool                  { return false }
func (u *AnonymousUser) HasPerm(perm string) bool           { return false }
func (u *AnonymousUser) HasPerms(perms []string) bool       { return false }
func (u *AnonymousUser) HasModulePerm(appLabel string) bool { return false }
