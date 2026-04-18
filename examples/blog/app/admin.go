package app

import (
	"github.com/pkshahid/JanGo/admin"
)

type PostAdmin struct {
	admin.ModelAdmin
}

func init() {
	admin.DefaultAdminSite.Register(Post{}, &admin.ModelAdmin{
		ListDisplay: []string{"Title", "IsPublished", "CategoryID"},
		SearchFields: []string{"Title", "Content"},
		ListFilter: []string{"IsPublished"},
	})
	admin.DefaultAdminSite.Register(Category{}, &admin.ModelAdmin{})
	admin.DefaultAdminSite.Register(Tag{}, &admin.ModelAdmin{})
	admin.DefaultAdminSite.Register(Comment{}, &admin.ModelAdmin{})
}
