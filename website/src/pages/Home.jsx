import { Link } from 'react-router-dom'
import { Zap, Shield, Box, Code, Database, Layout, Terminal, Globe, Layers, ArrowRight } from 'lucide-react'
import CodeBlock from '../components/CodeBlock'
import './Home.css'

const quickStartCode = `# Install the CLI tool
go install github.com/pkshahid/JanGo/cmd/godjango-admin@latest

# Create a new project
godjango-admin startproject myapp

# Run the development server
cd myapp && go run manage.go runserver`

const modelCode = `type Post struct {
    orm.Model
    Title       string \`gd:"CharField,max_length=200"\`
    Content     string \`gd:"TextField"\`
    IsPublished bool   \`gd:"BooleanField,default=false"\`
}

// Query with type-safe generics
posts, _ := orm.Objects[Post]().
    Filter("IsPublished=?", true).
    OrderBy("-CreatedAt").All()`

const viewCode = `// Generic class-based views with Go generics
var PostListView = generic.ListView[Post]{
    Model:             Post{},
    TemplateName:      "blog/post_list.html",
    ContextObjectName: "posts",
}

// Or use simple function views
func HomeView(req *request.Request) response.Response {
    return response.Render(req, "home.html", nil)
}`

function Home() {
  const features = [
    { icon: <Database size={24} />, title: 'Full ORM', desc: 'Type-safe QuerySets with Go generics, migrations, and multi-database support.' },
    { icon: <Layout size={24} />, title: 'Admin Panel', desc: 'Auto-generated admin UI for managing your data without writing a single line.' },
    { icon: <Shield size={24} />, title: 'Auth & Security', desc: 'Built-in authentication, permissions, CSRF, rate limiting, and CSP middleware.' },
    { icon: <Code size={24} />, title: 'Template Engine', desc: 'Django-compatible templating with inheritance, tags, filters, and context processors.' },
    { icon: <Zap size={24} />, title: 'Hot Reload', desc: 'Instant development feedback with built-in hot-reloading server.' },
    { icon: <Terminal size={24} />, title: 'Management CLI', desc: 'Cobra-powered commands: migrate, runserver, shell, collectstatic, and more.' },
    { icon: <Globe size={24} />, title: 'i18n & l10n', desc: 'Full internationalization with GNU .po parsing and context-based activation.' },
    { icon: <Layers size={24} />, title: 'Middleware Stack', desc: 'Composable middleware for sessions, caching, security, and custom logic.' },
    { icon: <Box size={24} />, title: 'Forms & Validation', desc: 'Declarative forms with auto-validation, widgets, and ModelForm generation.' },
  ]

  const stats = [
    { value: '20+', label: 'Built-in Packages' },
    { value: '100%', label: 'Django Parity' },
    { value: 'Go 1.18+', label: 'Generics Support' },
    { value: '0', label: 'External Build Tools' },
  ]

  return (
    <div className="home">
      {/* Hero Section */}
      <section className="hero">
        <div className="hero-bg-glow" />
        <div className="container hero-container">
          <div className="hero-content">
            <span className="badge badge-blue">Open Source Go Framework</span>
            <h1 className="hero-title">
              Django&apos;s Power.<br />
              <span className="gradient-text">Go&apos;s Speed.</span>
            </h1>
            <p className="hero-desc">
              JanGo is a full Django-equivalent web framework written in Go.
              Get the robust, batteries-included experience you love from Django,
              with Go&apos;s performance, concurrency, and static typing.
            </p>
            <div className="hero-actions">
              <Link to="/docs" className="btn btn-primary">
                Get Started <ArrowRight size={18} />
              </Link>
              <a href="https://github.com/pkshahid/JanGo" target="_blank" rel="noopener noreferrer" className="btn btn-secondary">
                View on GitHub
              </a>
            </div>
          </div>
          <div className="hero-code">
            <CodeBlock code={quickStartCode} language="bash" title="Quick Start" />
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="stats-section">
        <div className="container">
          <div className="stats-grid">
            {stats.map((stat, i) => (
              <div key={i} className="stat-item">
                <span className="stat-value">{stat.value}</span>
                <span className="stat-label">{stat.label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="section">
        <div className="container">
          <div className="section-header-center">
            <h2 className="section-title">Everything You Need, Built In</h2>
            <p className="section-subtitle">
              From ORM to admin panel, from auth to caching &mdash; JanGo ships with every component
              you need to build production-ready web applications.
            </p>
          </div>
          <div className="grid-3 features-grid">
            {features.map((feature, i) => (
              <div key={i} className="card feature-card">
                <div className="feature-icon">{feature.icon}</div>
                <h3>{feature.title}</h3>
                <p>{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Code Showcase */}
      <section className="section code-showcase">
        <div className="container">
          <div className="showcase-grid">
            <div className="showcase-text">
              <span className="badge badge-cyan">Type-Safe ORM</span>
              <h2>Models & QuerySets<br />with Go Generics</h2>
              <p>
                Define your schema with struct tags, query with a chainable
                type-safe API. No more runtime errors from mistyped field names.
                Migrations are auto-generated from your model changes.
              </p>
              <Link to="/docs" className="btn btn-secondary">
                Learn More <ArrowRight size={16} />
              </Link>
            </div>
            <div className="showcase-code">
              <CodeBlock code={modelCode} language="go" title="models.go" />
            </div>
          </div>

          <div className="showcase-grid reverse">
            <div className="showcase-text">
              <span className="badge badge-blue">Views & Routing</span>
              <h2>Generic Views for<br />Common Patterns</h2>
              <p>
                Use Go generics for class-based views like ListView, DetailView,
                CreateView, and more. Or keep it simple with function-based views.
                Pattern matching routes with typed parameters.
              </p>
              <Link to="/guides" className="btn btn-secondary">
                View Guides <ArrowRight size={16} />
              </Link>
            </div>
            <div className="showcase-code">
              <CodeBlock code={viewCode} language="go" title="views.go" />
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="section cta-section">
        <div className="container">
          <div className="cta-card">
            <h2>Ready to Build Something Great?</h2>
            <p>Get started with JanGo in minutes. No complex setup, no build tools required.</p>
            <div className="cta-actions">
              <Link to="/docs" className="btn btn-primary">
                Read the Docs <ArrowRight size={18} />
              </Link>
              <Link to="/features" className="btn btn-secondary">
                Explore Features
              </Link>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default Home
