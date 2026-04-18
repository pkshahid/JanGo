package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkshahid/JanGo/auth"
	"github.com/pkshahid/JanGo/auth/hashers"
	"github.com/pkshahid/JanGo/management"
	"github.com/pkshahid/JanGo/orm/queryset"
	"github.com/spf13/cobra"
)

func init() {
	management.Register(&CreateSuperuserCommand{})
}

type CreateSuperuserCommand struct{}

func (c *CreateSuperuserCommand) Name() string { return "createsuperuser" }
func (c *CreateSuperuserCommand) Help() string { return "Used to create a superuser." }
func (c *CreateSuperuserCommand) AddFlags(cmd *cobra.Command) {}

func (c *CreateSuperuserCommand) Execute(ctx context.Context, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email address: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Password (again): ")
	passwordConfirm, _ := reader.ReadString('\n')
	passwordConfirm = strings.TrimSpace(passwordConfirm)

	if password != passwordConfirm {
		return fmt.Errorf("Error: Your passwords didn't match.")
	}

	if len(password) < 8 {
		// Mock simple validation. Django checks similarity, commonness, etc.
		return fmt.Errorf("Error: Password must be at least 8 characters.")
	}

	encodedPassword := hashers.MakePassword(password)

	// Create user using ORM
	qs := queryset.NewQuerySet[auth.AbstractUser]()

	user := &auth.AbstractUser{
		UsernameStr:    username,
		EmailStr:       email,
		Password:       encodedPassword,
		IsActiveVal:    true,
		IsStaffVal:     true,
		IsSuperuserVal: true,
	}

	err := qs.Create(&user)
	if err != nil {
		return fmt.Errorf("Error: Could not create superuser: %v", err)
	}

	fmt.Println("Superuser created successfully.")
	return nil
}
