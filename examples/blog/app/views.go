package app

import (
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/views/generic"
)

// PostListView is the main blog index.
type PostListView struct {
	generic.ListView[Post]
}

func (v *PostListView) GetQuerySetFunc(req *godjangohttp.Request, mixin *generic.MultipleObjectMixin[Post]) ([]Post, error) {
	// In a real scenario:
	// return queryset.NewQuerySet[Post]().Filter(queryset.Lookup{"is_published": true}).All()
	return []Post{
		{Title: "First Post", Content: "Hello World"},
	}, nil
}

// PostDetailView shows a single post.
type PostDetailView struct {
	generic.DetailView[Post]
}

// CreatePostView allows authenticated users to post.
type CreatePostView struct {
	generic.CreateView[Post]
	generic.LoginRequiredMixin
}
