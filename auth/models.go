package auth

import (
	"strings"
	"time"

	"github.com/pkshahid/JanGo/orm"
)

func init() {
	orm.Register(&Permission{}, &Group{}, &AbstractUser{})
}

// Permission represents a specific action (e.g. 'add_user', 'change_group')
type Permission struct {
	orm.Model
	Name     string `gd:"CharField,max_length=255"`
	Codename string `gd:"CharField,max_length=100"`
	AppLabel string `gd:"CharField,max_length=100"`
}

func (m *Permission) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable: "auth_permission",
		UniqueTogether: [][]string{{"Codename", "AppLabel"}},
	}
}

// Group represents a collection of users with shared permissions.
type Group struct {
	orm.Model
	Name        string        `gd:"CharField,max_length=150,unique=true"`
	Permissions []*Permission `gd:"ManyToManyField,to=auth.Permission"`
}

func (m *Group) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable: "auth_group",
	}
}

// AbstractUser provides the core fields and methods for a User model.
type AbstractUser struct {
	orm.Model
	UsernameStr string    `gd:"CharField,max_length=150,unique=true"`
	EmailStr    string    `gd:"EmailField,max_length=254,blank=true"`
	Password    string    `gd:"CharField,max_length=128"`
	FirstName   string    `gd:"CharField,max_length=150,blank=true"`
	LastName    string    `gd:"CharField,max_length=150,blank=true"`
	IsActiveVal bool      `gd:"BooleanField,default=true"`
	IsStaffVal  bool      `gd:"BooleanField,default=false"`
	IsSuperuserVal bool   `gd:"BooleanField,default=false"`
	DateJoined  time.Time `gd:"DateTimeField,auto_now_add=true"`
	LastLogin   time.Time `gd:"DateTimeField,null=true,blank=true"`

	Groups          []*Group      `gd:"ManyToManyField,to=auth.Group,related_name=user_set"`
	UserPermissions []*Permission `gd:"ManyToManyField,to=auth.Permission,related_name=user_set"`
}

func (m *AbstractUser) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable: "auth_user",
	}
}

// Implement User interface for AbstractUser
func (u *AbstractUser) ID() uint64 {
	return u.Model.ID
}

func (u *AbstractUser) Username() string {
	return u.UsernameStr
}

func (u *AbstractUser) Email() string {
	return u.EmailStr
}

func (u *AbstractUser) IsAuthenticated() bool {
	return true
}

func (u *AbstractUser) IsAnonymous() bool {
	return false
}

func (u *AbstractUser) IsActive() bool {
	return u.IsActiveVal
}

func (u *AbstractUser) IsStaff() bool {
	return u.IsStaffVal
}

func (u *AbstractUser) IsSuperuser() bool {
	return u.IsSuperuserVal
}

// HasPerm checks if the user has a specific permission.
// Expected format: "app_label.codename"
func (u *AbstractUser) HasPerm(perm string) bool {
	if !u.IsActive() {
		return false
	}
	if u.IsSuperuser() {
		return true
	}

	parts := strings.Split(perm, ".")
	if len(parts) != 2 {
		return false // Invalid format
	}

	appLabel, codename := parts[0], parts[1]

	// Check direct user permissions
	for _, p := range u.UserPermissions {
		if p.AppLabel == appLabel && p.Codename == codename {
			return true
		}
	}

	// Check group permissions
	for _, g := range u.Groups {
		for _, p := range g.Permissions {
			if p.AppLabel == appLabel && p.Codename == codename {
				return true
			}
		}
	}

	return false
}

func (u *AbstractUser) HasPerms(perms []string) bool {
	for _, p := range perms {
		if !u.HasPerm(p) {
			return false
		}
	}
	return true
}

func (u *AbstractUser) HasModulePerm(appLabel string) bool {
	if !u.IsActive() {
		return false
	}
	if u.IsSuperuser() {
		return true
	}

	// Direct perms
	for _, p := range u.UserPermissions {
		if p.AppLabel == appLabel {
			return true
		}
	}

	// Group perms
	for _, g := range u.Groups {
		for _, p := range g.Permissions {
			if p.AppLabel == appLabel {
				return true
			}
		}
	}

	return false
}
