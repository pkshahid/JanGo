import CodeBlock from '../components/CodeBlock'
import { BookOpen, Rocket, Database, Shield, Box } from 'lucide-react'
import './DeveloperGuides.css'

function DeveloperGuides() {
  return (
    <div className="guides-page">
      <section className="page-header">
        <div className="container">
          <span className="badge badge-cyan">Developer Guides</span>
          <h1>Build Real Applications</h1>
          <p>
            Step-by-step tutorials to take you from zero to production-ready
            JanGo applications.
          </p>
        </div>
      </section>

      <div className="container guides-content">
        {/* Guide 1: First App */}
        <section className="guide-section">
          <div className="guide-header">
            <div className="guide-number">01</div>
            <div>
              <h2><Rocket size={24} /> Your First App</h2>
              <p className="guide-intro">
                Create your first JanGo application from scratch. Learn how to define views,
                map URLs, and serve responses.
              </p>
            </div>
          </div>

          <div className="guide-steps">
            <div className="guide-step">
              <h3>Create the App Directory</h3>
              <p>
                In JanGo, an &quot;app&quot; is simply a Go package. Create a directory for your app
                inside the project.
              </p>
              <CodeBlock
                code={`# Inside your project root
mkdir pages
touch pages/views.go pages/urls.go`}
                language="bash"
                title="Terminal"
              />
            </div>

            <div className="guide-step">
              <h3>Write Your First View</h3>
              <p>Views handle HTTP requests and return responses.</p>
              <CodeBlock
                code={`package pages

import (
    "net/http"
    "github.com/pkshahid/JanGo/http/request"
    "github.com/pkshahid/JanGo/http/response"
)

func HomePageView(req *request.Request) response.Response {
    return response.HttpResponse(
        "Welcome to my JanGo site!", http.StatusOK,
    )
}

func AboutPageView(req *request.Request) response.Response {
    return response.HttpResponse("About us...", http.StatusOK)
}`}
                language="go"
                title="pages/views.go"
              />
            </div>

            <div className="guide-step">
              <h3>Configure App URLs</h3>
              <p>Map URL patterns to your view functions.</p>
              <CodeBlock
                code={`package pages

import (
    "github.com/pkshahid/JanGo/http/urls"
    "github.com/pkshahid/JanGo/http/views"
)

var UrlPatterns = []urls.Path{
    urls.Path{"", views.Func(HomePageView), "home"},
    urls.Path{"about/", views.Func(AboutPageView), "about"},
}`}
                language="go"
                title="pages/urls.go"
              />
            </div>

            <div className="guide-step">
              <h3>Include in Root URLs</h3>
              <p>Connect your app&apos;s URLs to the project&apos;s root URL configuration.</p>
              <CodeBlock
                code={`package mywebsite

import (
    "github.com/pkshahid/JanGo/http/urls"
    "mywebsite/pages"
)

var RootUrlPatterns = []urls.Path{
    urls.Include{"", pages.UrlPatterns},
}`}
                language="go"
                title="mywebsite/urls.go"
              />
            </div>

            <div className="guide-step">
              <h3>Run and Test</h3>
              <CodeBlock
                code={`go run manage.go runserver

# Visit http://localhost:8000/       → "Welcome to my JanGo site!"
# Visit http://localhost:8000/about/ → "About us..."`}
                language="bash"
                title="Terminal"
              />
            </div>
          </div>
        </section>

        {/* Guide 2: Blog Application */}
        <section className="guide-section">
          <div className="guide-header">
            <div className="guide-number">02</div>
            <div>
              <h2><BookOpen size={24} /> Building a Blog</h2>
              <p className="guide-intro">
                Build a complete blog with models, views, templates, and the admin panel.
                Covers ORM, generic views, and template inheritance.
              </p>
            </div>
          </div>

          <div className="guide-steps">
            <div className="guide-step">
              <h3>Define Models</h3>
              <p>Create your data models using Go structs with JanGo&apos;s ORM tags.</p>
              <CodeBlock
                code={`package blog

import (
    "time"
    "github.com/pkshahid/JanGo/orm"
)

type Category struct {
    orm.Model
    Name string \`gd:"CharField,max_length=100,unique"\`
}

type Post struct {
    orm.Model
    Title       string    \`gd:"CharField,max_length=200"\`
    Slug        string    \`gd:"SlugField,max_length=200,unique"\`
    Content     string    \`gd:"TextField"\`
    IsPublished bool      \`gd:"BooleanField,default=false"\`
    PublishedAt time.Time \`gd:"DateTimeField,null"\`
    CategoryID  int       \`gd:"ForeignKey,to=Category,on_delete=CASCADE"\`
}

func init() {
    orm.Register(Category{})
    orm.Register(Post{})
}`}
                language="go"
                title="blog/models.go"
              />
            </div>

            <div className="guide-step">
              <h3>Register with Admin</h3>
              <p>Get a full CRUD admin interface with just a few lines.</p>
              <CodeBlock
                code={`package blog

import "github.com/pkshahid/JanGo/admin"

func init() {
    admin.Site.Register(Post{}, admin.ModelAdmin{
        ListDisplay:  []string{"Title", "IsPublished", "PublishedAt"},
        SearchFields: []string{"Title", "Content"},
        ListFilter:   []string{"IsPublished", "CategoryID"},
    })
    admin.Site.Register(Category{}, admin.ModelAdmin{})
}`}
                language="go"
                title="blog/admin.go"
              />
            </div>

            <div className="guide-step">
              <h3>Create Generic Views</h3>
              <p>Use Go generics for zero-boilerplate list and detail views.</p>
              <CodeBlock
                code={`package blog

import (
    "github.com/pkshahid/JanGo/http/request"
    "github.com/pkshahid/JanGo/http/views/generic"
    "github.com/pkshahid/JanGo/orm"
)

var PostListView = generic.ListView[Post]{
    Model:             Post{},
    TemplateName:      "blog/post_list.html",
    ContextObjectName: "posts",
    GetQuerySet: func(req *request.Request) (*orm.QuerySet[Post], error) {
        return orm.Objects[Post]().
            Filter("IsPublished=?", true).
            OrderBy("-PublishedAt"), nil
    },
}

var PostDetailView = generic.DetailView[Post]{
    Model:             Post{},
    TemplateName:      "blog/post_detail.html",
    ContextObjectName: "post",
    SlugField:         "Slug",
    SlugUrlKwarg:      "slug",
}`}
                language="go"
                title="blog/views.go"
              />
            </div>

            <div className="guide-step">
              <h3>Wire Up URLs</h3>
              <CodeBlock
                code={`package blog

import (
    "github.com/pkshahid/JanGo/admin"
    "github.com/pkshahid/JanGo/http/urls"
    "github.com/pkshahid/JanGo/http/views"
)

var UrlPatterns = []urls.Path{
    urls.Path{"admin/", urls.IncludeAdmin(admin.Site)},
    urls.Path{"", views.Class(PostListView), "post_list"},
    urls.Path{"post/<slug:slug>/", views.Class(PostDetailView), "post_detail"},
}`}
                language="go"
                title="blog/urls.go"
              />
            </div>

            <div className="guide-step">
              <h3>Create Templates</h3>
              <CodeBlock
                code={`{# templates/blog/post_list.html #}
{% extends "base.html" %}

{% block content %}
<h1>Blog Posts</h1>
{% for post in posts %}
    <article>
        <h2><a href="{% url 'post_detail' post.Slug %}">
            {{ post.Title }}
        </a></h2>
        <time>{{ post.PublishedAt|date:"Jan 2, 2006" }}</time>
        <p>{{ post.Content|truncatewords:30 }}</p>
    </article>
{% empty %}
    <p>No posts published yet.</p>
{% endfor %}
{% endblock %}`}
                language="html"
                title="templates/blog/post_list.html"
              />
            </div>

            <div className="guide-step">
              <h3>Run Migrations and Start</h3>
              <CodeBlock
                code={`# Generate and apply migrations
go run manage.go makemigrations
go run manage.go migrate

# Create admin user
go run manage.go createsuperuser

# Start the server
go run manage.go runserver

# Visit http://localhost:8000/admin/ to manage posts
# Visit http://localhost:8000/ to see published posts`}
                language="bash"
                title="Terminal"
              />
            </div>
          </div>
        </section>

        {/* Guide 3: Forms */}
        <section className="guide-section">
          <div className="guide-header">
            <div className="guide-number">03</div>
            <div>
              <h2><Box size={24} /> Forms & Validation</h2>
              <p className="guide-intro">
                Handle user input with declarative forms, automatic validation,
                and HTML widget rendering.
              </p>
            </div>
          </div>

          <div className="guide-steps">
            <div className="guide-step">
              <h3>Define a Form</h3>
              <CodeBlock
                code={`package contact

import "github.com/pkshahid/JanGo/forms"

type ContactForm struct {
    forms.Form
    Name    forms.CharField    \`form:"max_length=100"\`
    Email   forms.EmailField
    Subject forms.CharField    \`form:"max_length=200"\`
    Message forms.CharField    \`form:"widget=Textarea"\`
}

// Or auto-generate from a model
type PostForm struct {
    forms.ModelForm[Post]
    // Fields auto-derived from Post model
}`}
                language="go"
                title="contact/forms.go"
              />
            </div>

            <div className="guide-step">
              <h3>Process in a View</h3>
              <CodeBlock
                code={`func ContactView(req *request.Request) response.Response {
    form := ContactForm{}

    if req.Method == "POST" {
        form.Bind(req.PostData())
        if form.IsValid() {
            // Access validated data
            name := form.CleanedData["Name"].(string)
            email := form.CleanedData["Email"].(string)

            // Send email, save to DB, etc.
            sendContactEmail(name, email, form.CleanedData)

            return response.Redirect("/thank-you/")
        }
    }

    return response.Render(req, "contact.html",
        map[string]interface{}{"form": form},
    )
}`}
                language="go"
                title="contact/views.go"
              />
            </div>

            <div className="guide-step">
              <h3>Render in Template</h3>
              <CodeBlock
                code={`{% extends "base.html" %}

{% block content %}
<h1>Contact Us</h1>
<form method="post">
    {% csrf_token %}
    {{ form.as_p }}
    <button type="submit">Send Message</button>
</form>
{% endblock %}`}
                language="html"
                title="templates/contact.html"
              />
            </div>
          </div>
        </section>

        {/* Guide 4: Auth */}
        <section className="guide-section">
          <div className="guide-header">
            <div className="guide-number">04</div>
            <div>
              <h2><Shield size={24} /> Authentication & Permissions</h2>
              <p className="guide-intro">
                Implement user registration, login, logout, and permission-based
                access control.
              </p>
            </div>
          </div>

          <div className="guide-steps">
            <div className="guide-step">
              <h3>Custom User Model</h3>
              <CodeBlock
                code={`package accounts

import "github.com/pkshahid/JanGo/auth"

type User struct {
    auth.AbstractUser
    Bio       string \`gd:"TextField,blank"\`
    AvatarURL string \`gd:"URLField,blank"\`
}

func init() {
    auth.SetUserModel(User{})
}`}
                language="go"
                title="accounts/models.go"
              />
            </div>

            <div className="guide-step">
              <h3>Login & Logout Views</h3>
              <CodeBlock
                code={`func LoginView(req *request.Request) response.Response {
    if req.Method == "POST" {
        username := req.PostData().Get("username")
        password := req.PostData().Get("password")

        user, err := auth.Authenticate(req, username, password)
        if err == nil {
            auth.Login(req, user)
            return response.Redirect("/dashboard/")
        }
        return response.Render(req, "login.html",
            map[string]interface{}{"error": "Invalid credentials"},
        )
    }
    return response.Render(req, "login.html", nil)
}

func LogoutView(req *request.Request) response.Response {
    auth.Logout(req)
    return response.Redirect("/")
}`}
                language="go"
                title="accounts/views.go"
              />
            </div>

            <div className="guide-step">
              <h3>Permission Checks</h3>
              <CodeBlock
                code={`// In views — check permissions
func CreatePostView(req *request.Request) response.Response {
    if !req.User.IsAuthenticated() {
        return response.Redirect("/login/")
    }
    if !req.User.HasPerm("blog.add_post") {
        return response.HttpResponse("Forbidden", 403)
    }
    // ... handle post creation
}

// In templates
{% if user.is_authenticated %}
    <p>Welcome, {{ user.Username }}!</p>
    {% if user.has_perm "blog.add_post" %}
        <a href="/posts/new/">New Post</a>
    {% endif %}
{% endif %}`}
                language="go"
                title="Permission patterns"
              />
            </div>
          </div>
        </section>

        {/* Guide 5: Database */}
        <section className="guide-section">
          <div className="guide-header">
            <div className="guide-number">05</div>
            <div>
              <h2><Database size={24} /> Advanced ORM Patterns</h2>
              <p className="guide-intro">
                Master complex queries, relationships, aggregations, and
                multi-database configurations.
              </p>
            </div>
          </div>

          <div className="guide-steps">
            <div className="guide-step">
              <h3>Complex Queries</h3>
              <CodeBlock
                code={`// Chained filtering
posts, _ := orm.Objects[Post]().
    Filter("IsPublished=?", true).
    Filter("CategoryID=?", categoryID).
    OrderBy("-PublishedAt").
    Limit(10).
    All()

// Get or 404
post, err := orm.Objects[Post]().Get("Slug=?", slug)
if err != nil {
    return response.HttpResponse("Not Found", 404)
}

// Create and save
newPost := Post{
    Title:   "Hello World",
    Slug:    "hello-world",
    Content: "My first post!",
}
err = orm.Save(&newPost)

// Update
post.Title = "Updated Title"
err = orm.Save(&post)

// Delete
err = orm.Delete(&post)`}
                language="go"
                title="QuerySet patterns"
              />
            </div>

            <div className="guide-step">
              <h3>Multiple Database Backends</h3>
              <CodeBlock
                code={`// In settings.go
DATABASES: map[string]orm.DatabaseConfig{
    "default": {
        Engine: "sqlite3",
        Name:   "db.sqlite3",
    },
    "postgres": {
        Engine:   "postgresql",
        Name:     "myapp",
        User:     "dbuser",
        Password: "dbpass",
        Host:     "localhost",
        Port:     "5432",
    },
    "mysql": {
        Engine:   "mysql",
        Name:     "myapp",
        User:     "root",
        Host:     "localhost",
        Port:     "3306",
    },
}`}
                language="go"
                title="Multi-database configuration"
              />
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default DeveloperGuides
