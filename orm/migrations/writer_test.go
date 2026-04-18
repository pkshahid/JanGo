package migrations

import (
	"strings"
	"testing"
	"github.com/pkshahid/JanGo/orm"
)

func TestMigrationWriter(t *testing.T) {
	m := &Migration{
		App:  "testapp",
		Name: "0001_initial",
		Dependencies: []Dep{
			{"otherapp", "0001_initial"},
		},
		Operations: []Operation{
			CreateModel{
				Name: "Article",
				Fields: []*orm.Field{
					{Name: "ID", Column: "id", Type: orm.BigAutoField, PrimaryKey: true},
					{Name: "Title", Column: "title", Type: orm.CharField, Options: orm.FieldOptions{MaxLength: 100}},
				},
				Meta: &orm.Meta{DbTable: "testapp_article"},
			},
			AddField{
				Model: "Article",
				Name:  "Body",
				Field: &orm.Field{Name: "Body", Column: "body", Type: orm.TextField},
			},
		},
	}

	writer := NewMigrationWriter(m, "testdir")
	content, err := writer.AsString()
	if err != nil {
		t.Fatalf("Failed to generate string: %v", err)
	}

	if !strings.Contains(content, "package migrations") {
		t.Errorf("Missing package declaration")
	}
	if !strings.Contains(content, `RegisterMigration(&Migration_testapp_0001_initial)`) {
		t.Errorf("Missing self-registration")
	}
	if !strings.Contains(content, `{"otherapp", "0001_initial"}`) {
		t.Errorf("Missing dependency")
	}
	if !strings.Contains(content, `migrations.CreateModel{`) {
		t.Errorf("Missing CreateModel operation")
	}
	if !strings.Contains(content, `migrations.AddField{`) {
		t.Errorf("Missing AddField operation")
	}
}
