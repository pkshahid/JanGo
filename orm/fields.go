package orm

import (
	"reflect"
)

// FieldType represents the kind of database field.
type FieldType string

const (
	CharField           FieldType = "CharField"
	TextField           FieldType = "TextField"
	IntegerField        FieldType = "IntegerField"
	SmallIntegerField   FieldType = "SmallIntegerField"
	BigIntegerField     FieldType = "BigIntegerField"
	FloatField          FieldType = "FloatField"
	DecimalField        FieldType = "DecimalField"
	BooleanField        FieldType = "BooleanField"
	NullBooleanField    FieldType = "NullBooleanField"
	DateField           FieldType = "DateField"
	TimeField           FieldType = "TimeField"
	DateTimeField       FieldType = "DateTimeField"
	DurationField       FieldType = "DurationField"
	EmailField          FieldType = "EmailField"
	URLField            FieldType = "URLField"
	SlugField           FieldType = "SlugField"
	IPAddressField      FieldType = "IPAddressField"
	UUIDField           FieldType = "UUIDField"
	JSONField           FieldType = "JSONField"
	BinaryField         FieldType = "BinaryField"
	FileField           FieldType = "FileField"
	ImageField          FieldType = "ImageField"
	ForeignKey          FieldType = "ForeignKey"
	OneToOneField       FieldType = "OneToOneField"
	ManyToManyField     FieldType = "ManyToManyField"
	BigAutoField     FieldType = "BigAutoField"
)

// Field holds parsed information about a struct field.
type Field struct {
	Name       string        // Struct field name
	Column     string        // Database column name
	Type       FieldType     // Parsed `gd` type
	GoType     reflect.Type  // The underlying Go type
	Options    FieldOptions  // Parsed options from the `gd` tag
	PrimaryKey bool          // Is this field the primary key?
}

// FieldOptions stores standard kwargs passed to a Django field.
type FieldOptions struct {
	MaxLength       int
	MaxDigits       int
	DecimalPlaces   int
	Blank           bool
	Null            bool
	Default         any
	Unique          bool
	DbIndex         bool
	AutoNow         bool
	AutoNowAdd      bool
	AutoCreated     bool     // Field was auto-created by base Model (can be overridden by child)
	UploadTo        string
	To              string   // ForeignKey target
	OnDelete        string   // CASCADE, SET_NULL, PROTECT, RESTRICT, DO_NOTHING
	RelatedName     string
	DbColumn        string
	Through         string   // ManyToManyField through model
}
