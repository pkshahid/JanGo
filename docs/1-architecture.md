# Architecture and Structure

GoDjango is a full Django-equivalent web framework written in Go. It aims to provide the robust, "batteries-included" experience of Django, while leveraging Go's performance, concurrency, and static typing.

## Design Goals

*   **Django-like Experience**: Familiar patterns, directory structure, and APIs for developers coming from Django.
*   **Batteries Included**: Built-in ORM, admin panel, form handling, authentication, templating, and caching.
*   **No Build Required**: Run your project instantly with `go run manage.go <command>`.
*   **Pure Go**: Relies on Go standards and popular native libraries where appropriate.
*   **Developer Friendly**: Excellent developer experience with hot-reloading built-in.

## Framework Structure

The framework is divided into logical packages mimicking Django's structure:

*   **`core/`** - Framework kernel. Handles app registry, settings loader, signals, and central version management.
*   **`http/`** - Router, request, response, and middleware chain. Also includes ASGI equivalents, generic views, exception handling, and WebSockets integration.
*   **`template/`** - Template engine, tags, filters, context processors, and template loading mechanisms (Filesystem, AppDirectories, Cached, Embed).
*   **`orm/`** - Models, querysets, schema management, database backends (SQLite, PostgreSQL, MySQL), and migrations. Implements a robust `QuerySet[T any]` generic API.
*   **`forms/`** - Form fields, validation, widgets, and rendering. Supports `ModelForm` for auto-generating forms from ORM models.
*   **`admin/`** - Auto-generated admin UI to easily manage database records.
*   **`auth/`** - User model (`AbstractUser`), permissions, session management, password hashing, and authentication views.
*   **`cache/`** - Cache framework with pluggable backends (Redis, Memcached, Database, LocMem, Dummy) and view decorators.
*   **`sessions/`** - Session framework with multiple storage backends and HMAC-SHA256 secured JSON encoding.
*   **`static/`** - Static and media file handling, storages (`FileSystemStorage`, `S3Storage`), finders, and `collectstatic` functionality.
*   **`management/`** - Management command infrastructure (equivalent to `manage.py`) utilizing `spf13/cobra`.
*   **`security/`** - Security hardening middleware (CSP, Rate Limit, Allowed Hosts, X-Frame-Options).
*   **`email/`** - Django-like email framework (`EmailMessage`, `EmailMultiAlternatives`) with pluggable backends.
*   **`i18n/`** - Internationalization using Go's `x/text` package with in-memory GNU `.po` parsing and context-based language activation.
*   **`monitoring/`** - OpenTelemetry tracing, Prometheus metrics, and structured logging (`log/slog`).
*   **`debug/`** - Django-equivalent debug toolbar with panels for SQL, templates, headers, profiling, etc.
*   **`test/`** - Django-equivalent testing utilities, including test cases with isolated databases, request factory, and HTTP test client.
*   **`cmd/godjango-admin/`** - Global CLI binary for generating new projects.
