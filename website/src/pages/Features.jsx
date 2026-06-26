import { Database, Layout, Shield, Code, Terminal, Globe, Layers, Box, Zap, Mail, Eye, Activity, Server, FileText, FlaskConical } from 'lucide-react'
import CodeBlock from '../components/CodeBlock'
import './Features.css'

function Features() {
  const featureCategories = [
    {
      title: 'Data & Storage',
      features: [
        {
          icon: <Database size={22} />,
          name: 'ORM & QuerySets',
          desc: 'Type-safe generic QuerySets, model definitions with struct tags, auto-migrations, and support for SQLite, PostgreSQL, and MySQL backends.',
          badge: 'Core'
        },
        {
          icon: <Server size={22} />,
          name: 'Cache Framework',
          desc: 'Pluggable cache backends including Redis, Memcached, Database, Local Memory, and Dummy. Per-view caching decorators and template fragment caching.',
          badge: 'Core'
        },
        {
          icon: <Layers size={22} />,
          name: 'Sessions',
          desc: 'Multiple session storage backends with HMAC-SHA256 secured JSON encoding. Cookie-based, database-backed, and file-based sessions supported.',
          badge: 'Core'
        },
      ]
    },
    {
      title: 'Web & Routing',
      features: [
        {
          icon: <Globe size={22} />,
          name: 'HTTP Router',
          desc: 'Django-style URL patterns with named routes, typed parameters (int, slug, uuid), URL namespacing, and include() for modular apps.',
          badge: 'Core'
        },
        {
          icon: <Code size={22} />,
          name: 'Generic Views',
          desc: 'Go-generic class-based views: ListView, DetailView, CreateView, UpdateView, DeleteView. Reduce boilerplate to near zero.',
          badge: 'Core'
        },
        {
          icon: <Layers size={22} />,
          name: 'Middleware',
          desc: 'Composable middleware stack for security, sessions, CSRF, CORS, caching, GZip compression, and custom request/response processing.',
          badge: 'Core'
        },
      ]
    },
    {
      title: 'UI & Templates',
      features: [
        {
          icon: <FileText size={22} />,
          name: 'Template Engine',
          desc: 'Django-compatible syntax with template inheritance (extends/block), custom tags, filters, context processors, and multiple loaders (Filesystem, Embed, Cached).',
          badge: 'Core'
        },
        {
          icon: <Box size={22} />,
          name: 'Forms & Validation',
          desc: 'Declarative form fields with automatic validation, HTML widget rendering, bound/unbound states, and ModelForm auto-generation from ORM models.',
          badge: 'Core'
        },
        {
          icon: <Layout size={22} />,
          name: 'Admin Panel',
          desc: 'Auto-generated CRUD admin interface. Register models and get list views, search, filtering, inline editing, and custom actions out of the box.',
          badge: 'Core'
        },
      ]
    },
    {
      title: 'Security & Auth',
      features: [
        {
          icon: <Shield size={22} />,
          name: 'Authentication',
          desc: 'Full user model with AbstractUser, permission system, login/logout views, password hashing (bcrypt/argon2), and session-based auth.',
          badge: 'Security'
        },
        {
          icon: <Shield size={22} />,
          name: 'Security Middleware',
          desc: 'Content Security Policy (CSP), rate limiting, Allowed Hosts validation, X-Frame-Options, HSTS, and clickjacking protection.',
          badge: 'Security'
        },
        {
          icon: <Shield size={22} />,
          name: 'CSRF Protection',
          desc: 'Automatic CSRF token generation and validation for all POST/PUT/DELETE requests. Template tags for easy form integration.',
          badge: 'Security'
        },
      ]
    },
    {
      title: 'Developer Tools',
      features: [
        {
          icon: <Terminal size={22} />,
          name: 'Management Commands',
          desc: 'Cobra-powered CLI: runserver, migrate, makemigrations, createsuperuser, shell, collectstatic, check, and custom commands.',
          badge: 'DX'
        },
        {
          icon: <Zap size={22} />,
          name: 'Hot Reload',
          desc: 'Built-in file watcher that automatically restarts the development server on code changes. Zero-config development experience.',
          badge: 'DX'
        },
        {
          icon: <Eye size={22} />,
          name: 'Debug Toolbar',
          desc: 'Django-equivalent debug toolbar with panels for SQL queries, templates, HTTP headers, profiling, cache operations, and signals.',
          badge: 'DX'
        },
        {
          icon: <FlaskConical size={22} />,
          name: 'Testing Utilities',
          desc: 'Test cases with isolated databases, HTTP test client, request factory, and Django-style test assertions. Run with standard `go test`.',
          badge: 'DX'
        },
      ]
    },
    {
      title: 'Production & Ops',
      features: [
        {
          icon: <Activity size={22} />,
          name: 'Monitoring',
          desc: 'OpenTelemetry tracing integration, Prometheus metrics export, and structured logging with Go\'s slog package.',
          badge: 'Ops'
        },
        {
          icon: <Mail size={22} />,
          name: 'Email Framework',
          desc: 'Django-like EmailMessage and EmailMultiAlternatives with pluggable backends. SMTP, console, file-based, and custom backends.',
          badge: 'Core'
        },
        {
          icon: <Globe size={22} />,
          name: 'Internationalization',
          desc: 'Full i18n/l10n support using Go\'s x/text package. GNU .po catalog parsing, context-based language activation, and template translation tags.',
          badge: 'Core'
        },
        {
          icon: <Server size={22} />,
          name: 'Static & Media Files',
          desc: 'Static file finders, WhiteNoise middleware for self-contained serving, S3 storage backend, and collectstatic management command.',
          badge: 'Core'
        },
      ]
    },
  ]

  const comparisonCode = `// Django (Python)
class Post(models.Model):
    title = models.CharField(max_length=200)
    content = models.TextField()
    published = models.BooleanField(default=False)

posts = Post.objects.filter(published=True)

# JanGo (Go) — Same pattern, type-safe
type Post struct {
    orm.Model
    Title     string \`gd:"CharField,max_length=200"\`
    Content   string \`gd:"TextField"\`
    Published bool   \`gd:"BooleanField,default=false"\`
}

posts, _ := orm.Objects[Post]().Filter("Published=?", true).All()`

  return (
    <div className="features-page">
      <section className="page-header">
        <div className="container">
          <span className="badge badge-cyan">Full Feature Set</span>
          <h1>Features</h1>
          <p>
            Every component you need to build production-ready web applications.
            All built in, all type-safe, all blazing fast.
          </p>
        </div>
      </section>

      {/* Feature Categories */}
      {featureCategories.map((category, ci) => (
        <section
          key={ci}
          className="section"
          style={{ background: ci % 2 === 1 ? 'var(--navy-800)' : 'transparent' }}
        >
          <div className="container">
            <h2 className="category-title">{category.title}</h2>
            <div className={`grid-${category.features.length > 3 ? '4' : '3'} features-detail-grid`}>
              {category.features.map((feature, fi) => (
                <div key={fi} className="card feature-detail-card">
                  <div className="feature-detail-header">
                    <div className="feature-detail-icon">{feature.icon}</div>
                    <span className="badge badge-blue">{feature.badge}</span>
                  </div>
                  <h3>{feature.name}</h3>
                  <p>{feature.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>
      ))}

      {/* Comparison */}
      <section className="section">
        <div className="container">
          <div className="comparison-section">
            <div className="comparison-text">
              <h2 className="section-title">Django to JanGo</h2>
              <p className="section-subtitle">
                The same mental models and patterns you know from Django,
                now with compile-time type safety and Go performance.
              </p>
            </div>
            <CodeBlock code={comparisonCode} language="go" title="Side-by-side comparison" />
          </div>
        </div>
      </section>
    </div>
  )
}

export default Features
