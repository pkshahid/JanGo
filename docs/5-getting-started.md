# Getting Started

This guide will help you create your first application within a GoDjango project. We assume you have already completed the steps in the [Setup Guide](./4-setup.md).

In Django terminology, a "project" is the entire website, while an "app" is a module that does something specific (like a blog system, a forum, or a user profile manager). A project can contain multiple apps.

## 1. Creating an App

Unlike Django, which uses `manage.py startapp`, in GoDjango you simply create a new folder and package for your app. Let's create a simple `pages` app to handle static-like pages.

```bash
mkdir pages
```

Inside the `pages` directory, create a few standard files:

```bash
touch pages/views.go
touch pages/urls.go
```

## 2. Writing a View

Open `pages/views.go` and create a simple view that returns a string.

```go
package pages

import (
    "net/http"
    "github.com/godjango/godjango/http/request"
    "github.com/godjango/godjango/http/response"
)

func HomePageView(req *request.Request) response.Response {
    return response.HttpResponse("Welcome to my GoDjango Site!", http.StatusOK)
}

func AboutPageView(req *request.Request) response.Response {
    return response.HttpResponse("About us...", http.StatusOK)
}
```

## 3. Configuring App URLs

Next, map these views to specific URLs in `pages/urls.go`.

```go
package pages

import (
    "github.com/godjango/godjango/http/urls"
    "github.com/godjango/godjango/http/views"
)

var UrlPatterns = []urls.Path{
    urls.Path{"", views.Func(HomePageView), "home"},
    urls.Path{"about/", views.Func(AboutPageView), "about"},
}
```

## 4. Connecting App URLs to the Project

Now you need to tell the main project about the URLs defined in your `pages` app.

Open your project's root `urls.go` (e.g., `mywebsite/urls.go`) and include the `pages` URLs.

```go
package mywebsite

import (
    "github.com/godjango/godjango/http/urls"
    "mywebsite/pages" // Import your new app
)

var RootUrlPatterns = []urls.Path{
    // Include the URLs from the pages app
    urls.Include{"", pages.UrlPatterns},
}
```

## 5. Test It Out

Make sure your development server is running:

```bash
go run manage.go runserver
```

*   Visit `http://localhost:8000/` and you should see "Welcome to my GoDjango Site!".
*   Visit `http://localhost:8000/about/` and you should see "About us...".

## 6. Using Templates (Optional but recommended)

Instead of hardcoding HTML strings in your views, you should use templates.

1.  Create a `templates` folder inside your `pages` app: `mkdir -p pages/templates/pages`
2.  Create an `home.html` file inside:

    ```html
    <!-- pages/templates/pages/home.html -->
    <!DOCTYPE html>
    <html>
    <head><title>Home Page</title></head>
    <body>
        <h1>Welcome to my GoDjango Site!</h1>
        <p>This page is rendered using a template.</p>
    </body>
    </html>
    ```

3.  Update your view in `pages/views.go` to render the template:

    ```go
    func HomePageView(req *request.Request) response.Response {
        // Render takes the request, the template path, and an optional context map
        return response.Render(req, "pages/home.html", nil)
    }
    ```

Refresh your browser, and you will see the updated, templated page! You are now ready to explore more advanced features like the ORM and Forms.
