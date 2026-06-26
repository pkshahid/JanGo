import { Target, Rocket, Users, Heart, Code2, Gauge } from 'lucide-react'
import './About.css'

function About() {
  const principles = [
    {
      icon: <Target size={28} />,
      title: 'Django-like Experience',
      desc: 'Familiar patterns, directory structure, and APIs for developers coming from Django. If you know Django, you already know JanGo.'
    },
    {
      icon: <Rocket size={28} />,
      title: 'Batteries Included',
      desc: 'Built-in ORM, admin panel, form handling, authentication, templating, caching, sessions, email, and more. No assembly required.'
    },
    {
      icon: <Code2 size={28} />,
      title: 'No Build Required',
      desc: 'Run your project instantly with `go run manage.go runserver`. No webpack, no bundler, no transpiler. Just Go.'
    },
    {
      icon: <Gauge size={28} />,
      title: 'Pure Go Performance',
      desc: 'Leverages Go\'s compiled performance, goroutine-based concurrency, and static typing for production-grade applications.'
    },
    {
      icon: <Users size={28} />,
      title: 'Developer Friendly',
      desc: 'Excellent developer experience with hot-reloading, clear error messages, comprehensive documentation, and a familiar project structure.'
    },
    {
      icon: <Heart size={28} />,
      title: 'Open Source',
      desc: 'JanGo is fully open source and community-driven. Contribute, report issues, or simply use it for your next project.'
    },
  ]

  const architecture = [
    { pkg: 'core/', desc: 'Framework kernel — app registry, settings loader, signals, version management' },
    { pkg: 'http/', desc: 'Router, request/response, middleware chain, WebSockets, generic views' },
    { pkg: 'template/', desc: 'Template engine with inheritance, tags, filters, and multiple loaders' },
    { pkg: 'orm/', desc: 'Models, QuerySets, schema management, migrations, multi-database backends' },
    { pkg: 'forms/', desc: 'Form fields, validation, widgets, rendering, ModelForm generation' },
    { pkg: 'admin/', desc: 'Auto-generated admin UI for CRUD operations on any registered model' },
    { pkg: 'auth/', desc: 'User model, permissions, session management, password hashing' },
    { pkg: 'cache/', desc: 'Pluggable backends: Redis, Memcached, Database, LocMem, Dummy' },
    { pkg: 'sessions/', desc: 'Multiple storage backends with HMAC-SHA256 secured JSON encoding' },
    { pkg: 'security/', desc: 'CSP, rate limiting, allowed hosts, X-Frame-Options, CSRF protection' },
    { pkg: 'email/', desc: 'EmailMessage, EmailMultiAlternatives with pluggable backends' },
    { pkg: 'i18n/', desc: 'Internationalization with GNU .po parsing and language activation' },
    { pkg: 'monitoring/', desc: 'OpenTelemetry tracing, Prometheus metrics, structured logging' },
    { pkg: 'static/', desc: 'Static/media file handling, S3 storage, finders, collectstatic' },
    { pkg: 'test/', desc: 'Testing utilities, isolated databases, request factory, HTTP client' },
  ]

  return (
    <div className="about-page">
      <section className="page-header">
        <div className="container">
          <span className="badge badge-blue">About the Framework</span>
          <h1>Why JanGo?</h1>
          <p>
            JanGo brings the proven architecture and developer productivity of Django
            to the Go ecosystem, delivering enterprise-grade features with native performance.
          </p>
        </div>
      </section>

      {/* Mission */}
      <section className="section">
        <div className="container">
          <div className="about-mission">
            <div className="mission-text">
              <h2>Our Mission</h2>
              <p>
                Django has proven itself as one of the most productive web frameworks ever created.
                Its &quot;batteries-included&quot; philosophy lets developers focus on building applications,
                not infrastructure.
              </p>
              <p>
                JanGo brings this same philosophy to Go. We believe you shouldn&apos;t have to choose
                between developer productivity and runtime performance. With JanGo, you get both &mdash;
                a familiar, comprehensive framework that compiles to a single binary and handles
                massive concurrent workloads effortlessly.
              </p>
              <p>
                Every component is designed to feel natural to both Django developers and Go developers.
                The same patterns, the same mental models, but with type safety, compiled performance,
                and goroutine-based concurrency built in from day one.
              </p>
            </div>
            <div className="mission-comparison">
              <div className="comparison-card">
                <h4>Django (Python)</h4>
                <ul>
                  <li>Interpreted, dynamic typing</li>
                  <li>GIL limits concurrency</li>
                  <li>Requires WSGI/ASGI server</li>
                  <li>Extensive ecosystem</li>
                  <li>Batteries included</li>
                </ul>
              </div>
              <div className="comparison-card highlight">
                <h4>JanGo (Go)</h4>
                <ul>
                  <li>Compiled, static typing</li>
                  <li>Native goroutine concurrency</li>
                  <li>Single binary deployment</li>
                  <li>Growing ecosystem</li>
                  <li>Batteries included</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Design Principles */}
      <section className="section" style={{ background: 'var(--navy-800)' }}>
        <div className="container">
          <div className="section-header-center">
            <h2 className="section-title">Design Principles</h2>
            <p className="section-subtitle">
              The core values that guide every design decision in JanGo.
            </p>
          </div>
          <div className="grid-3">
            {principles.map((p, i) => (
              <div key={i} className="card principle-card">
                <div className="principle-icon">{p.icon}</div>
                <h3>{p.title}</h3>
                <p>{p.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Architecture */}
      <section className="section">
        <div className="container">
          <div className="section-header-center">
            <h2 className="section-title">Framework Architecture</h2>
            <p className="section-subtitle">
              JanGo is divided into logical packages that mirror Django&apos;s proven structure.
            </p>
          </div>
          <div className="architecture-list">
            {architecture.map((item, i) => (
              <div key={i} className="arch-item">
                <code className="arch-pkg">{item.pkg}</code>
                <span className="arch-desc">{item.desc}</span>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}

export default About
