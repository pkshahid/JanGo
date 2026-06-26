package orm

import (
	"testing"
	"time"
)

type User struct {
	Model
	Username string `gd:"CharField,max_length=150,unique=true"`
	IsActive bool   `gd:"BooleanField,default=true"`
}

type Article struct {
	Model
	Title    string    `gd:"CharField,max_length=200,db_index=true"`
	Body     string    `gd:"TextField"`
	Author   *User     `gd:"ForeignKey,to=auth.User,on_delete=CASCADE,db_column=author_id"`
	Created  time.Time `gd:"DateTimeField,auto_now_add=true"`
	Views    int       `gd:"IntegerField,default=0"`
	Rating   float64   `gd:"FloatField,null=true,blank=true"`
	Price    float64   `gd:"DecimalField,max_digits=10,decimal_places=2"`
	Image    string    `gd:"ImageField,upload_to=images/"`
	Internal string    `gd:"-"` // Ignore
}

func (a *Article) ModelMeta() *Meta {
	return &Meta{
		DbTable:        "blog_article",
		Ordering:       []string{"-created"},
		UniqueTogether: [][]string{{"Title", "Author"}},
		Indexes: []Index{
			{Name: "title_idx", Fields: []string{"Title"}, Unique: false},
		},
		VerboseName:       "Article",
		VerboseNamePlural: "Articles",
	}
}

type InvalidModel struct {
	ID uint64 `gd:"BigAutoField,primary_key=true"`
	PK uint64 `gd:"BigAutoField,primary_key=true"` // Multiple PKs
}

// ProxyUser is a proxy of User — shares the same table, no own table.
type ProxyUser struct {
	User
}

func (p *ProxyUser) ModelMeta() *Meta {
	return &Meta{
		Proxy: true,
	}
}

// ProxyWithAbstract is invalid — cannot be both proxy and abstract.
type ProxyWithAbstract struct {
	User
}

func (p *ProxyWithAbstract) ModelMeta() *Meta {
	return &Meta{
		Proxy:    true,
		Abstract: true,
	}
}

// ProxyWithoutParent has no embedded parent model.
type ProxyWithoutParent struct {
	Model
	Name string `gd:"CharField,max_length=100"`
}

func (p *ProxyWithoutParent) ModelMeta() *Meta {
	return &Meta{
		Proxy: true,
	}
}

func TestParser(t *testing.T) {
	ClearRegistry()

	// Parse valid model
	info, err := parseModel(&Article{})
	if err != nil {
		t.Fatalf("Failed to parse Article: %v", err)
	}

	if info.Name != "Article" {
		t.Errorf("Expected name 'Article', got '%s'", info.Name)
	}

	// Verify Meta
	if info.Meta.DbTable != "blog_article" {
		t.Errorf("Expected DbTable 'blog_article', got '%s'", info.Meta.DbTable)
	}
	if info.Meta.Ordering[0] != "-created" {
		t.Errorf("Expected ordering '-created', got '%s'", info.Meta.Ordering[0])
	}

	// Verify embedded fields
	if f, ok := info.FieldByName["ID"]; !ok {
		t.Errorf("Embedded field ID missing")
	} else if f.Type != BigAutoField || !f.PrimaryKey {
		t.Errorf("Embedded ID field incorrect: %+v", f)
	}

	// Verify fields
	title := info.FieldByName["Title"]
	if title == nil || title.Type != CharField || title.Options.MaxLength != 200 || !title.Options.DbIndex {
		t.Errorf("Title field parsed incorrectly: %+v", title)
	}

	author := info.FieldByName["Author"]
	if author == nil || author.Type != ForeignKey || author.Options.To != "auth.User" || author.Options.OnDelete != "CASCADE" || author.Column != "author_id" {
		t.Errorf("Author field parsed incorrectly: %+v", author)
	}

	created := info.FieldByName["Created"]
	if created == nil || created.Type != DateTimeField || !created.Options.AutoNowAdd {
		t.Errorf("Created field parsed incorrectly: %+v", created)
	}

	price := info.FieldByName["Price"]
	if price == nil || price.Options.MaxDigits != 10 || price.Options.DecimalPlaces != 2 {
		t.Errorf("Price field parsed incorrectly: %+v", price)
	}

	image := info.FieldByName["Image"]
	if image == nil || image.Options.UploadTo != "images/" {
		t.Errorf("Image field parsed incorrectly: %+v", image)
	}

	if _, ok := info.FieldByName["Internal"]; ok {
		t.Errorf("Internal ignored field should not be parsed")
	}

	// Check Snake case column conversion
	if info.FieldByName["IsActive"] != nil { // Wait IsActive is in User, not Article
		t.Errorf("IsActive shouldn't be here")
	}
}

func TestParserSnakeCase(t *testing.T) {
	ClearRegistry()

	info, _ := parseModel(&User{})
	isActive := info.FieldByName["IsActive"]
	if isActive.Column != "is_active" {
		t.Errorf("Expected snake_case column 'is_active', got '%s'", isActive.Column)
	}
}

func TestParserErrors(t *testing.T) {
	_, err := parseModel(&InvalidModel{})
	if err == nil {
		t.Errorf("Expected error for multiple primary keys")
	}

	_, err = parseModel("not a struct")
	if err == nil {
		t.Errorf("Expected error for non-struct type")
	}
}

func TestRegistry(t *testing.T) {
	ClearRegistry()

	err := Register(&User{})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Double registration should fail
	err = Register(&User{})
	if err == nil {
		t.Errorf("Expected error on duplicate registration")
	}

	info, err := GetModelInfo(&User{})
	if err != nil {
		t.Fatalf("GetModelInfo failed: %v", err)
	}

	if info.Name != "User" {
		t.Errorf("Expected User, got %s", info.Name)
	}

	// Unregistered model
	_, err = GetModelInfo(&Article{})
	if err == nil {
		t.Errorf("Expected error for unregistered model")
	}
}

// Test extracting generic field types
func TestFieldTypes(t *testing.T) {
	type TypesModel struct {
		Model
		F1  string         `gd:"CharField"`
		F2  string         `gd:"TextField"`
		F3  int            `gd:"IntegerField"`
		F4  int16          `gd:"SmallIntegerField"`
		F5  int64          `gd:"BigIntegerField"`
		F6  float32        `gd:"FloatField"`
		F7  float64        `gd:"DecimalField"`
		F8  bool           `gd:"BooleanField"`
		F9  *bool          `gd:"NullBooleanField"`
		F10 time.Time      `gd:"DateField"`
		F11 time.Time      `gd:"TimeField"`
		F12 time.Time      `gd:"DateTimeField"`
		F13 time.Duration  `gd:"DurationField"`
		F14 string         `gd:"EmailField"`
		F15 string         `gd:"URLField"`
		F16 string         `gd:"SlugField"`
		F17 string         `gd:"IPAddressField"`
		F18 string         `gd:"UUIDField"`
		F19 map[string]any `gd:"JSONField"`
		F20 []byte         `gd:"BinaryField"`
		F21 string         `gd:"FileField"`
		F22 string         `gd:"ImageField"`
		F23 *User          `gd:"ForeignKey"`
		F24 *User          `gd:"OneToOneField"`
		F25 []*User        `gd:"ManyToManyField"`
	}

	info, _ := parseModel(&TypesModel{})

	expected := map[string]FieldType{
		"F1": CharField, "F2": TextField, "F3": IntegerField,
		"F4": SmallIntegerField, "F5": BigIntegerField,
		"F6": FloatField, "F7": DecimalField, "F8": BooleanField,
		"F9": NullBooleanField, "F10": DateField, "F11": TimeField,
		"F12": DateTimeField, "F13": DurationField, "F14": EmailField,
		"F15": URLField, "F16": SlugField, "F17": IPAddressField,
		"F18": UUIDField, "F19": JSONField, "F20": BinaryField,
		"F21": FileField, "F22": ImageField, "F23": ForeignKey,
		"F24": OneToOneField, "F25": ManyToManyField,
	}

	for name, typ := range expected {
		if info.FieldByName[name].Type != typ {
			t.Errorf("Expected %s for %s, got %s", typ, name, info.FieldByName[name].Type)
		}
		// Verify underlying Go type maps perfectly. We just check existence.
		if info.FieldByName[name].GoType == nil {
			t.Errorf("GoType nil for %s", name)
		}
	}
}

func TestProxyModel(t *testing.T) {
	ClearRegistry()

	// Register parent first
	if err := Register(&User{}); err != nil {
		t.Fatalf("Failed to register User: %v", err)
	}

	// Parse proxy model
	info, err := parseModel(&ProxyUser{})
	if err != nil {
		t.Fatalf("Failed to parse ProxyUser: %v", err)
	}

	if !info.Meta.Proxy {
		t.Errorf("Expected Proxy=true")
	}

	// Proxy should use parent's table
	if info.Meta.DbTable != "user" {
		t.Errorf("Expected DbTable 'user' (parent's table), got '%s'", info.Meta.DbTable)
	}

	// Proxy should inherit parent's fields
	if _, ok := info.FieldByName["Username"]; !ok {
		t.Errorf("Proxy should inherit parent field 'Username'")
	}
	if _, ok := info.FieldByName["IsActive"]; !ok {
		t.Errorf("Proxy should inherit parent field 'IsActive'")
	}
}

func TestProxyModelErrors(t *testing.T) {
	ClearRegistry()

	// Proxy + Abstract is invalid
	_, err := parseModel(&ProxyWithAbstract{})
	if err == nil {
		t.Errorf("Expected error for proxy+abstract model")
	}

	// Proxy without parent model is invalid
	_, err = parseModel(&ProxyWithoutParent{})
	if err == nil {
		t.Errorf("Expected error for proxy model without parent")
	}
}

func TestProxyModelRegistration(t *testing.T) {
	ClearRegistry()

	// Register parent and proxy
	if err := Register(&User{}, &ProxyUser{}); err != nil {
		t.Fatalf("Failed to register User and ProxyUser: %v", err)
	}

	// Proxy model should be in registry
	info, err := GetModelInfo(&ProxyUser{})
	if err != nil {
		t.Fatalf("ProxyUser not found in registry: %v", err)
	}

	if info.Meta.DbTable != "user" {
		t.Errorf("Expected proxy DbTable 'user', got '%s'", info.Meta.DbTable)
	}
}
