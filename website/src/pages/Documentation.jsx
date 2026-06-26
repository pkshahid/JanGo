import { Link } from 'react-router-dom'
import { Book, ArrowRight, Terminal, Database, Layout, Shield, Code, Layers, Box, Globe } from 'lucide-react'
import CodeBlock from '../components/CodeBlock'
import './Documentation.css'

const installCode = `# Install the JanGo CLI globally
go install github.com/pkshahid/JanGo/cmd/godjango-admin@latest

# Verify installation
godjango-admin version

# Create a new project
godjango-admin startproject mywebsite

# Navigate and run
cd mywebsite
go run manage.go runserver`

const projectStructure = `mywebsite/
├── manage.go          # Management command entrypoint
├── go.mod             # Go module definition
├── go.sum             # Dependencies checksums
└── mywebsite/         # Core application package
    ├── settings.go    # Project configuration
    └── urls.go        # Root URL routing`

const settingsCode = `package mywebsite

import (
    "github.com/pkshahid/JanGo/core"
    "github.com/pkshahid/JanGo/orm"
)

func init() {
    core.Configure(core.Settings{
        DEBUG:      true,
        SECRET_KEY: "your-secret-key-here",
        DATABASES: map[string]orm.DatabaseConfig{
            "default": {
                Engine: "sqlite3",
                Name:   "db.sqlite3",
            },
        },
        INSTALLED_APPS: []string{
            "github.com/pkshahid/JanGo/auth",
            "github.com/pkshahid/JanGo/admin",
            "github.com/pkshahid/JanGo/sessions",
            "mywebsite/blog",
        },
        MIDDLEWARE: []string{
            "SecurityMiddleware",
            "SessionMiddleware",
            "CSRFMiddleware",
            "AuthenticationMiddleware",
        },
        ROOT_URLCONF: &RootUrlPatterns,
        TEMPLATES: []core.TemplateConfig{{
            Backend: "godjango",
            Dirs:    []string{"templates"},
            AppDirs: true,
        }},
        STATIC_URL:  "/static/",
        STATIC_ROOT: "staticfiles",
    })
}`

const migrateCode = `# Create migrations from model changes
go run manage.go makemigrations

# Apply migrations to the database
go run manage.go migrate

# Create an admin superuser
go run manage.go createsuperuser

# Open interactive shell
go run manage.go shell

# Collect static files for production
go run manage.go collectstatic

# Check deployment readiness
go run manage.go check`

function Documentation() {
  const sections = [
    { id: 'installation', icon: <Terminal size={20} />, label: 'Installation' },
    { id: 'project-structure', icon: <Layout size={20} />, label: 'Project Structure' },
    { id: 'configuration', icon: <Code size={20} />, label: 'Configuration' },
    { id: 'commands', icon: <Terminal size={20} />, label: 'Commands' },
    { id: 'orm', icon: <Database size={20} />, label: 'ORM Basics' },
    { id: 'views', icon: <Globe size={20} />, label: 'Views & Routing' },
    { id: 'templates', icon: <Box size={20} />, label: 'Templates' },
    { id: 'middleware', icon: <Layers size={20} />, label: 'Middleware' },
    { id: 'auth', icon: <Shield size={20} />, label: 'Authentication' },
  ]

  return (
    <div className="docs-page">
      <section className="page-header">
        <div className="container">
          <span className="badge badge-blue">Documentation</span>
          <h1>Getting Started with JanGo</h1>
          <p>
            Everything you need to install, configure, and start building
            web applications with JanGo.
          </p>
        </div>
      </section>

      <div className="container docs-layout">
        {/* Sidebar */}
        <aside className="docs-sidebar">
          <nav className="docs-nav">
            <h4>On this page</h4>
            {sections.map(s => (
              <a key={s.id} href={`#${s.id}`} className="docs-nav-link">
                {s.icon} {s.label}
              </a>
            ))}
          </nav>
        </aside>

        {/* Main Content */}
        <div className="docs-content">
          {/* Installation */}
          <section id="installation" className="docs-section">
            <h2><Terminal size={24} /> Installation</h2>
            <p>
              JanGo requires <strong>Go 1.18+</strong> (for generics support) and <strong>Git</strong>.
              Install the framework CLI globally, then bootstrap your project.
            </p>
            <CodeBlock code={installCode} language="bash" title="Terminal" />
            <div className="docs-note">
              <strong>Note:</strong> Ensure <code>$GOPATH/bin</code> is in your system PATH so the{' '}
              <code>godjango-admin</code> command is available globally.
            </div>
          </section>

          {/* Project Structure */}
          <section id="project-structure" className="docs-section">
            <h2><Layout size={24} /> Project Structure</h2>
            <p>
              After running <code>startproject</code>, you get a clean, minimal project structure
              similar to Django&apos;s.
            </p>
            <CodeBlock code={projectStructure} language="text" title="Project layout" />
            <p>
              The <code>manage.go</code> file is your primary entrypoint for all management commands.
              The inner package contains your configuration and root URL routing.
            </p>
          </section>

          {/* Configuration */}
          <section id="configuration" className="docs-section">
            <h2><Code size={24} /> Configuration</h2>
            <p>
              Configure your project in the <code>settings.go</code> file using <code>core.Configure()</code>.
              This defines databases, installed apps, middleware, templates, and more.
            </p>
            <CodeBlock code={settingsCode} language="go" title="settings.go" />
            <p>
              See the <Link to="/configuration">Configuration Reference</Link> for all available settings.
            </p>
          </section>

          {/* Commands */}
          <section id="commands" className="docs-section">
            <h2><Terminal size={24} /> Management Commands</h2>
            <p>
              JanGo uses Cobra to provide a robust CLI. All commands are executed via <code>go run manage.go</code>.
            </p>
            <CodeBlock code={migrateCode} language="bash" title="Common commands" />
            <div className="commands-grid">
              <div className="command-item">
                <code>runserver</code>
                <span>Start dev server with hot reload</span>
              </div>
              <div className="command-item">
                <code>migrate</code>
                <span>Apply database migrations</span>
              </div>
              <div className="command-item">
                <code>makemigrations</code>
                <span>Generate migration files from model changes</span>
              </div>
              <div className="command-item">
                <code>createsuperuser</code>
                <span>Create an admin user</span>
              </div>
              <div className="command-item">
                <code>shell</code>
                <span>Interactive Go REPL with project context</span>
              </div>
              <div className="command-item">
                <code>collectstatic</code>
                <span>Gather static files to STATIC_ROOT</span>
              </div>
              <div className="command-item">
                <code>check</code>
                <span>Inspect deployment readiness</span>
              </div>
              <div className="command-item">
                <code>clearsessions</code>
                <span>Purge expired sessions</span>
              </div>
            </div>
          </section>

          {/* ORM */}
          <section id="orm" className="docs-section">
            <h2><Database size={24} /> ORM Basics</h2>
            <p>
              Define models as Go structs with <code>gd</code> struct tags. The ORM handles schema
              creation, migrations, and provides a chainable QuerySet API using Go generics.
            </p>
            <CodeBlock
              code={`type Article struct {
    orm.Model
    Title       string    \`gd:"CharField,max_length=200"\`
    Slug        string    \`gd:"SlugField,max_length=200,unique"\`
    Content     string    \`gd:"TextField"\`
    IsPublished bool      \`gd:"BooleanField,default=false"\`
    PublishedAt time.Time \`gd:"DateTimeField,null"\`
    CategoryID  int       \`gd:"ForeignKey,to=Category,on_delete=CASCADE"\`
}

// Register in init()
func init() { orm.Register(Article{}) }

// QuerySet operations
articles, _ := orm.Objects[Article]().
    Filter("IsPublished=?", true).
    OrderBy("-PublishedAt").
    All()

article, _ := orm.Objects[Article]().Get("Slug=?", "hello-world")`}
              language="go"
              title="models.go"
            />
          </section>

          {/* Views */}
          <section id="views" className="docs-section">
            <h2><Globe size={24} /> Views & Routing</h2>
            <p>
              Map URL patterns to view functions or generic class-based views. Supports typed
              parameters, namespacing, and modular includes.
            </p>
            <CodeBlock
              code={`// URL Patterns
var UrlPatterns = []urls.Path{
    urls.Path{"", views.Func(HomeView), "home"},
    urls.Path{"posts/", views.Class(PostListView), "post_list"},
    urls.Path{"post/<slug:slug>/", views.Func(PostDetail), "post_detail"},
    urls.Path{"admin/", urls.IncludeAdmin(admin.Site)},
}

// Function view
func HomeView(req *request.Request) response.Response {
    ctx := map[string]interface{}{"title": "Welcome"}
    return response.Render(req, "home.html", ctx)
}

// Generic ListView
var PostListView = generic.ListView[Post]{
    Model:        Post{},
    TemplateName: "post_list.html",
}`}
              language="go"
              title="urls.go / views.go"
            />
          </section>

          {/* Templates */}
          <section id="templates" className="docs-section">
            <h2><Box size={24} /> Templates</h2>
            <p>
              Django-compatible template syntax with inheritance, blocks, tags, and filters.
            </p>
            <CodeBlock
              code={`{# base.html #}
<!DOCTYPE html>
<html>
<head><title>{% block title %}My Site{% endblock %}</title></head>
<body>
    {% block content %}{% endblock %}
</body>
</html>

{# page.html #}
{% extends "base.html" %}
{% block title %}{{ page.Title }} - {{ block.super }}{% endblock %}
{% block content %}
    <h1>{{ page.Title|upper }}</h1>
    {% for post in posts %}
        <article>{{ post.Content|truncatewords:50 }}</article>
    {% empty %}
        <p>No posts yet.</p>
    {% endfor %}
{% endblock %}`}
              language="html"
              title="templates/"
            />
          </section>

          {/* Middleware */}
          <section id="middleware" className="docs-section">
            <h2><Layers size={24} /> Middleware</h2>
            <p>
              Middleware processes requests and responses in a composable pipeline. JanGo ships
              with middleware for security, sessions, CSRF, CORS, caching, and more.
            </p>
            <CodeBlock
              code={`MIDDLEWARE: []string{
    "SecurityMiddleware",       // HSTS, X-Content-Type, etc.
    "WhiteNoiseMiddleware",    // Static file serving
    "SessionMiddleware",       // Session management
    "CSRFMiddleware",          // CSRF token validation
    "AuthenticationMiddleware", // User loading from session
    "GZipMiddleware",          // Response compression
    "CacheMiddleware",         // Full-page caching
}`}
              language="go"
              title="Middleware stack in settings.go"
            />
          </section>

          {/* Auth */}
          <section id="auth" className="docs-section">
            <h2><Shield size={24} /> Authentication</h2>
            <p>
              JanGo provides a complete authentication system with user models, login/logout,
              password hashing, permissions, and session management.
            </p>
            <CodeBlock
              code={`// User model extends AbstractUser
type User struct {
    auth.AbstractUser
    Bio     string \`gd:"TextField,blank"\`
    Avatar  string \`gd:"ImageField,blank"\`
}

// Login view
func LoginView(req *request.Request) response.Response {
    user, err := auth.Authenticate(req, username, password)
    if err == nil {
        auth.Login(req, user)
        return response.Redirect("/dashboard/")
    }
    return response.Render(req, "login.html", map[string]interface{}{
        "error": "Invalid credentials",
    })
}

// Permission check
if req.User.HasPerm("blog.add_post") {
    // Allow action
}`}
              language="go"
              title="Authentication"
            />
          </section>

          {/* Next Steps */}
          <section className="docs-section docs-next">
            <h2><Book size={24} /> Next Steps</h2>
            <div className="next-steps-grid">
              <Link to="/guides" className="next-step-card">
                <h4>Developer Guides</h4>
                <p>Step-by-step tutorials for building real applications</p>
                <ArrowRight size={16} />
              </Link>
              <Link to="/features" className="next-step-card">
                <h4>All Features</h4>
                <p>Explore the complete feature set of JanGo</p>
                <ArrowRight size={16} />
              </Link>
              <Link to="/configuration" className="next-step-card">
                <h4>Configuration</h4>
                <p>Full reference for all settings and options</p>
                <ArrowRight size={16} />
              </Link>
            </div>
          </section>
        </div>
      </div>
    </div>
  )
}

export default Documentation
