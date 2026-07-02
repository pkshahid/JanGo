package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/project/* templates/project/project/* templates/app/*
var templateFS embed.FS

var version = "0.0.1" // Set the framework version here

func main() {
	var rootCmd = &cobra.Command{
		Use:   "godjango-admin",
		Short: "GoDjango administrative tool",
	}

	var startProjectCmd = &cobra.Command{
		Use:   "startproject [name]",
		Short: "Scaffolds a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			return createProject(projectName)
		},
	}

	var startAppCmd = &cobra.Command{
		Use:   "startapp [name]",
		Short: "Scaffolds a new app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			return createApp(appName)
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints framework version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("GoDjango version", version)
		},
	}

	rootCmd.AddCommand(startProjectCmd)
	rootCmd.AddCommand(startAppCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createProject(projectName string) error {
	// Create project root
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// manage.go
	if err := renderTemplate("templates/project/manage.go.tmpl", filepath.Join(projectName, "manage.go"), map[string]string{"ProjectName": projectName}); err != nil {
		return err
	}

	// Create inner project directory for settings, urls, wsgi, apps
	innerDir := filepath.Join(projectName, projectName)
	if err := os.MkdirAll(innerDir, 0755); err != nil {
		return fmt.Errorf("failed to create inner project directory: %w", err)
	}

	files := []string{"settings.go", "urls.go", "wsgi.go", "apps.go"}
	for _, f := range files {
		tmplPath := fmt.Sprintf("templates/project/project/%s.tmpl", f)
		outPath := filepath.Join(innerDir, f)
		if err := renderTemplate(tmplPath, outPath, map[string]string{"ProjectName": projectName}); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully created project %s\n", projectName)
	return nil
}

func createApp(appName string) error {
	// Create app directory
	if err := os.MkdirAll(appName, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Create tests directory
	if err := os.MkdirAll(filepath.Join(appName, "tests"), 0755); err != nil {
		return fmt.Errorf("failed to create tests directory: %w", err)
	}

	files := []string{"app.go", "models.go", "views.go", "urls.go", "admin.go", "forms.go"}
	for _, f := range files {
		tmplPath := fmt.Sprintf("templates/app/%s.tmpl", f)
		outPath := filepath.Join(appName, f)
		if err := renderTemplate(tmplPath, outPath, map[string]string{"AppName": appName}); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully created app %s\n", appName)
	return nil
}

func renderTemplate(tmplPath, outPath string, data interface{}) error {
	content, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", tmplPath, err)
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outPath, err)
	}

	if err := tmpl.Execute(f, data); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to execute template %s: %w", tmplPath, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", outPath, err)
	}

	return nil
}
