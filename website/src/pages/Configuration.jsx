import CodeBlock from '../components/CodeBlock'
import { Settings, Database, Shield, Globe, Mail, Layers, Server, Eye } from 'lucide-react'
import './Configuration.css'

function Configuration() {
  const configSections = [
    {
      id: 'core',
      icon: <Settings size={22} />,
      title: 'Core Settings',
      desc: 'Essential project configuration options.',
      settings: [
        { name: 'DEBUG', type: 'bool', default: 'false', desc: 'Enable debug mode. Shows detailed error pages and enables the debug toolbar. Never use in production.' },
        { name: 'SECRET_KEY', type: 'string', default: '""', desc: 'Cryptographic secret for signing sessions, CSRF tokens, and other security features. Must be unique and kept secret.' },
        { name: 'ALLOWED_HOSTS', type: '[]string', default: '[]', desc: 'List of hostnames the server will respond to. Use ["*"] in development, specific domains in production.' },
        { name: 'ROOT_URLCONF', type: '*[]urls.Path', default: 'nil', desc: 'Pointer to the root URL patterns slice. Defines all URL routing for the project.' },
        { name: 'INSTALLED_APPS', type: '[]string', default: '[]', desc: 'List of app import paths to include. Order matters for template resolution and migrations.' },
        { name: 'LANGUAGE_CODE', type: 'string', default: '"en-us"', desc: 'Default language for the project. Used by the i18n system.' },
        { name: 'TIME_ZONE', type: 'string', default: '"UTC"', desc: 'Default timezone for the project.' },
      ]
    },
    {
      id: 'database',
      icon: <Database size={22} />,
      title: 'Database Configuration',
      desc: 'Configure one or more database connections.',
      code: `DATABASES: map[string]orm.DatabaseConfig{
    "default": {
        Engine:   "sqlite3",      // "sqlite3", "postgresql", "mysql"
        Name:     "db.sqlite3",   // Database name or file path
        User:     "",             // Database user (not needed for SQLite)
        Password: "",             // Database password
        Host:     "localhost",    // Database host
        Port:     "5432",         // Database port
    },
}`,
      settings: [
        { name: 'Engine', type: 'string', default: '"sqlite3"', desc: 'Database backend: "sqlite3", "postgresql", or "mysql".' },
        { name: 'Name', type: 'string', default: '""', desc: 'Database name. For SQLite, this is the file path.' },
        { name: 'User', type: 'string', default: '""', desc: 'Database username. Not required for SQLite.' },
        { name: 'Password', type: 'string', default: '""', desc: 'Database password. Not required for SQLite.' },
        { name: 'Host', type: 'string', default: '"localhost"', desc: 'Database server hostname or IP.' },
        { name: 'Port', type: 'string', default: '""', desc: 'Database server port (e.g., "5432" for PostgreSQL).' },
      ]
    },
    {
      id: 'middleware',
      icon: <Layers size={22} />,
      title: 'Middleware',
      desc: 'Configure the request/response processing pipeline.',
      code: `MIDDLEWARE: []string{
    "SecurityMiddleware",        // HSTS, X-Content-Type-Options
    "WhiteNoiseMiddleware",     // Static file serving
    "SessionMiddleware",        // Session management
    "CSRFMiddleware",           // CSRF protection
    "AuthenticationMiddleware", // Load user from session
    "GZipMiddleware",           // Response compression
    "CacheMiddleware",          // Full-page caching
    "LocaleMiddleware",         // i18n language detection
}`,
      settings: [
        { name: 'SecurityMiddleware', type: 'built-in', default: '-', desc: 'Adds security headers: HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy.' },
        { name: 'WhiteNoiseMiddleware', type: 'built-in', default: '-', desc: 'Serves static files directly from the application. Great for single-binary deployments.' },
        { name: 'SessionMiddleware', type: 'built-in', default: '-', desc: 'Enables session support. Must come before AuthenticationMiddleware.' },
        { name: 'CSRFMiddleware', type: 'built-in', default: '-', desc: 'Validates CSRF tokens on POST/PUT/DELETE requests.' },
        { name: 'AuthenticationMiddleware', type: 'built-in', default: '-', desc: 'Loads the authenticated user from the session into req.User.' },
        { name: 'GZipMiddleware', type: 'built-in', default: '-', desc: 'Compresses responses using gzip for clients that support it.' },
        { name: 'CacheMiddleware', type: 'built-in', default: '-', desc: 'Enables full-page response caching.' },
      ]
    },
    {
      id: 'templates',
      icon: <Globe size={22} />,
      title: 'Template Configuration',
      desc: 'Configure template engine, directories, and loaders.',
      code: `TEMPLATES: []core.TemplateConfig{
    {
        Backend: "godjango",        // Template engine backend
        Dirs:    []string{          // Template search directories
            "templates",
        },
        AppDirs: true,              // Search app-level template/ dirs
        Options: map[string]interface{}{
            "context_processors": []string{
                "auth",             // Adds 'user' and 'perms'
                "request",          // Adds 'request' object
                "messages",         // Adds 'messages' framework
                "static",          // Adds 'STATIC_URL'
            },
        },
    },
}`,
      settings: [
        { name: 'Backend', type: 'string', default: '"godjango"', desc: 'Template engine to use. Currently "godjango" (Django-compatible syntax).' },
        { name: 'Dirs', type: '[]string', default: '[]', desc: 'List of directories to search for templates (in order).' },
        { name: 'AppDirs', type: 'bool', default: 'false', desc: 'If true, also searches templates/ directories inside each installed app.' },
        { name: 'Options', type: 'map[string]interface{}', default: '{}', desc: 'Engine-specific options like context processors and caching.' },
      ]
    },
    {
      id: 'static',
      icon: <Server size={22} />,
      title: 'Static & Media Files',
      desc: 'Configure static asset handling and media uploads.',
      code: `// Static files (CSS, JS, images)
STATIC_URL:  "/static/",
STATIC_ROOT: "staticfiles",         // collectstatic output
STATICFILES_DIRS: []string{         // Additional static dirs
    "assets",
},

// Media files (user uploads)
MEDIA_URL:  "/media/",
MEDIA_ROOT: "media",`,
      settings: [
        { name: 'STATIC_URL', type: 'string', default: '"/static/"', desc: 'URL prefix for serving static files.' },
        { name: 'STATIC_ROOT', type: 'string', default: '"staticfiles"', desc: 'Directory where collectstatic places gathered files.' },
        { name: 'STATICFILES_DIRS', type: '[]string', default: '[]', desc: 'Additional directories to search for static files.' },
        { name: 'MEDIA_URL', type: 'string', default: '"/media/"', desc: 'URL prefix for media file uploads.' },
        { name: 'MEDIA_ROOT', type: 'string', default: '"media"', desc: 'Filesystem path for storing uploaded media files.' },
      ]
    },
    {
      id: 'cache',
      icon: <Server size={22} />,
      title: 'Cache Configuration',
      desc: 'Configure caching backends for improved performance.',
      code: `CACHES: map[string]core.CacheConfig{
    "default": {
        Backend:  "redis",          // redis, memcached, db, locmem, dummy
        Location: "redis://localhost:6379/0",
        Options: map[string]interface{}{
            "KEY_PREFIX":   "myapp",
            "TIMEOUT":     300,     // Default TTL in seconds
            "MAX_ENTRIES":  1000,
        },
    },
}`,
      settings: [
        { name: 'Backend', type: 'string', default: '"locmem"', desc: 'Cache backend: "redis", "memcached", "db", "locmem", or "dummy".' },
        { name: 'Location', type: 'string', default: '""', desc: 'Backend-specific connection string (Redis URL, Memcached host:port, etc.).' },
        { name: 'KEY_PREFIX', type: 'string', default: '""', desc: 'Prefix for all cache keys. Useful for multi-app environments.' },
        { name: 'TIMEOUT', type: 'int', default: '300', desc: 'Default cache entry TTL in seconds. Set to 0 for no expiry.' },
      ]
    },
    {
      id: 'security',
      icon: <Shield size={22} />,
      title: 'Security Settings',
      desc: 'Harden your application for production deployment.',
      code: `// Security settings for production
SECURE_SSL_REDIRECT:          true,
SECURE_HSTS_SECONDS:          31536000,
SECURE_HSTS_INCLUDE_SUBDOMAINS: true,
SECURE_CONTENT_TYPE_NOSNIFF:  true,
SESSION_COOKIE_SECURE:        true,
CSRF_COOKIE_SECURE:           true,
X_FRAME_OPTIONS:              "DENY",

// Rate limiting
RATE_LIMIT_ENABLED:           true,
RATE_LIMIT_REQUESTS:          100,
RATE_LIMIT_WINDOW:            "1m",`,
      settings: [
        { name: 'SECURE_SSL_REDIRECT', type: 'bool', default: 'false', desc: 'Redirect all HTTP requests to HTTPS.' },
        { name: 'SECURE_HSTS_SECONDS', type: 'int', default: '0', desc: 'HTTP Strict Transport Security header max-age value.' },
        { name: 'SESSION_COOKIE_SECURE', type: 'bool', default: 'false', desc: 'Only send session cookies over HTTPS.' },
        { name: 'CSRF_COOKIE_SECURE', type: 'bool', default: 'false', desc: 'Only send CSRF cookies over HTTPS.' },
        { name: 'X_FRAME_OPTIONS', type: 'string', default: '"SAMEORIGIN"', desc: 'Controls whether the site can be embedded in frames.' },
        { name: 'RATE_LIMIT_REQUESTS', type: 'int', default: '100', desc: 'Maximum requests allowed per rate limit window.' },
      ]
    },
    {
      id: 'email',
      icon: <Mail size={22} />,
      title: 'Email Configuration',
      desc: 'Configure email sending backends and SMTP settings.',
      code: `// SMTP Email backend
EMAIL_BACKEND:  "smtp",
EMAIL_HOST:     "smtp.gmail.com",
EMAIL_PORT:     587,
EMAIL_USE_TLS:  true,
EMAIL_HOST_USER:     "your-email@gmail.com",
EMAIL_HOST_PASSWORD: "your-app-password",
DEFAULT_FROM_EMAIL:  "noreply@example.com",

// For development, use console backend
// EMAIL_BACKEND: "console",`,
      settings: [
        { name: 'EMAIL_BACKEND', type: 'string', default: '"console"', desc: 'Email backend: "smtp", "console", "filebased", or custom.' },
        { name: 'EMAIL_HOST', type: 'string', default: '"localhost"', desc: 'SMTP server hostname.' },
        { name: 'EMAIL_PORT', type: 'int', default: '25', desc: 'SMTP server port (587 for TLS, 465 for SSL).' },
        { name: 'EMAIL_USE_TLS', type: 'bool', default: 'false', desc: 'Enable TLS encryption for SMTP.' },
        { name: 'DEFAULT_FROM_EMAIL', type: 'string', default: '""', desc: 'Default "From" address for emails.' },
      ]
    },
    {
      id: 'monitoring',
      icon: <Eye size={22} />,
      title: 'Monitoring & Logging',
      desc: 'Configure observability: tracing, metrics, and structured logging.',
      code: `// OpenTelemetry tracing
TRACING_ENABLED:  true,
TRACING_EXPORTER: "stdout",    // "stdout", "jaeger", "otlp"
TRACING_ENDPOINT: "localhost:4317",

// Prometheus metrics
METRICS_ENABLED:  true,
METRICS_PATH:     "/metrics",

// Structured logging (slog)
LOGGING: core.LoggingConfig{
    Level:   "INFO",           // DEBUG, INFO, WARN, ERROR
    Format:  "json",           // "json" or "text"
    Output:  "stdout",
},`,
      settings: [
        { name: 'TRACING_ENABLED', type: 'bool', default: 'false', desc: 'Enable OpenTelemetry distributed tracing.' },
        { name: 'TRACING_EXPORTER', type: 'string', default: '"stdout"', desc: 'Trace export destination: "stdout", "jaeger", or "otlp".' },
        { name: 'METRICS_ENABLED', type: 'bool', default: 'false', desc: 'Enable Prometheus metrics endpoint.' },
        { name: 'LOGGING.Level', type: 'string', default: '"INFO"', desc: 'Minimum log level: DEBUG, INFO, WARN, ERROR.' },
        { name: 'LOGGING.Format', type: 'string', default: '"text"', desc: 'Log output format: "json" for structured, "text" for human-readable.' },
      ]
    },
  ]

  return (
    <div className="config-page">
      <section className="page-header">
        <div className="container">
          <span className="badge badge-blue">Reference</span>
          <h1>Configuration</h1>
          <p>
            Complete reference for all JanGo configuration options.
            Configure your project in <code>settings.go</code> using <code>core.Configure()</code>.
          </p>
        </div>
      </section>

      <div className="container config-layout">
        {/* Sidebar */}
        <aside className="config-sidebar">
          <nav className="config-nav">
            <h4>Sections</h4>
            {configSections.map(s => (
              <a key={s.id} href={`#${s.id}`} className="config-nav-link">
                {s.icon} {s.title}
              </a>
            ))}
          </nav>
        </aside>

        {/* Content */}
        <div className="config-content">
          {configSections.map(section => (
            <section key={section.id} id={section.id} className="config-section">
              <div className="config-section-header">
                <div className="config-section-icon">{section.icon}</div>
                <div>
                  <h2>{section.title}</h2>
                  <p>{section.desc}</p>
                </div>
              </div>

              {section.code && (
                <CodeBlock code={section.code} language="go" title={`${section.title} Example`} />
              )}

              <div className="settings-table">
                <div className="settings-table-header">
                  <span>Setting</span>
                  <span>Type</span>
                  <span>Default</span>
                  <span>Description</span>
                </div>
                {section.settings.map((setting, i) => (
                  <div key={i} className="settings-table-row">
                    <code className="setting-name">{setting.name}</code>
                    <code className="setting-type">{setting.type}</code>
                    <code className="setting-default">{setting.default}</code>
                    <span className="setting-desc">{setting.desc}</span>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Configuration
