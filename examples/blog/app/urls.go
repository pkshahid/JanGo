package app

import (
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
)

func init() {
	r := urls.GetGlobalRouter()

	list := &PostListView{}
	list.PaginateBy = 10
	list.TemplateName = "post_list.html"

	detail := &PostDetailView{}
	detail.TemplateName = "post_detail.html"

	create := &CreatePostView{}
	create.TemplateName = "post_form.html"
	create.SuccessUrl = "/"

	loginView := func(req *godjangohttp.Request) godjangohttp.Response {
		return godjangohttp.Render(req, "login.html", nil)
	}

	// Manually adapt generic views to ViewFunc since generic type dispatching requires explicit bounds
	r.Add(urls.Path("/", func(req *godjangohttp.Request) godjangohttp.Response { return list.Get(req) }, "home", nil))
	r.Add(urls.Path("/post/<int:pk>/", func(req *godjangohttp.Request) godjangohttp.Response { return detail.Get(req) }, "post_detail", nil))
	r.Add(urls.Path("/post/new/", func(req *godjangohttp.Request) godjangohttp.Response { return create.Get(req) }, "post_create", nil))
	r.Add(urls.Path("/login/", loginView, "login", nil))
}
