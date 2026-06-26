package main

// todoListTemplate is the main page template showing all todos with add/toggle/delete forms.
const todoListTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo App - JanGo</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        .container {
            max-width: 680px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            text-align: center;
            padding: 40px 0 20px;
        }
        header h1 {
            font-size: 2.5em;
            color: #2c3e50;
            margin-bottom: 5px;
        }
        header p {
            color: #7f8c8d;
            font-size: 1.1em;
        }
        .stats {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin: 20px 0;
        }
        .stat {
            background: white;
            padding: 12px 24px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
        }
        .stat-number {
            font-size: 1.8em;
            font-weight: bold;
            color: #2c3e50;
        }
        .stat-label {
            font-size: 0.85em;
            color: #7f8c8d;
            text-transform: uppercase;
        }
        .add-form {
            display: flex;
            gap: 10px;
            margin: 30px 0 20px;
        }
        .add-form input[type="text"] {
            flex: 1;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            transition: border-color 0.3s;
        }
        .add-form input[type="text"]:focus {
            outline: none;
            border-color: #3498db;
        }
        .add-form button {
            padding: 12px 24px;
            background: #3498db;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1em;
            cursor: pointer;
            transition: background 0.3s;
        }
        .add-form button:hover {
            background: #2980b9;
        }
        .todo-list {
            list-style: none;
            margin-top: 10px;
        }
        .todo-item {
            display: flex;
            align-items: center;
            gap: 12px;
            background: white;
            padding: 16px;
            margin-bottom: 8px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            transition: transform 0.2s;
        }
        .todo-item:hover {
            transform: translateX(4px);
        }
        .todo-item.completed .todo-title {
            text-decoration: line-through;
            color: #95a5a6;
        }
        .todo-title {
            flex: 1;
            font-size: 1.05em;
        }
        .todo-date {
            font-size: 0.8em;
            color: #95a5a6;
        }
        .btn {
            padding: 6px 12px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.85em;
            transition: opacity 0.3s;
        }
        .btn:hover {
            opacity: 0.8;
        }
        .btn-toggle {
            background: #27ae60;
            color: white;
        }
        .btn-toggle.undo {
            background: #f39c12;
        }
        .btn-delete {
            background: #e74c3c;
            color: white;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #95a5a6;
        }
        .empty-state p {
            font-size: 1.2em;
            margin-bottom: 10px;
        }
        footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            color: #95a5a6;
            font-size: 0.9em;
        }
        footer a {
            color: #3498db;
            text-decoration: none;
        }
        footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Todo App</h1>
            <p>Built with JanGo Framework</p>
        </header>

        <div class="stats">
            <div class="stat">
                <div class="stat-number">{{.total}}</div>
                <div class="stat-label">Total</div>
            </div>
            <div class="stat">
                <div class="stat-number">{{.completed}}</div>
                <div class="stat-label">Done</div>
            </div>
            <div class="stat">
                <div class="stat-number">{{.pending}}</div>
                <div class="stat-label">Pending</div>
            </div>
        </div>

        <form class="add-form" method="POST" action="/add/">
            <input type="hidden" name="csrfmiddlewaretoken" value="{{.csrf_token}}">
            <input type="text" name="title" placeholder="What needs to be done?" required>
            <button type="submit">Add</button>
        </form>

        {{if .todos}}
        <ul class="todo-list">
            {{range .todos}}
            <li class="todo-item {{if .Completed}}completed{{end}}">
                <form method="POST" action="/toggle/{{.ID}}/" style="display:inline;">
                    <input type="hidden" name="csrfmiddlewaretoken" value="{{$.csrf_token}}">
                    <button type="submit" class="btn btn-toggle {{if .Completed}}undo{{end}}">
                        {{if .Completed}}Undo{{else}}Done{{end}}
                    </button>
                </form>
                <span class="todo-title">{{.Title}}</span>
                <span class="todo-date">{{.CreatedAt.Format "Jan 2, 15:04"}}</span>
                <form method="POST" action="/delete/{{.ID}}/" style="display:inline;">
                    <input type="hidden" name="csrfmiddlewaretoken" value="{{$.csrf_token}}">
                    <button type="submit" class="btn btn-delete">Delete</button>
                </form>
            </li>
            {{end}}
        </ul>
        {{else}}
        <div class="empty-state">
            <p>No todos yet!</p>
            <p>Add your first task above.</p>
        </div>
        {{end}}

        <footer>
            <p>Powered by <a href="/about/">JanGo</a> &mdash; A Django-like framework for Go</p>
        </footer>
    </div>
</body>
</html>`

// todoAboutTemplate is the about page for the todo app.
const todoAboutTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>About - Todo App</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        .container {
            max-width: 680px;
            margin: 0 auto;
            padding: 40px 20px;
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 20px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            margin-bottom: 10px;
        }
        p {
            margin-bottom: 15px;
            color: #555;
        }
        ul {
            margin: 10px 0 20px 20px;
        }
        li {
            margin-bottom: 8px;
            color: #555;
        }
        code {
            background: #ecf0f1;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.9em;
        }
        a {
            color: #3498db;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .back-link {
            display: inline-block;
            margin-top: 30px;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            border-radius: 8px;
        }
        .back-link:hover {
            background: #2980b9;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>About This App</h1>
        <p>This is a Todo web application built using the <strong>JanGo</strong> framework
           &mdash; a Django-like web framework written in pure Go.</p>

        <h2>Features Demonstrated</h2>
        <ul>
            <li><strong>URL Routing</strong> &mdash; Django-style path converters (<code>/toggle/&lt;int:id&gt;/</code>)</li>
            <li><strong>Template Rendering</strong> &mdash; Go templates with context data</li>
            <li><strong>Form Handling</strong> &mdash; POST request processing for CRUD operations</li>
            <li><strong>Redirect Responses</strong> &mdash; PRG pattern after form submissions</li>
            <li><strong>Management Commands</strong> &mdash; <code>go run main.go runserver</code></li>
        </ul>

        <h2>How to Run</h2>
        <p>From the <code>examples/todo</code> directory:</p>
        <p><code>go run . runserver</code></p>
        <p>Then visit <a href="http://localhost:8000">http://localhost:8000</a> in your browser.</p>

        <a href="/" class="back-link">Back to Todos</a>
    </div>
</body>
</html>`
