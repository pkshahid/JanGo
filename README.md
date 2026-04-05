# GoDjango

GoDjango is a full Django-equivalent web framework written in Go. It aims to provide the robust, "batteries-included" experience of Django, while leveraging Go's performance, concurrency, and static typing.

## Design Goals
*   **Django-like Experience**: Familiar patterns, directory structure, and APIs for Django developers.
*   **Batteries Included**: Built-in ORM, admin panel, form handling, authentication, templating, and caching.
*   **No Build Required**: Run your project instantly with `go run manage.go <command>`.
*   **Pure Go**: Relies on Go standards and popular native libraries where appropriate.
*   **Developer Friendly**: Excellent developer experience with hot-reloading built-in.

## Architecture

The framework is divided into logical packages mimicking Django's structure:

*   `core/` - Framework kernel (app registry, settings loader, signals)
*   `http/` - Router, request, response, and middleware chain
*   `template/` - Template engine, tags, filters, and context processors
*   `orm/` - Models, querysets, schema management, and migrations
*   `forms/` - Form fields, validation, widgets, and rendering
*   `admin/` - Auto-generated admin UI
*   `auth/` - User model, permissions, and session management
*   `cache/` - Cache framework and pluggable backends (Redis, in-memory)
*   `static/` - Static file handling and `collectstatic` command
*   `management/` - Management command infrastructure (equivalent to `manage.py`)
*   `contrib/` - Optional bundled apps
*   `test/` - Testing utilities and a test client
*   `cmd/godjango-admin/` - Global CLI binary for generating new projects

## Getting Started

To run a project using GoDjango:

```bash
# Initialize a new project (similar to django-admin startproject)
go run cmd/godjango-admin/main.go startproject myproject

# Navigate to your project
cd myproject

# Run the development server
go run manage.go runserver
```

## Management Commands

GoDjango uses Cobra to provide a robust CLI experience. Common commands include:

*   `runserver` - Starts the development server with hot reload.
*   `migrate` - Applies database migrations.
*   `makemigrations` - Creates new migrations based on model changes.
*   `createsuperuser` - Creates an administrative user.
*   `shell` - Opens an interactive Go shell with context loaded.
