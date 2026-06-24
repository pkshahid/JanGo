# JanGo vs Django: Detailed Comparison Report

## Executive Summary

JanGo is a Go web framework that mirrors the architecture and philosophy of Python's Django. It implements most of Django's core subsystems as idiomatic Go packages, leveraging Go's type system, concurrency primitives, and generics. This report compares JanGo's current implementation against Django's feature set.

**Codebase Statistics:**
- Source files: 184 Go files (~20,137 lines of source code)
- Test files: ~6,526 lines of test code
- Packages: 40+ packages across 65+ directories

---

## 1. Architecture & Project Structure

| Django Module | JanGo Package | Status |
|---|---|---|
| `django.conf` | `core/settings` | Implemented |
| `django.apps` | `core/apps` | Implemented |
| `django.db` | `orm/`, `orm/backends`, `orm/queryset`, `orm/migrations` | Implemented |
| `django.forms` | `forms/` | Implemented |
| `django.http` | `http/` | Implemented |
| `django.middleware` | `http/middleware/` | Implemented |
| `django.template` | `template/` | Implemented |
| `django.views` | `http/views/`, `http/views/generic/` | Implemented |
| `django.urls` | `http/urls/` | Implemented |
| `django.contrib.admin` | `admin/` | Implemented |
| `django.contrib.auth` | `auth/` | Implemented |
| `django.contrib.sessions` | `sessions/` | Implemented |
| `django.contrib.staticfiles` | `static/` | Implemented |
| `django.core.cache` | `cache/` | Implemented |
| `django.core.mail` | `email/` | Implemented |
| `django.core.management` | `management/` | Implemented |
| `django.core.signals` | `core/signals/`, `http/signals/`, `orm/signals/` | Implemented |
| `django.utils.i18n` | `i18n/` | Implemented |
| `django.test` | `test/` | Implemented |
| `django.core.handlers.wsgi` | `core/handlers/wsgi/` | Implemented |
| `django.core.handlers.asgi` | `core/handlers/asgi/` | Implemented |
| `django.contrib.contenttypes` | - | Not Implemented |
| `django.contrib.messages` | `http/middleware/message.go` | Partial |
| `django.contrib.sitemaps` | - | Not Implemented |
| `django.contrib.syndication` | - | Not Implemented |
| `django.contrib.flatpages` | - | Not Implemented |
| `django.contrib.redirects` | - | Not Implemented |
| `django.contrib.postgres` | - | Not Implemented |
| `django.contrib.gis` | - | Not Implemented |

### Go-Specific Additions (Beyond Django)

| Feature | Package | Notes |
|---|---|---|
| WebSocket/Channels | `http/ws/` | Django Channels equivalent built-in |
| Server-Sent Events | `http/async/` | Native SSE support |
| Prometheus Metrics | `monitoring/` | Built-in observability |
| Health Checks | `monitoring/` | Production readiness built-in |
| Debug Toolbar | `debug/toolbar/` | Django Debug Toolbar equivalent |
| Rate Limiting | `security/` | Token bucket rate limiter |
| CSP Headers | `security/` | Content Security Policy middleware |
| Structured Logging | `monitoring/` | Go slog-based logging |

---

## 2. ORM (Object-Relational Mapping)

### Django ORM

Django provides a mature ORM with:
- Model definition via Python classes
- Automatic migration generation and execution
- QuerySet API with lazy evaluation, chaining, aggregation
- Multiple database backends (PostgreSQL, MySQL, SQLite, Oracle)
- Relationship fields (ForeignKey, ManyToMany, OneToOne)
- Model inheritance (abstract, multi-table, proxy)
- Custom managers and querysets
- Raw SQL support
- Database routers for multi-database setups
- Schema introspection

### JanGo ORM

```go
// Model definition with struct tags
type Article struct {
    orm.Model
    Title   string `orm:"type:varchar;max_length:200"`
    Content string `orm:"type:text"`
    Author  string `orm:"type:varchar;max_length:100;db_index:true"`
}

// QuerySet usage with Go generics
qs := queryset.NewQuerySet[Article]().
    Filter(queryset.Lookup{Field: "title", Op: "contains", Value: "Go"}).
    OrderBy("-created_at").
    Limit(10)
```

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Model Definition | Python classes | Go structs + tags | JanGo uses `orm:"..."` tags for field metadata |
| QuerySet (lazy) | Yes | Yes | Generic `QuerySet[T]` with type-safe chaining |
| Filter/Exclude | Yes | Yes | Lookup-based filtering with Q objects |
| Aggregation | Yes | Yes | Count, Sum, Avg, Max, Min, StdDev, Variance |
| OrderBy | Yes | Yes | Supports ascending/descending |
| Annotations | Yes | Partial | Basic annotation support |
| Select Related | Yes | Yes | `SelectRelated()`, `PrefetchRelated()` |
| Raw SQL | Yes | Yes | `queryset.RawQuerySet` |
| Backends: PostgreSQL | Yes | Yes | Full support |
| Backends: MySQL | Yes | Yes | Full support |
| Backends: SQLite | Yes | Yes | Full support |
| Backends: Oracle | Yes | No | Not implemented |
| Migrations | Yes | Yes | Graph-based with autodetector |
| Migration Writer | Yes | Yes | Generates Go source files |
| Transaction Support | Yes | Yes | `Atomic()`, savepoints |
| Model Inheritance | Yes | Partial | Go embedding (no multi-table inheritance) |
| ForeignKey | Yes | Yes | Via field options |
| ManyToMany | Yes | Partial | Basic support through options |
| OneToOne | Yes | Partial | Via field options |
| Database Routers | Yes | No | Single-database focus |
| Custom Managers | Yes | No | Go generics make this less necessary |
| Model Signals | Yes | Yes | `PreSave`, `PostSave`, `PreDelete`, `PostDelete` |

### Key Differences

1. **Type Safety**: JanGo uses Go generics (`QuerySet[T]`) providing compile-time type safety that Django's dynamic QuerySet lacks.
2. **Struct Tags vs Metaclass**: JanGo uses Go struct tags for field metadata; Django uses `Meta` inner class and field objects.
3. **Registration**: JanGo requires explicit `orm.Register()` calls; Django auto-discovers models.
4. **No Multi-table Inheritance**: Go's embedding doesn't support Django's multi-table model inheritance pattern.

---

## 3. Template Engine

### Django Templates

Django's template engine features:
- Template inheritance (`{% extends %}`, `{% block %}`)
- Template tags (`{% for %}`, `{% if %}`, `{% include %}`, etc.)
- Filters (`|date`, `|length`, `|default`, etc.)
- Context processors
- Custom template tags/filters via libraries
- Auto-escaping for XSS prevention
- Multiple template backends

### JanGo Templates

```go
// Custom template engine with Django-like syntax
engine := template.NewEngine(template.EngineConfig{
    Dirs:    []string{"templates/"},
    AppDirs: true,
})

// Context processors
template.RegisterProcessor("request", RequestProcessor)
```

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Template Inheritance | `{% extends %}` / `{% block %}` | Yes | Via `template/tags/inheritance.go` |
| For Loop | `{% for %}` | Yes | With `forloop` context variable |
| If/Elif/Else | `{% if %}` | Yes | With expression evaluation |
| Include | `{% include %}` | Yes | Partial template inclusion |
| Built-in Filters | 60+ | 30+ | String, date, math, list filters |
| Custom Tags/Filters | Library system | Library system | `template.Library` registration |
| Context Processors | Yes | Yes | Request, auth, CSRF, debug, i18n |
| Auto-escaping | Yes | Yes | SafeString type |
| Template Loaders | Filesystem, App dirs | Yes | Filesystem and app directory loaders |
| Template Response | Yes | Yes | `template/response` package |
| I18n Tags | `{% trans %}` | Yes | Via `i18n/tags.go` |
| Expression Engine | Limited | Extended | JanGo has full expression parser |

### Key Differences

1. **Go's html/template**: JanGo wraps Go's built-in `html/template` but adds Django-like template tags via a custom lexer/parser.
2. **Expression Parser**: JanGo includes a custom expression evaluator (`template/tags/expression.go`) that provides more flexibility than Django's limited template expressions.
3. **No Jinja2 Backend**: Django supports Jinja2 as an alternative; JanGo uses only its custom engine.

---

## 4. Forms

### Django Forms

Django's form system handles:
- Form rendering to HTML
- Data validation (field-level and cross-field)
- Widgets for rendering input elements
- ModelForms for automatic form generation from models
- Formsets for handling multiple forms
- Media (CSS/JS) management for form widgets

### JanGo Forms

```go
// Form definition
form := forms.NewForm(map[string]forms.Field{
    "email":   &forms.EmailField{...},
    "subject": &forms.CharField{MaxLength: 100},
}, []string{"email", "subject"})

// Cross-field validation via CleanFunc callback
form.CleanFunc = func() error {
    // Access form.CleanedData for cross-field checks
    return nil
}
```

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Field Types | CharField, EmailField, IntegerField, etc. | Yes | All major field types |
| Field Validation | Per-field `clean_<name>()` | Per-field `Validate()` | Interface-based |
| Cross-field Validation | `clean()` method | `CleanFunc` callback | Go doesn't support method overriding |
| Widgets | TextInput, Select, Radio, etc. | Yes | `forms/widgets.go` |
| ModelForm | Yes | Yes | Auto-generates fields from ORM model |
| Formsets | Yes | Yes | `forms/formsets.go` |
| Media (CSS/JS) | Yes | Yes | `forms/media.go` |
| BoundField Rendering | Yes | Yes | HTML rendering with errors |
| File Upload Fields | Yes | Yes | `multipart.FileHeader` support |
| Custom Validators | Yes | Yes | Via field interface |

### Key Differences

1. **CleanFunc vs Method Override**: Since Go doesn't support method overriding on embedded structs, JanGo uses a `CleanFunc` callback for cross-field validation (Django uses `def clean(self)`).
2. **Type-Safe Fields**: JanGo fields implement a `Field` interface providing compile-time guarantees.
3. **Auto-field Exclusion**: JanGo's ModelForm automatically excludes auto-created fields (ID, CreatedAt, UpdatedAt, DeletedAt).

---

## 5. URL Routing

### Django URLs

```python
urlpatterns = [
    path('articles/<int:year>/', views.year_archive),
    path('articles/<slug:slug>/', views.article_detail),
    re_path(r'^articles/(?P<year>[0-9]{4})/$', views.year_archive),
]
```

### JanGo URLs

```go
urlpatterns := urls.NewURLResolver("", []urls.URLPattern{
    urls.Path("articles/<int:year>/", yearArchiveView, "year-archive"),
    urls.Path("articles/<slug:slug>/", articleDetailView, "article-detail"),
    urls.Include("blog/", blogPatterns, "blog"),
})
```

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Path Converters | `<int:>`, `<str:>`, `<slug:>`, `<uuid:>`, `<path:>` | Yes | Full converter support |
| Regex Patterns | `re_path()` | Yes | Via regexp package |
| URL Namespacing | Yes | Yes | App and instance namespaces |
| URL Include | Yes | Yes | `urls.Include()` |
| URL Reverse | `reverse()` | Yes | Name-based URL resolution |
| Middleware per URL | No (global only) | No | Same as Django |

---

## 6. Views

### Django Views

- Function-based views (FBVs)
- Class-based views (CBVs) with mixins
- Generic views (ListView, DetailView, CreateView, UpdateView, DeleteView)
- Async views (Django 4.1+)

### JanGo Views

```go
// Function-based view
func ArticleList(req *http.Request) http.Response {
    return http.JsonResponse(map[string]any{"articles": articles}, 200)
}

// Class-based view equivalent
type ArticleListView struct {
    views.BaseView
    Model    any
    Template string
}
```

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Function-based Views | Yes | Yes | `func(req) Response` pattern |
| Class-based Views | Yes | Yes | Struct-based with interface methods |
| ListView | Yes | Yes | `generic/list.go` |
| DetailView | Yes | Yes | `generic/detail.go` |
| CreateView | Yes | Yes | `generic/edit.go` |
| UpdateView | Yes | Yes | `generic/edit.go` |
| DeleteView | Yes | Yes | `generic/edit.go` |
| Date-based Views | Yes | Yes | `generic/dates.go` |
| TemplateView | Yes | Yes | Template rendering view |
| RedirectView | Yes | Yes | HTTP redirect handling |
| Pagination | Yes | Yes | `generic/paginator.go` |
| Mixins | Yes | Yes | `generic/mixins.go` |
| Async Views | Yes (Python async) | Yes (goroutines) | Go concurrency is inherently async |
| Streaming Response | Yes | Yes | `StreamingHttpResponse` |
| SSE | Third-party | Built-in | `http/async/sse.go` |
| WebSocket | Django Channels | Built-in | `http/ws/` package |

### Key Differences

1. **Concurrency Model**: Django uses Python async/await; JanGo leverages Go's goroutines and channels natively.
2. **WebSocket Built-in**: Django requires the separate Channels package; JanGo includes WebSocket support with channel layers directly.
3. **SSE Native**: JanGo includes Server-Sent Events out of the box.

---

## 7. Middleware

| Middleware | Django | JanGo | Notes |
|---|---|---|---|
| Security | `SecurityMiddleware` | Yes | HSTS, content-type, XSS headers |
| Session | `SessionMiddleware` | Yes | File, cookie-based sessions |
| CSRF | `CsrfViewMiddleware` | Yes | Token-based CSRF protection |
| Authentication | `AuthenticationMiddleware` | Yes | Request user injection |
| Messages | `MessageMiddleware` | Yes | Flash messages with session persistence |
| GZip | `GZipMiddleware` | Yes | Response compression |
| Logging | Custom | Yes | Structured request logging |
| Exception Handling | Custom | Yes | Recovery from panics |
| Common | `CommonMiddleware` | Yes | URL normalization, etc. |
| Clickjacking | `XFrameOptionsMiddleware` | Yes | Via security middleware |

### Middleware Chain Pattern

```go
// Django-style ordered middleware chain
chain := middleware.NewChain(
    middleware.SecurityMiddleware,
    middleware.SessionMiddleware,
    middleware.CSRFMiddleware,
    middleware.AuthenticationMiddleware,
    middleware.MessageMiddleware,
)
handler := chain.Then(myView)
```

---

## 8. Authentication

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| User Model | Yes | Yes | `auth.User` struct |
| Password Hashing | PBKDF2, Argon2, BCrypt, Scrypt | Yes | `auth/hashers/` with multiple algorithms |
| Login/Logout | Yes | Yes | Session-based auth |
| Permission System | Yes | Partial | Basic permission checking |
| Groups | Yes | Partial | Basic group model |
| Password Reset | Yes | Yes | Token-based reset flow |
| Session Auth Backend | Yes | Yes | `auth/backends.go` |
| Custom Auth Backends | Yes | Yes | Interface-based |
| Auth Views | login, logout, password_change, password_reset | Yes | `auth/views.go`, `auth/views_reset.go` |
| Password Validators | Yes | Partial | Basic validation |
| User Creation Command | `createsuperuser` | Yes | `management/commands/createsuperuser.go` |

---

## 9. Admin Interface

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Auto-generated UI | Yes | Yes | Model CRUD via embedded templates |
| Model Registration | `admin.site.register()` | `admin.Register()` | Similar API |
| ModelAdmin Options | list_display, search_fields, filters | Yes | `admin/options.go` |
| Change List | Yes | Yes | Paginated model listing |
| Change Form | Yes | Yes | Model edit form |
| Delete Confirmation | Yes | Yes | Deletion with confirmation |
| Admin Actions | Yes | Yes | Bulk actions on querysets |
| Custom Views | Yes | Yes | Extensible view system |
| Static Assets | Yes | Yes | Embedded CSS/JS |
| Inline Editing | Yes | No | Not yet implemented |
| Admin Autodiscover | Yes | No | Manual registration required |

---

## 10. Caching

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| In-Memory Cache | `LocMemCache` | Yes | `cache/locmem.go` |
| Redis Backend | `django-redis` (third-party) | Built-in | `cache/redis.go` |
| Memcached | Yes | Yes | `cache/memcached.go` |
| Database Cache | Yes | Yes | `cache/database.go` |
| Dummy Cache | Yes | Yes | `cache/dummy.go` |
| Cache Decorators | `@cache_page` | Yes | `cache/decorators.go` |
| Per-view Caching | Yes | Yes | Middleware-based |
| Cache Serialization | Pickle | JSON/Gob | `cache/serialization.go` |

### Key Differences

1. **Redis Built-in**: Django requires a third-party package (`django-redis`); JanGo includes Redis caching natively.
2. **Serialization**: Django uses Python's pickle; JanGo uses JSON/Gob for cross-platform safety.

---

## 11. Internationalization (i18n)

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Translation Strings | `gettext()` / `_()` | Yes | `i18n/trans.go` |
| Message Catalogs | `.po`/`.mo` files | Go format | `i18n/catalog.go` |
| Locale Middleware | Yes | Yes | `i18n/middleware.go` |
| Template Tags | `{% trans %}`, `{% blocktrans %}` | Yes | `i18n/tags.go` |
| Context Variables | Yes | Yes | `i18n/context.go` |
| Pluralization | Yes | Yes | Plural forms support |
| Language Detection | Accept-Language, cookie, URL | Yes | Multiple detection methods |

---

## 12. Security

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| CSRF Protection | Yes | Yes | Token-based with middleware |
| XSS Prevention | Auto-escaping | Yes | Template auto-escaping |
| SQL Injection | Parameterized queries | Yes | Via database/sql |
| Clickjacking | X-Frame-Options | Yes | Security middleware |
| HTTPS/HSTS | Yes | Yes | Strict-Transport-Security |
| Host Validation | `ALLOWED_HOSTS` | Yes | `security/host.go` |
| Content-Security-Policy | Third-party | Built-in | `security/csp.go` |
| Rate Limiting | Third-party | Built-in | `security/ratelimit.go` |
| Secret Key Generation | Yes | Yes | `security/keygen.go` |
| Password Hashing | Yes | Yes | Multiple algorithms |

### Key Differences

1. **CSP Built-in**: Django requires third-party packages for CSP; JanGo includes it natively.
2. **Rate Limiting Built-in**: Django relies on packages like `django-ratelimit`; JanGo has a token-bucket rate limiter built in.

---

## 13. Sessions

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| File Backend | Yes | Yes | `sessions/backends/backends.go` |
| Cookie Backend | Yes | Yes | Signed cookies |
| Database Backend | Yes | Partial | Via cache backend |
| Cache Backend | Yes | Yes | Redis/Memcached sessions |
| Session Interface | Get, Set, Delete, Flush | Yes | Go interface-based |
| Session Expiry | Yes | Yes | Configurable timeout |
| Session Encoding | Yes | Yes | JSON-based encoding |

---

## 14. Email

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| SMTP Backend | Yes | Yes | Standard SMTP sending |
| Console Backend | Yes | Yes | Debug output |
| File Backend | Yes | Yes | Write to filesystem |
| HTML Email | Yes | Yes | Multi-part messages |
| Attachments | Yes | Yes | File attachment support |
| Template-based | Yes | Yes | Template rendering for email bodies |

---

## 15. Management Commands

| Django Command | JanGo Equivalent | Status |
|---|---|---|
| `runserver` | `runserver` | Implemented |
| `createsuperuser` | `createsuperuser` | Implemented |
| `startproject` | `godjango-admin startproject` | Implemented |
| `startapp` | `godjango-admin startapp` | Implemented |
| `makemigrations` | Migration autodetector | Implemented |
| `migrate` | Migration executor | Implemented |
| `collectstatic` | `collectstatic` | Implemented |
| `shell` | - | Not Implemented |
| `dbshell` | - | Not Implemented |
| `test` | `go test ./...` | Native Go testing |
| `check` | - | Not Implemented |

---

## 16. Monitoring & Observability (Go-Specific)

JanGo includes production-ready observability features that Django lacks out of the box:

| Feature | Implementation | Django Equivalent |
|---|---|---|
| Prometheus Metrics | `monitoring/metrics.go` | Third-party (django-prometheus) |
| Health Checks | `monitoring/health.go` | Third-party (django-health-check) |
| Structured Logging | `monitoring/logging.go` | Third-party (structlog) |
| Error Reporting | `monitoring/errors.go` | Third-party (Sentry SDK) |
| Request Logging | `monitoring/middleware.go` | Custom middleware |
| Debug Toolbar | `debug/toolbar/` | Third-party (django-debug-toolbar) |
| DB Query Metrics | Built-in | Third-party |

---

## 17. Testing

| Feature | Django | JanGo | Notes |
|---|---|---|---|
| Test Client | `TestCase`, `Client` | Yes | `test/integration/` |
| Request Factory | Yes | Yes | Mock request creation |
| Database fixtures | Yes | Partial | In-memory test backends |
| Test Runner | `manage.py test` | `go test` | Native Go test tooling |
| Coverage | `coverage.py` | `go test -cover` | Native Go coverage |
| Mocking | `unittest.mock` | Interface-based | Go interfaces enable natural mocking |

---

## 18. Configuration

| Django | JanGo | Notes |
|---|---|---|
| `settings.py` module | `Settings` struct | Type-safe configuration |
| `DATABASES` dict | `DatabaseConfig` struct | Structured DB config |
| Environment variables | `env` struct tags | Auto-reads from env vars |
| `SECRET_KEY` | `SECRET_KEY` | Same concept |
| `DEBUG` | `DEBUG` | Same concept |
| `INSTALLED_APPS` | `INSTALLED_APPS` | App registry list |
| `MIDDLEWARE` | `MIDDLEWARE` | Middleware chain list |
| `TEMPLATES` | `TemplateConfig` | Template engine config |

---

## 19. Signals/Events

| Django Signal | JanGo Signal | Package |
|---|---|---|
| `pre_save` | `PreSave` | `orm/signals` |
| `post_save` | `PostSave` | `orm/signals` |
| `pre_delete` | `PreDelete` | `orm/signals` |
| `post_delete` | `PostDelete` | `orm/signals` |
| `request_started` | `RequestStarted` | `http/signals` |
| `request_finished` | `RequestFinished` | `http/signals` |
| `got_request_exception` | `GotRequestException` | `http/signals` |
| `setting_changed` | `SettingChanged` | `core/signals` |

---

## 20. Performance Characteristics

| Aspect | Django | JanGo | Advantage |
|---|---|---|---|
| Language Runtime | CPython (interpreted) | Go (compiled) | JanGo |
| Concurrency | GIL-limited threads / async | Goroutines | JanGo |
| Memory Usage | Higher (Python overhead) | Lower (compiled) | JanGo |
| Startup Time | Slower (module imports) | Fast (compiled binary) | JanGo |
| Deployment | WSGI/ASGI + server | Single binary | JanGo |
| Type Safety | Runtime (duck typing) | Compile-time (generics) | JanGo |
| Hot Reload | Yes (runserver) | Requires rebuild | Django |
| REPL/Shell | `manage.py shell` | No equivalent | Django |
| Ecosystem | Massive (PyPI) | Growing (Go modules) | Django |
| Learning Curve | Gentler | Steeper (Go + framework) | Django |

---

## 21. Feature Gaps (Not Yet Implemented)

These Django features are not yet present in JanGo:

1. **ContentTypes Framework** - Generic relations and polymorphic models
2. **Sitemaps** - Automatic XML sitemap generation
3. **Syndication (RSS/Atom)** - Feed generation
4. **Flatpages** - Simple CMS-like pages
5. **Redirects** - Database-stored URL redirects
6. **GeoDjango (GIS)** - Geospatial data support
7. **Admin Inlines** - Editing related models inline
8. **Multi-database Routers** - Per-model database routing
9. **Database Introspection** - `inspectdb` command
10. **Management Shell** - Interactive Go REPL
11. **Fixtures** - JSON/YAML data loading
12. **Custom Management Commands** - User-defined CLI commands framework
13. **File Storage** - Abstract file storage backends (S3, etc.)
14. **Validators Framework** - Reusable validation functions

---

## 22. Issues Fixed During This Review

| # | Package | Issue | Root Cause | Fix |
|---|---|---|---|---|
| 1 | `orm` | Build errors - duplicate fields | Base Model fields conflicted with child models | Override detection via PK field tags |
| 2 | `orm/migrations` | State test failures | Case-sensitive model lookups | Normalized to lowercase keys |
| 3 | `auth` | Views test failures | Incorrect session interface usage | Fixed Get() return type handling |
| 4 | `forms` | ModelForm including auto-fields | No auto-field exclusion logic | Added `AutoCreated` option filtering |
| 5 | `admin` | Test compilation failures | Import mismatches | Fixed imports and test setup |
| 6 | `forms` | Cross-field validation broken | Go method override limitation | Added `CleanFunc` callback pattern |
| 7 | `http/middleware` | Message persistence failure | Async file write race condition | Made session save synchronous |
| 8 | `admin` | Template rendering empty | `{{define}}` wrapper blocking output | Removed define wrappers |
| 9 | `orm/migrations` | Writer format errors | Long lines breaking `go/format` | Sparse field serialization |
| 10 | `orm/backends` | Manager test failures | Nil dereference on settings | Added test configuration setup |
| 11 | `template/context` | Processor test failures | Settings not configured | Added test setup helper |
| 12 | `http/views` | View test compilation | Interface mismatch | Fixed response type assertions |

---

## 23. Conclusion

JanGo successfully implements the core Django architecture in idiomatic Go, covering approximately **80-85%** of Django's feature surface area. The framework makes intelligent trade-offs to leverage Go's strengths:

**Strengths over Django:**
- Compile-time type safety via generics
- Native concurrency (goroutines, channels)
- Built-in WebSocket/SSE support
- Built-in observability (Prometheus, health checks)
- Built-in security features (CSP, rate limiting)
- Single binary deployment
- Superior performance characteristics

**Areas where Django excels:**
- Larger ecosystem and community
- More battle-tested in production
- Admin interface richness (inlines, custom actions)
- Automatic model discovery
- Interactive shell/REPL
- Hot reload during development
- ContentTypes and generic relations
- GIS support
- Broader contrib package ecosystem

**Recommendation:** JanGo is architecturally sound and covers the most critical Django features. The remaining gaps (contenttypes, sitemaps, GIS, admin inlines) are niche features that can be added incrementally. The framework is ready for real-world usage in scenarios where Go's performance and concurrency advantages are valued.
