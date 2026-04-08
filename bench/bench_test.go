package bench

import (
	"net/http/httptest"
	"testing"

	"github.com/godjango/godjango/http/middleware"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
	"github.com/godjango/godjango/template"
	"github.com/godjango/godjango/orm"
	"github.com/godjango/godjango/orm/queryset"
)

// BenchmarkRouter tests the throughput of the URL matching engine.
func BenchmarkRouter(b *testing.B) {
	router := urls.GetGlobalRouter()

	// Create some mock views
	dummyView := func(req *godjangohttp.Request) godjangohttp.Response { return godjangohttp.NewHttpResponse("OK", 200) }

	router.Add(urls.Path("/", dummyView, "home", nil))
	router.Add(urls.Path("/about/", dummyView, "about", nil))
	router.Add(urls.Path("/blog/<slug:slug>/", dummyView, "blog_detail", nil))
	router.Add(urls.Path("/api/v1/users/<int:id>/profile/", dummyView, "api_user_profile", nil))
	router.Add(urls.Path("/products/<str:category>/<int:id>/", dummyView, "product_detail", nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match("/api/v1/users/42/profile/")
	}
}

// BenchmarkTemplateRender tests the performance of the template engine.
func BenchmarkTemplateRender(b *testing.B) {
	tmplStr := `
	<h1>{{ title }}</h1>
	<ul>
	{% for item in items %}
		<li>{{ item.Name }} - {{ item.Price }}</li>
	{% endfor %}
	</ul>
	`
	engine := template.NewEngine([]string{}, false)
	// Mock a loader and register it, or directly parse AST
	// We'll directly use the lexer and parser to get an AST, then render
	lexer := template.NewLexer(tmplStr)
	tokens := lexer.Lex()
	parser := template.NewParser(tokens, engine)
	ast, _ := parser.Parse([]string{})

	type Item struct {
		Name  string
		Price int
	}

	ctxData := map[string]any{
		"title": "Product List",
		"items": []Item{
			{"Widget A", 100},
			{"Widget B", 200},
			{"Widget C", 300},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := template.NewContext(ctxData)
		ast.Render(ctx)
	}
}

// Dummy model for ORM benchmark
type BenchModel struct {
	orm.Model
	Name string `gd:"CharField,max_length=200"`
}

func init() {
	orm.Register(BenchModel{})
}

func BenchmarkQuerySet(b *testing.B) {
	// QuerySet construction and cloning (since DB execution involves disk/network we test the builder)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs := queryset.NewQuerySet[BenchModel]()
		_ = qs.Filter(queryset.Lookup{"name": "test"}).
			Exclude(queryset.Lookup{"id": 5}).
			OrderBy("-name").
			Limit(10).
			Offset(20)
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	// Create a chain of 10 generic middlewares
	dummyMiddleware := func(next middleware.Handler) middleware.Handler {
		return func(req *godjangohttp.Request) godjangohttp.Response {
			req.META["seen"] = "true"
			return next(req)
		}
	}

	middlewares := make([]middleware.MiddlewareFunc, 10)
	for i := 0; i < 10; i++ {
		middlewares[i] = dummyMiddleware
	}

	chain := middleware.NewChain(middlewares...)

	finalHandler := func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.NewHttpResponse("OK", 200)
	}

	chained := chain.Then(finalHandler)

	r := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := godjangohttp.NewRequest(r)
		_ = chained(req)
	}
}
