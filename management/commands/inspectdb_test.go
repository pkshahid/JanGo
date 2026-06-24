package commands

import (
	"strings"
	"testing"
)

func TestInspectDBGenerateModels(t *testing.T) {
	tables := []TableInfo{
		{
			Name: "blog_post",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
				{Name: "title", DataType: "VARCHAR", MaxLength: 200},
				{Name: "content", DataType: "TEXT", Nullable: true},
				{Name: "views", DataType: "INTEGER"},
				{Name: "is_published", DataType: "BOOLEAN"},
				{Name: "created_at", DataType: "DATETIME"},
			},
		},
		{
			Name: "auth_user",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
				{Name: "username", DataType: "VARCHAR", MaxLength: 150},
				{Name: "email", DataType: "VARCHAR", MaxLength: 254},
				{Name: "is_active", DataType: "BOOLEAN"},
			},
		},
	}

	output := InspectDBFromTableInfo(tables)

	// Check package declaration
	if !strings.Contains(output, "package models") {
		t.Error("Expected 'package models' in output")
	}

	// Check import
	if !strings.Contains(output, `"github.com/pkshahid/JanGo/orm"`) {
		t.Error("Expected orm import")
	}

	// Check model struct generated
	if !strings.Contains(output, "type BlogPost struct") {
		t.Error("Expected BlogPost struct")
	}
	if !strings.Contains(output, "type AuthUser struct") {
		t.Error("Expected AuthUser struct")
	}

	// Check orm.Model embedded
	if !strings.Contains(output, "orm.Model") {
		t.Error("Expected orm.Model embedding")
	}

	// Check field types generated
	if !strings.Contains(output, "Title string") {
		t.Error("Expected Title string field")
	}
	if !strings.Contains(output, "CharField") {
		t.Error("Expected CharField tag")
	}
	if !strings.Contains(output, "max_length=200") {
		t.Error("Expected max_length=200 in tag")
	}

	// Check nullable field
	if !strings.Contains(output, "null=true") {
		t.Error("Expected null=true for nullable field")
	}

	// Check base model fields are excluded (id, created_at)
	if strings.Contains(output, `Id `) {
		t.Error("Base model field 'id' should be excluded")
	}

	// Check ModelMeta generated
	if !strings.Contains(output, `DbTable: "blog_post"`) {
		t.Error("Expected DbTable in Meta")
	}
}

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"blog_post", "BlogPost"},
		{"auth_user", "AuthUser"},
		{"simple", "Simple"},
		{"multi_word_name", "MultiWordName"},
	}

	for _, tt := range tests {
		result := toGoName(tt.input)
		if result != tt.expected {
			t.Errorf("toGoName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSqlTypeToGoType(t *testing.T) {
	tests := []struct {
		sqlType  string
		nullable bool
		expected string
	}{
		{"INTEGER", false, "int"},
		{"INTEGER", true, "*int"},
		{"VARCHAR", false, "string"},
		{"VARCHAR", true, "string"}, // strings don't get pointer
		{"TEXT", false, "string"},
		{"BOOLEAN", false, "bool"},
		{"BOOLEAN", true, "*bool"},
		{"FLOAT", false, "float64"},
		{"REAL", false, "float64"},
		{"BLOB", false, "[]byte"},
	}

	for _, tt := range tests {
		result := sqlTypeToGoType(tt.sqlType, tt.nullable)
		if result != tt.expected {
			t.Errorf("sqlTypeToGoType(%q, %v) = %q, want %q", tt.sqlType, tt.nullable, result, tt.expected)
		}
	}
}

func TestSqlTypeToFieldType(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		{"INTEGER", "IntegerField"},
		{"VARCHAR", "CharField"},
		{"TEXT", "TextField"},
		{"BOOLEAN", "BooleanField"},
		{"FLOAT", "FloatField"},
		{"DATE", "DateField"},
		{"DATETIME", "DateTimeField"},
		{"JSON", "JSONField"},
		{"BLOB", "BinaryField"},
	}

	for _, tt := range tests {
		result := sqlTypeToFieldType(tt.sqlType)
		if result != tt.expected {
			t.Errorf("sqlTypeToFieldType(%q) = %q, want %q", tt.sqlType, result, tt.expected)
		}
	}
}
