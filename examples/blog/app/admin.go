package app

import (
	"github.com/pkshahid/JanGo/admin"
	"github.com/pkshahid/JanGo/orm"
)

type PostAdmin struct {
	admin.ModelAdmin
}

var publishedFilter = &admin.SimpleListFilter{
	FilterTitle:   "Published status",
	ParameterName: "published",
	Lookups: []admin.FilterLookup{
		{Display: "Published", Value: "yes"},
		{Display: "Unpublished", Value: "no"},
	},
	QuerysetFn: func(val string, info *orm.ModelInfo) (string, []any) {
		if val == "yes" {
			return "IsPublished = ?", []any{true}
		}
		if val == "no" {
			return "IsPublished = ?", []any{false}
		}
		return "", nil
	},
}

func init() {
	admin.DefaultAdminSite.Register(Post{}, &admin.ModelAdmin{
		ListDisplay:  []string{"Title", "IsPublished", "CategoryID"},
		SearchFields: []string{"Title", "Content"},
		ListFilter:   []any{"IsPublished", publishedFilter},
	})
	admin.DefaultAdminSite.Register(Category{}, &admin.ModelAdmin{})
	admin.DefaultAdminSite.Register(Tag{}, &admin.ModelAdmin{})
	admin.DefaultAdminSite.Register(Comment{}, &admin.ModelAdmin{})
}
