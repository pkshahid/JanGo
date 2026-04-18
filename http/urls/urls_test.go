package urls

import (
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

func dummyView(req *godjangohttp.Request) godjangohttp.Response {
	return nil
}

func TestParseRoutePattern(t *testing.T) {
	re, convMap := parseRoutePattern("articles/<int:year>/<slug:slug>/", false)

	if convMap["year"] != "int" || convMap["slug"] != "slug" {
		t.Errorf("Converters mapping failed: %v", convMap)
	}

	match := re.FindStringSubmatch("articles/2023/hello-world/")
	if match == nil {
		t.Errorf("Regex match failed")
	}

	names := re.SubexpNames()
	if names[1] != "year" || names[2] != "slug" {
		t.Errorf("Regex subexp names incorrect: %v", names)
	}
}

func TestRouterMatch(t *testing.T) {
	router := &Router{}

	router.Add(Path("articles/<int:year>/<slug:slug>/", dummyView, "article_detail", map[string]any{"extra_val": 42}))
	router.Add(Path("about/", dummyView, "about", nil))

	match, err := router.Match("articles/2023/my-post/")
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if match.Kwargs["year"] != "2023" {
		t.Errorf("Expected year 2023, got %v", match.Kwargs["year"])
	}
	if match.Kwargs["slug"] != "my-post" {
		t.Errorf("Expected slug my-post, got %v", match.Kwargs["slug"])
	}
	if match.Kwargs["extra_val"] != 42 {
		t.Errorf("Expected extra_val 42, got %v", match.Kwargs["extra_val"])
	}
	if match.URLName != "article_detail" {
		t.Errorf("Expected name article_detail, got %s", match.URLName)
	}
}

func TestIncludeAndNamespace(t *testing.T) {
	router := &Router{}

	appConf := &URLconf{
		Patterns: []*URLPattern{
			Path("detail/<int:id>/", dummyView, "detail", nil),
		},
		Namespace: "myapp",
	}

	router.Add(Include("app/", appConf))

	match, err := router.Match("app/detail/123/")
	if err != nil {
		t.Fatalf("Include match failed: %v", err)
	}

	if match.Namespace != "myapp" {
		t.Errorf("Expected namespace myapp, got %s", match.Namespace)
	}
	if match.URLName != "myapp:detail" {
		t.Errorf("Expected URLName myapp:detail, got %s", match.URLName)
	}
	if match.Kwargs["id"] != "123" {
		t.Errorf("Expected id 123, got %v", match.Kwargs["id"])
	}
}

func TestReverse(t *testing.T) {
	router := &Router{}

	router.Add(Path("articles/<int:year>/<slug:slug>/", dummyView, "article_detail", nil))

	appConf := &URLconf{
		Patterns: []*URLPattern{
			Path("detail/<int:id>/", dummyView, "detail", nil),
		},
		Namespace: "myapp",
	}
	router.Add(Include("app/", appConf))

	urlStr, err := router.Reverse("article_detail", map[string]any{"year": 2023, "slug": "test-post"})
	if err != nil {
		t.Fatalf("Reverse failed: %v", err)
	}
	if urlStr != "articles/2023/test-post/" {
		t.Errorf("Reverse output mismatch: %s", urlStr)
	}

	urlStrNamespace, err := router.Reverse("myapp:detail", map[string]any{"id": 42})
	if err != nil {
		t.Fatalf("Reverse with namespace failed: %v", err)
	}
	if urlStrNamespace != "app/detail/42/" {
		t.Errorf("Reverse output mismatch for namespace: %s", urlStrNamespace)
	}
}

func TestGlobalReverse(t *testing.T) {
	globalRouter.Add(Path("test/<str:name>/", dummyView, "test_global", nil))

	urlStr := Reverse("test_global", map[string]any{"name": "hello"})
	if urlStr != "/test/hello/" {
		t.Errorf("Global Reverse output mismatch: %s", urlStr)
	}
}
