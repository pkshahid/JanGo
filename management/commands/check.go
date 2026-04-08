package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/management"
	"github.com/spf13/cobra"
)

type CheckCmd struct{}

func (c *CheckCmd) Name() string {
	return "check"
}

func (c *CheckCmd) Help() string {
	return "Inspects the entire GoDjango project for common security and configuration issues."
}

func (c *CheckCmd) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("deploy", false, "Check deployment readiness")
}

func (c *CheckCmd) Execute(ctx context.Context, args []string) error {
	// A hack to parse the --deploy flag manually since we're called via Execute dynamically in tests or via generic interface
	deploy := false
	for _, arg := range os.Args {
		if arg == "--deploy" {
			deploy = true
			break
		}
	}

	var errors []string
	var warnings []string

	s := settings.Get()

	if deploy {
		if s.DEBUG {
			errors = append(errors, "DEBUG is set to true in deployment.")
		}

		if s.SECRET_KEY == "" || len(s.SECRET_KEY) < 50 || s.SECRET_KEY == "django-insecure-default-key" {
			errors = append(errors, "SECRET_KEY is either missing, insecure, or too short (needs 50 chars).")
		}

		if len(s.ALLOWED_HOSTS) == 0 {
			errors = append(errors, "ALLOWED_HOSTS must not be empty in deployment.")
		}

		if !s.SECURE_SSL_REDIRECT {
			warnings = append(warnings, "SECURE_SSL_REDIRECT is not set to true. Your site might not force HTTPS.")
		}

		if !s.SESSION_COOKIE_SECURE {
			warnings = append(warnings, "SESSION_COOKIE_SECURE is not set to true.")
		}

		if !s.SESSION_COOKIE_HTTPONLY {
			warnings = append(warnings, "SESSION_COOKIE_HTTPONLY is not set to true.")
		}

		// Assume CSRF_COOKIE_SECURE exists in settings or warn
		// Using a generic warning since it's not explicitly in settings struct but requested
		warnings = append(warnings, "Ensure CSRF_COOKIE_SECURE is true.")

		if s.SECURE_HSTS_SECONDS <= 0 {
			warnings = append(warnings, "SECURE_HSTS_SECONDS is 0 or not set. HSTS will not be enabled.")
		}

		if s.X_FRAME_OPTIONS != "DENY" && s.X_FRAME_OPTIONS != "SAMEORIGIN" {
			warnings = append(warnings, "X_FRAME_OPTIONS is not DENY or SAMEORIGIN. Risk of clickjacking.")
		}

		// Checking hashers is usually done via looking at the AUTH_PASSWORD_VALIDATORS or hashers config.
		// For now we will warn if we can't confirm argon2id or bcrypt.
		warnings = append(warnings, "Ensure your default password hasher is argon2id or bcrypt.")
	} else {
		if s.SECRET_KEY == "" {
			errors = append(errors, "SECRET_KEY is empty.")
		}
	}

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("System check identified no issues (0 silenced).")
		return nil
	}

	fmt.Println("System check identified some issues:")
	fmt.Println()

	if len(warnings) > 0 {
		fmt.Println("WARNINGS:")
		for _, w := range warnings {
			fmt.Printf("- %s\n", w)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Println("ERRORS:")
		for _, e := range errors {
			fmt.Printf("- %s\n", e)
		}
		fmt.Println()
		return fmt.Errorf("system check found %d errors", len(errors))
	}

	return nil
}

func init() {
	management.Register(&CheckCmd{})
}
