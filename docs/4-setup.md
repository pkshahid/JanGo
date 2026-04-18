# Setup and Installation

This guide will walk you through setting up GoDjango on your local machine.

## Prerequisites

Before installing GoDjango, you must have the following installed:

1.  **Go (Golang)**: Go version 1.18 or higher is required due to the framework's use of Go Generics.
    *   Download and install from [golang.org/dl](https://golang.org/dl/).
    *   Verify the installation by running `go version` in your terminal.

2.  **Git**: Required for fetching the framework and its dependencies.

## Installation

GoDjango provides a global CLI tool called `godjango-admin` that is used to bootstrap new projects.

1.  **Install the `godjango-admin` tool globally:**

    ```bash
    go install github.com/pkshahid/JanGo/cmd/godjango-admin@latest
    ```

    *Note: Ensure that your `GOPATH/bin` directory is added to your system's `PATH` environment variable so you can run the command from anywhere.*

2.  **Verify the installation:**

    ```bash
    godjango-admin version
    ```

    You should see the current version of the GoDjango framework printed to your console.

## Creating Your First Project

Once the CLI tool is installed, you can create a new project. Let's call it `mywebsite`.

```bash
godjango-admin startproject mywebsite
```

This will create a new directory called `mywebsite` with the initial project structure, which looks like this:

```text
mywebsite/
├── manage.go          # The main entrypoint for management commands
├── go.mod             # Go module definition
├── go.sum             # Go dependencies checksums
└── mywebsite/         # The core application package
    ├── settings.go    # Project configuration (database, middleware, apps)
    └── urls.go        # Root URL routing
```

## Running the Development Server

1.  Navigate into your newly created project directory:

    ```bash
    cd mywebsite
    ```

2.  Run the development server using the `manage.go` script:

    ```bash
    go run manage.go runserver
    ```

3.  Open your web browser and navigate to `http://localhost:8000`. You should see the default GoDjango welcome page, indicating that your setup was successful!

## Database Configuration

By default, GoDjango configures projects to use SQLite, which requires no external database setup. The database file will be created automatically when you run migrations.

To apply the initial migrations for built-in apps (like authentication and sessions):

```bash
go run manage.go migrate
```

You are now ready to start building your application! Head over to the **Getting Started** guide to create your first app.
