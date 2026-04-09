# Management Commands

GoDjango uses `spf13/cobra` to provide a robust CLI experience, similar to Django's `manage.py`. A standard GoDjango project includes a `manage.go` file at the root, which acts as the entrypoint for executing these commands.

## Common Built-in Commands

*   **`runserver`**
    Starts the development server. Often includes hot-reload capabilities for a better developer experience. It can serve static files during development via WhiteNoiseMiddleware.
    ```bash
    go run manage.go runserver
    # Or on a specific port
    go run manage.go runserver 8080
    ```

*   **`makemigrations`**
    Inspects your ORM models and creates new migration files based on the changes detected compared to the current database schema.
    ```bash
    go run manage.go makemigrations
    ```

*   **`migrate`**
    Applies pending database migrations, updating your database schema to match your models.
    ```bash
    go run manage.go migrate
    ```

*   **`createsuperuser`**
    Prompts for credentials to create an administrative user, useful for accessing the admin panel.
    ```bash
    go run manage.go createsuperuser
    ```

*   **`shell`**
    Opens an interactive Go shell (REPL) with your project's context loaded, allowing you to interact with your ORM and other code easily.
    ```bash
    go run manage.go shell
    ```

*   **`collectstatic`**
    Gathers static files from all applications and configured directories into the `STATIC_ROOT` directory. Supports flags like `--noinput`, `--clear`, and `--dry-run`.
    ```bash
    go run manage.go collectstatic
    ```

*   **`clearsessions`**
    Purges expired sessions from the configured session storage backends.
    ```bash
    go run manage.go clearsessions
    ```

*   **`createcachetable`**
    Provisions the database table required if you are using the database cache backend.
    ```bash
    go run manage.go createcachetable
    ```

*   **`check`**
    Inspects the project for deployment readiness, evaluating security settings (e.g., `DEBUG`, `SECRET_KEY`, `ALLOWED_HOSTS`, SSL, cookies) and returning warnings or errors.
    ```bash
    go run manage.go check
    ```

*   **`version`**
    Prints the current version of the GoDjango framework.
    ```bash
    go run manage.go version
    ```

## `godjango-admin` Global Command

Before you have a project, you use the global binary `godjango-admin` to bootstrap one.

*   **`startproject`**
    Initializes a new GoDjango project with the necessary directory structure and default files (like `manage.go`, `settings.go`, `urls.go`).
    ```bash
    go run cmd/godjango-admin/main.go startproject myproject
    ```
