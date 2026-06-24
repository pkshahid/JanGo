package migrations

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkshahid/JanGo/orm"
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
	buf.WriteString("\t\"github.com/pkshahid/JanGo/orm\"\n")
	buf.WriteString("\t\"github.com/pkshahid/JanGo/orm/migrations\"\n")
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
			fieldsStr = append(fieldsStr, "\t\t\t\t"+serializeField(f)+",")
		}

		metaStr := "nil"
		if o.Meta != nil {
			metaStr = fmt.Sprintf("&orm.Meta{DbTable: %q}", o.Meta.DbTable)
		}

		return fmt.Sprintf("migrations.CreateModel{\n"+
			"\t\t\tName: %q,\n"+
			"\t\t\tFields: []*orm.Field{\n"+
			"%s\n"+
			"\t\t\t},\n"+
			"\t\t\tMeta: %s,\n"+
			"\t\t}", o.Name, strings.Join(fieldsStr, "\n"), metaStr)

	case DeleteModel:
		return fmt.Sprintf("migrations.DeleteModel{Name: %q}", o.Name)

	case AddField:
		return fmt.Sprintf("migrations.AddField{\n"+
			"\t\t\tModel: %q,\n"+
			"\t\t\tName:  %q,\n"+
			"\t\t\tField: %s,\n"+
			"\t\t}", o.Model, o.Name, serializeField(o.Field))

	case RemoveField:
		return fmt.Sprintf("migrations.RemoveField{\n"+
			"\t\t\tModel: %q,\n"+
			"\t\t\tName:  %q,\n"+
			"\t\t}", o.Model, o.Name)

	case AlterField:
		return fmt.Sprintf("migrations.AlterField{\n"+
			"\t\t\tModel: %q,\n"+
			"\t\t\tName:  %q,\n"+
			"\t\t\tField: %s,\n"+
			"\t\t}", o.Model, o.Name, serializeField(o.Field))

	case RunSQL:
		return fmt.Sprintf("migrations.RunSQL{SQL: `%s`, ReverseSQL: `%s`}", o.SQL, o.ReverseSQL)

	default:
		return fmt.Sprintf("/* unsupported operation type %T */", op)
	}
}

func serializeField(f *orm.Field) string {
	opts := f.Options
	var optParts []string
	if opts.MaxLength != 0 {
		optParts = append(optParts, fmt.Sprintf("MaxLength: %d", opts.MaxLength))
	}
	if opts.MaxDigits != 0 {
		optParts = append(optParts, fmt.Sprintf("MaxDigits: %d", opts.MaxDigits))
	}
	if opts.DecimalPlaces != 0 {
		optParts = append(optParts, fmt.Sprintf("DecimalPlaces: %d", opts.DecimalPlaces))
	}
	if opts.Blank {
		optParts = append(optParts, "Blank: true")
	}
	if opts.Null {
		optParts = append(optParts, "Null: true")
	}
	if opts.Unique {
		optParts = append(optParts, "Unique: true")
	}
	if opts.DbIndex {
		optParts = append(optParts, "DbIndex: true")
	}
	if opts.AutoNow {
		optParts = append(optParts, "AutoNow: true")
	}
	if opts.AutoNowAdd {
		optParts = append(optParts, "AutoNowAdd: true")
	}
	if opts.AutoCreated {
		optParts = append(optParts, "AutoCreated: true")
	}
	if opts.UploadTo != "" {
		optParts = append(optParts, fmt.Sprintf("UploadTo: %q", opts.UploadTo))
	}
	if opts.To != "" {
		optParts = append(optParts, fmt.Sprintf("To: %q", opts.To))
	}
	if opts.OnDelete != "" {
		optParts = append(optParts, fmt.Sprintf("OnDelete: %q", opts.OnDelete))
	}
	if opts.RelatedName != "" {
		optParts = append(optParts, fmt.Sprintf("RelatedName: %q", opts.RelatedName))
	}
	if opts.DbColumn != "" {
		optParts = append(optParts, fmt.Sprintf("DbColumn: %q", opts.DbColumn))
	}

	var fieldParts []string
	fieldParts = append(fieldParts, fmt.Sprintf("Name: %q", f.Name))
	fieldParts = append(fieldParts, fmt.Sprintf("Column: %q", f.Column))
	fieldParts = append(fieldParts, fmt.Sprintf("Type: orm.%s", string(f.Type)))
	if f.PrimaryKey {
		fieldParts = append(fieldParts, "PrimaryKey: true")
	}
	if len(optParts) > 0 {
		fieldParts = append(fieldParts, fmt.Sprintf("Options: orm.FieldOptions{%s}", strings.Join(optParts, ", ")))
	}

	return "&orm.Field{" + strings.Join(fieldParts, ", ") + "}"
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
