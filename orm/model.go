package orm

import (
	"time"
)

// Model is the base struct to be embedded by all models.
// It provides default primary key and timestamp fields.
type Model struct {
	ID        uint64     `gd:"BigAutoField,primary_key=true,auto_created=true"`
	CreatedAt time.Time  `gd:"DateTimeField,auto_now_add=true,auto_created=true"`
	UpdatedAt time.Time  `gd:"DateTimeField,auto_now=true,auto_created=true"`
	DeletedAt *time.Time `gd:"DateTimeField,null=true,blank=true,auto_created=true"`
}

// Index represents a database index on one or more fields.
type Index struct {
	Name   string
	Fields []string
	Unique bool
}

// Meta holds configuration for a model, similar to Django's Meta inner class.
type Meta struct {
	DbTable           string
	Ordering          []string
	UniqueTogether    [][]string
	Indexes           []Index
	VerboseName       string
	VerboseNamePlural string
	Abstract          bool
}

// ModelInterface allows models to define custom metadata via a ModelMeta method.
type ModelInterface interface {
	ModelMeta() *Meta
}
