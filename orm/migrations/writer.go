package migrations

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/godjango/godjango/orm"
)

// MigrationWriter handles generating Go code for a Migration.
type MigrationWriter struct {
	Migration *Migration
	BaseDir   string
}

func NewMigrationWriter(m *Migration, baseDir string) *MigrationWriter {
	return &MigrationWriter{
		Migration: m,
		BaseDir:   baseDir,
	}
}

// AsString returns the generated Go source code for the migration.
func (w *MigrationWriter) AsString() (string, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString("package migrations\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"github.com/godjango/godjango/orm\"\n")
	buf.WriteString("\t\"github.com/godjango/godjango/orm/migrations\"\n")
	buf.WriteString(")\n\n")

	// Write init function for self-registration
	buf.WriteString("func init() {\n")
	buf.WriteString(fmt.Sprintf("\tmigrations.RegisterMigration(&Migration_%s_%s)\n", w.Migration.App, w.Migration.Name))
	buf.WriteString("}\n\n")

	// Write migration struct
	buf.WriteString(fmt.Sprintf("var Migration_%s_%s = migrations.Migration{\n", w.Migration.App, w.Migration.Name))
	buf.WriteString(fmt.Sprintf("\tApp: \"%s\",\n", w.Migration.App))
	buf.WriteString(fmt.Sprintf("\tName: \"%s\",\n", w.Migration.Name))

	// Dependencies
	buf.WriteString("\tDependencies: []migrations.Dep{\n")
	for _, dep := range w.Migration.Dependencies {
		buf.WriteString(fmt.Sprintf("\t\t{\"%s\", \"%s\"},\n", dep.App, dep.Name))
	}
	buf.WriteString("\t},\n")

	// Operations
	buf.WriteString("\tOperations: []migrations.Operation{\n")
	for _, op := range w.Migration.Operations {
		buf.WriteString("\t\t" + serializeOperation(op) + ",\n")
	}
	buf.WriteString("\t},\n")
	buf.WriteString("}\n")

	// Format code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String(), fmt.Errorf("failed to format generated code: %v", err)
	}

	return string(formatted), nil
}

func serializeOperation(op Operation) string {
	switch o := op.(type) {
	case CreateModel:
		var fieldsStr []string
		for _, f := range o.Fields {
			fieldsStr = append(fieldsStr, serializeField(f))
		}

		metaStr := "nil"
		if o.Meta != nil {
			metaStr = fmt.Sprintf("&orm.Meta{DbTable: \"%s\"}", o.Meta.DbTable)
		}

		return fmt.Sprintf(`migrations.CreateModel{
			Name: "%s",
			Fields: []*orm.Field{
				%s
			},
			Meta: %s,
		}`, o.Name, strings.Join(fieldsStr, ",\n\t\t\t\t"), metaStr)

	case DeleteModel:
		return fmt.Sprintf(`migrations.DeleteModel{Name: "%s"}`, o.Name)

	case AddField:
		return fmt.Sprintf(`migrations.AddField{
			Model: "%s",
			Name:  "%s",
			Field: %s,
		}`, o.Model, o.Name, serializeField(o.Field))

	case RemoveField:
		return fmt.Sprintf(`migrations.RemoveField{
			Model: "%s",
			Name:  "%s",
		}`, o.Model, o.Name)

	case AlterField:
		return fmt.Sprintf(`migrations.AlterField{
			Model: "%s",
			Name:  "%s",
			Field: %s,
		}`, o.Model, o.Name, serializeField(o.Field))

	case RunSQL:
		return fmt.Sprintf("migrations.RunSQL{SQL: `%s`, ReverseSQL: `%s`}", o.SQL, o.ReverseSQL)

	default:
		return fmt.Sprintf("/* unsupported operation type %T */", op)
	}
}

func serializeField(f *orm.Field) string {
	opts := f.Options
	optsStr := fmt.Sprintf(`MaxLength: %d, MaxDigits: %d, DecimalPlaces: %d, Blank: %t, Null: %t, Unique: %t, DbIndex: %t, AutoNow: %t, AutoNowAdd: %t, UploadTo: "%s", To: "%s", OnDelete: "%s", RelatedName: "%s", DbColumn: "%s"`,
		opts.MaxLength, opts.MaxDigits, opts.DecimalPlaces, opts.Blank, opts.Null, opts.Unique, opts.DbIndex, opts.AutoNow, opts.AutoNowAdd, opts.UploadTo, opts.To, opts.OnDelete, opts.RelatedName, opts.DbColumn)

	return fmt.Sprintf(`&orm.Field{Name: "%s", Column: "%s", Type: orm.%s, PrimaryKey: %t, Options: orm.FieldOptions{%s}}`,
		f.Name, f.Column, string(f.Type), f.PrimaryKey, optsStr)
}

// WriteToFile generates the Go code and writes it to the app's migrations directory.
func (w *MigrationWriter) WriteToFile() (string, error) {
	dirPath := filepath.Join(w.BaseDir, w.Migration.App, "migrations")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	content, err := w.AsString()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(dirPath, w.Migration.Name+".go")
	err = os.WriteFile(filePath, []byte(content), 0644)
	return filePath, err
}
