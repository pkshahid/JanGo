package auth

import (
	"testing"
)

func TestAbstractUserModel(t *testing.T) {
	// Create permissions
	perm1 := &Permission{Name: "Can add user", Codename: "add_user", AppLabel: "auth"}
	perm2 := &Permission{Name: "Can view user", Codename: "view_user", AppLabel: "auth"}

	// Create group
	group := &Group{Name: "Editors", Permissions: []*Permission{perm2}}

	// Create user
	user := &AbstractUser{
		UsernameStr:     "testuser",
		IsActiveVal:     true,
		IsStaffVal:      false,
		IsSuperuserVal:  false,
		UserPermissions: []*Permission{perm1},
		Groups:          []*Group{group},
	}

	if user.Username() != "testuser" {
		t.Errorf("Expected username testuser, got %s", user.Username())
	}
	if !user.IsAuthenticated() {
		t.Errorf("Expected IsAuthenticated to be true")
	}
	if user.IsAnonymous() {
		t.Errorf("Expected IsAnonymous to be false")
	}

	// Test HasPerm
	if !user.HasPerm("auth.add_user") {
		t.Errorf("Expected user to have auth.add_user via direct perms")
	}
	if !user.HasPerm("auth.view_user") {
		t.Errorf("Expected user to have auth.view_user via group perms")
	}
	if user.HasPerm("auth.delete_user") {
		t.Errorf("User should not have auth.delete_user")
	}

	// Test HasPerms
	if !user.HasPerms([]string{"auth.add_user", "auth.view_user"}) {
		t.Errorf("Expected user to have multiple perms")
	}
	if user.HasPerms([]string{"auth.add_user", "auth.delete_user"}) {
		t.Errorf("Expected false for missing perm in list")
	}

	// Test HasModulePerm
	if !user.HasModulePerm("auth") {
		t.Errorf("Expected user to have module perm for auth")
	}
	if user.HasModulePerm("contenttypes") {
		t.Errorf("User should not have module perm for contenttypes")
	}

	// Test Superuser
	su := &AbstractUser{IsActiveVal: true, IsSuperuserVal: true}
	if !su.HasPerm("anything.any_perm") {
		t.Errorf("Superuser should have all perms")
	}
}
