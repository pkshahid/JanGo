package orm_test

import (
	"fmt"

	"github.com/godjango/godjango/orm"
	"github.com/godjango/godjango/orm/queryset"
)

type ExampleModel struct {
	orm.Model
	Title string `gd:"CharField,max_length=200"`
}

// Example of querying the database.
func ExampleQuerySet() {
	orm.Register(ExampleModel{})
	qs := queryset.NewQuerySet[ExampleModel]()
	qs = qs.Filter(queryset.Lookup{"title": "GoDjango"}).OrderBy("-id")

	// Usually you would fetch all or one:
	// results := qs.All()
	// for _, res := range results {
	//     fmt.Println(res.Title)
	// }
	fmt.Println("Filtered QuerySet constructed")
	// Output: Filtered QuerySet constructed
}
