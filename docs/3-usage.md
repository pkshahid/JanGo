# Usage

This guide covers the core features of GoDjango and how to use them to build your web application.

## Routing (`urls.go`)

GoDjango uses a straightforward routing system. You define URL patterns and map them to views.

```go
package myapp

import (
    "github.com/godjango/godjango/http/urls"
    "github.com/godjango/godjango/http/views"
)

var UrlPatterns = []urls.Path{
    urls.Path{"", views.Func(HomeView), "home"},
    urls.Path{"about/", views.Func(AboutView), "about"},
    urls.Path{"post/<int:id>/", views.Func(PostDetailView), "post_detail"},
}
```

## Views (`views.go`)

Views handle HTTP requests and return HTTP responses. GoDjango provides functional views and generic class-based views.

### Functional Views

```go
package myapp

import (
    "net/http"
    "github.com/godjango/godjango/http/request"
    "github.com/godjango/godjango/http/response"
)

func HomeView(req *request.Request) response.Response {
    return response.HttpResponse("Hello, GoDjango!", http.StatusOK)
}

func TemplateView(req *request.Request) response.Response {
    context := map[string]interface{}{"name": "World"}
    return response.Render(req, "index.html", context)
}
```

### Generic Views

GoDjango leverages Go generics to provide reusable class-based views (e.g., `ListView`, `DetailView`).

```go
package myapp

import (
    "github.com/godjango/godjango/http/views/generic"
)

var PostListView = generic.ListView[Post]{
    Model:        Post{},
    TemplateName: "post_list.html",
    ContextObjectName: "posts",
}
```

## Models (ORM) (`models.go`)

The ORM allows you to define your database schema using Go structs and struct tags.

```go
package myapp

import (
    "github.com/godjango/godjango/orm"
)

type Post struct {
    orm.Model // Provides ID, CreatedAt, UpdatedAt
    Title     string `gd:"CharField,max_length=200"`
    Content   string `gd:"TextField"`
    IsPublished bool `gd:"BooleanField,default=false"`
}

// Optional: Define metadata
func (p Post) ModelMeta() *orm.Meta {
    return &orm.Meta{
        TableName: "blog_posts",
        Ordering:  []string{"-CreatedAt"},
    }
}

// Register models at startup
func init() {
    orm.Register(Post{})
}
```

### Querying Data (QuerySets)

The generic `QuerySet` API provides a chainable interface for querying the database.

```go
// Get all posts
posts, err := orm.Objects[Post]().All()

// Filter posts
publishedPosts, err := orm.Objects[Post]().Filter("IsPublished=?", true).All()

// Get a single post
post, err := orm.Objects[Post]().Get("ID=?", 1)

// Create a new post
newPost := Post{Title: "Hello", Content: "World", IsPublished: true}
err = orm.Save(&newPost)
```

## Templates (`templates/`)

GoDjango uses a Django-like templating engine, supporting inheritance, tags, and filters.

`base.html`:
```html
<!DOCTYPE html>
<html>
<head><title>{% block title %}My Site{% endblock %}</title></head>
<body>
    <div id="content">
        {% block content %}{% endblock %}
    </div>
</body>
</html>
```

`index.html`:
```html
{% extends "base.html" %}

{% block title %}Home - {{ block.super }}{% endblock %}

{% block content %}
    <h1>Welcome, {{ name|default:"Guest" }}!</h1>
    <ul>
        {% for item in items %}
            <li>{{ item }}</li>
        {% endfor %}
    </ul>
{% endblock %}
```

## Forms (`forms.go`)

Forms handle input validation, binding, and HTML generation. You can define regular forms or `ModelForm`s.

```go
package myapp

import (
    "github.com/godjango/godjango/forms"
)

// A standard form
type ContactForm struct {
    forms.Form
    Name    forms.CharField `form:"max_length=100"`
    Email   forms.EmailField
    Message forms.CharField `form:"widget=Textarea"`
}

// A ModelForm derived from the Post model
type PostForm struct {
    forms.ModelForm[Post]
}
```

In your view:
```go
func ContactView(req *request.Request) response.Response {
    form := ContactForm{}
    if req.Method == "POST" {
        form.Bind(req.PostData())
        if form.IsValid() {
            // Process form.CleanedData
            return response.Redirect("/")
        }
    }
    return response.Render(req, "contact.html", map[string]interface{}{"form": form})
}
```
