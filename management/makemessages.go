package management

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/godjango/godjango/i18n"
	"github.com/spf13/cobra"
)

import (
	"context"
)

type MakeMessagesCmd struct{}

func (c *MakeMessagesCmd) Name() string {
	return "makemessages"
}

func (c *MakeMessagesCmd) Help() string {
	return "Extracts translation strings from code and templates"
}

func (c *MakeMessagesCmd) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("locale", "l", "", "Locale(s) to process (e.g. pt_BR, fr)")
}

func (c *MakeMessagesCmd) Execute(ctx context.Context, args []string) error {
	// A small hack since we don't have the cobra command directly to get flags easily without passing it.
	// We'll parse os.Args directly for demonstration or assume it's passed somehow.
	// Actually, the framework usually extracts flags into a struct or we can parse os.Args.

	// Better yet, we can loop os.Args to find -l or --locale
	var localeStr string
	for i, arg := range os.Args {
		if (arg == "-l" || arg == "--locale") && i+1 < len(os.Args) {
			localeStr = os.Args[i+1]
		} else if strings.HasPrefix(arg, "--locale=") {
			localeStr = strings.SplitN(arg, "=", 2)[1]
		}
	}

	if localeStr == "" {
		return fmt.Errorf("--locale or -l is required")
	}

	locales := strings.Split(localeStr, ",")

	rxTrans := regexp.MustCompile(`(?:Gettext|GettextLazy|_\()\s*\"([^\"]+)\"`)
	rxNtrans := regexp.MustCompile(`Ngettext\s*\(\s*.*?,?\s*\"([^\"]+)\"\s*,\s*\"([^\"]+)\"`)
	rxTmplTrans := regexp.MustCompile(`\{%\s*trans\s*\"([^\"]+)\"\s*(?:.*?)\%\}`)

	stringsFound := make(map[string]map[string]bool)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".html") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			for _, m := range rxTrans.FindAllStringSubmatch(string(content), -1) {
				if len(m) > 1 {
					if stringsFound[m[1]] == nil {
						stringsFound[m[1]] = make(map[string]bool)
					}
					stringsFound[m[1]][""] = true
				}
			}

			for _, m := range rxNtrans.FindAllStringSubmatch(string(content), -1) {
				if len(m) > 2 {
					if stringsFound[m[1]] == nil {
						stringsFound[m[1]] = make(map[string]bool)
					}
					stringsFound[m[1]][m[2]] = true
				}
			}

			for _, m := range rxTmplTrans.FindAllStringSubmatch(string(content), -1) {
				if len(m) > 1 {
					if stringsFound[m[1]] == nil {
						stringsFound[m[1]] = make(map[string]bool)
					}
					stringsFound[m[1]][""] = true
				}
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories: %v", err)
	}

	for _, loc := range locales {
		localeDir := filepath.Join(i18n.Config.LocalePaths[0], loc, "LC_MESSAGES")
		if err := os.MkdirAll(localeDir, 0755); err != nil {
			return fmt.Errorf("could not create locale dir: %v", err)
		}

		poPath := filepath.Join(localeDir, "django.po")
		var existingPO []byte
		if _, err := os.Stat(poPath); err == nil {
			existingPO, _ = os.ReadFile(poPath)
		}
		existingStr := string(existingPO)

		f, err := os.OpenFile(poPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		if len(existingStr) == 0 {
			f.WriteString("msgid \"\"\nmsgstr \"\"\n\"Content-Type: text/plain; charset=UTF-8\\n\"\n\"Language: " + loc + "\\n\"\n\n")
		}

		for msgid, plurals := range stringsFound {
			for plural := range plurals {
				if !strings.Contains(existingStr, "msgid \""+msgid+"\"") {
					if plural == "" {
						f.WriteString(fmt.Sprintf("msgid \"%s\"\nmsgstr \"\"\n\n", msgid))
					} else {
						f.WriteString(fmt.Sprintf("msgid \"%s\"\nmsgid_plural \"%s\"\nmsgstr[0] \"\"\nmsgstr[1] \"\"\n\n", msgid, plural))
					}
				}
			}
		}
		f.Close()
		fmt.Printf("Updated %s\n", poPath)
	}

	return nil
}

func init() {
	Register(&MakeMessagesCmd{})
}
