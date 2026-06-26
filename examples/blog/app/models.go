package app

import (
	"github.com/pkshahid/JanGo/http/urls"
	"github.com/pkshahid/JanGo/orm"
)

type Category struct {
	orm.Model
	Name string `gd:"CharField,max_length=100"`
	Slug string `gd:"SlugField,max_length=100"`
}

type Tag struct {
	orm.Model
	Name string `gd:"CharField,max_length=50"`
}

type Post struct {
	orm.Model
	Title       string `gd:"CharField,max_length=200"`
	Content     string `gd:"TextField"`
	CategoryID  int    `gd:"ForeignKey,to=Category"`
	IsPublished bool   `gd:"BooleanField,default=false"`
	// Tags []Tag `gd:"ManyToManyField"`
}

// GetAbsoluteURL returns the canonical URL for this post, using the
// "post_detail" URL pattern registered in urls.go.
func (p *Post) GetAbsoluteURL() string {
	return urls.Reverse("post_detail", map[string]any{"pk": p.Model.ID})
}

type Comment struct {
	orm.Model
	PostID int    `gd:"ForeignKey,to=Post"`
	Author string `gd:"CharField,max_length=100"`
	Body   string `gd:"TextField"`
}

func init() {
	orm.Register(Category{}, Tag{}, Post{}, Comment{})
}
