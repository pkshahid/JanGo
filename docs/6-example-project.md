# Example Project: A Simple Blog

GoDjango includes a complete example blog application located in the `examples/blog/` directory of the repository. This example demonstrates many of the framework's core features working together, including the ORM, generic views, the admin interface, forms, and templates.

This guide will walk you through the key components of that example to help you understand how a real GoDjango application is structured.

## 1. The Entrypoint (`manage.go`)

Like all GoDjango projects, the blog example has a `manage.go` file at its root. It imports the main application package to ensure settings and models are registered, and then executes the management command system.

```go
// examples/blog/manage.go
package main

import (
	"fmt"
	"os"

	_ "github.com/godjango/godjango/examples/blog/app"
	"github.com/godjango/godjango/management"
)

func main() {
	if err := management.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## 2. Configuration (`app/settings.go`)

The `app/settings.go` file defines the project's configuration. It configures the database (SQLite), middleware, installed apps, template directories, and static file settings. Notice how it uses the `core.Configure()` function.

```go
// examples/blog/app/settings.go (snippet)
func init() {
	core.Configure(core.Settings{
		DEBUG:      true,
		SECRET_KEY: "secret-key-for-blog-example",
		DATABASES: map[string]orm.DatabaseConfig{
			"default": {Engine: "sqlite3", Name: "blog.db"},
		},
		INSTALLED_APPS: []string{
			"github.com/godjango/godjango/auth",
			"github.com/godjango/godjango/admin",
			// ... other core apps
			"github.com/godjango/godjango/examples/blog/app", // The blog app itself
		},
		ROOT_URLCONF: &UrlPatterns,
		// ... template and static settings
	})
}
```

## 3. Models (`app/models.go`)

The blog uses two simple models: `Category` and `Post`.

```go
// examples/blog/app/models.go (snippet)
type Category struct {
	orm.Model
	Name string `gd:"CharField,max_length=100,unique"`
}

type Post struct {
	orm.Model
	Title       string    `gd:"CharField,max_length=200"`
	Slug        string    `gd:"SlugField,max_length=200,unique"`
	Content     string    `gd:"TextField"`
	IsPublished bool      `gd:"BooleanField,default=false"`
	PublishedAt time.Time `gd:"DateTimeField,null"`
	CategoryID  int       `gd:"ForeignKey,to=Category,on_delete=CASCADE"`
	Category    *Category `gd:"-"` // Related object, not a DB column
}
```

## 4. Admin Integration (`app/admin.go`)

The models are registered with the admin site so they can be managed via the generated UI.

```go
// examples/blog/app/admin.go
func init() {
	admin.Site.Register(Post{}, admin.ModelAdmin{
		ListDisplay: []string{"Title", "Slug", "IsPublished", "PublishedAt"},
		SearchFields: []string{"Title", "Content"},
		ListFilter: []string{"IsPublished", "CategoryID"},
	})
	admin.Site.Register(Category{}, admin.ModelAdmin{})
}
```

## 5. Views (`app/views.go`)

The blog uses GoDjango's generic class-based views to handle common patterns without writing boilerplate code.

```go
// examples/blog/app/views.go (snippet)
var PostListView = generic.ListView[Post]{
	Model:             Post{},
	TemplateName:      "blog/post_list.html",
	ContextObjectName: "posts",
	GetQuerySet: func(req *request.Request) (*orm.QuerySet[Post], error) {
		// Only show published posts, ordered by newest first
		return orm.Objects[Post]().Filter("IsPublished=?", true).OrderBy("-PublishedAt"), nil
	},
}

var PostDetailView = generic.DetailView[Post]{
	Model:             Post{},
	TemplateName:      "blog/post_detail.html",
	ContextObjectName: "post",
	SlugField:         "Slug",
	SlugUrlKwarg:      "slug",
}
```

## 6. URLs (`app/urls.go`)

The views are mapped to URLs. Notice the inclusion of the admin URLs.

```go
// examples/blog/app/urls.go
var UrlPatterns = []urls.Path{
	urls.Path{"admin/", urls.IncludeAdmin(admin.Site)},
	urls.Path{"", views.Class(PostListView), "post_list"},
	urls.Path{"post/<slug:slug>/", views.Class(PostDetailView), "post_detail"},
}
```

## Running the Example

To run this example yourself:

1.  Clone the GoDjango repository.
2.  Navigate to `examples/blog`.
3.  Run migrations: `go run manage.go migrate`
4.  Create a superuser to access the admin: `go run manage.go createsuperuser`
5.  Start the server: `go run manage.go runserver`
6.  Visit `http://localhost:8000/admin/` to log in and create some categories and posts.
7.  Visit `http://localhost:8000/` to see your published posts!
