package generic

import (
	"fmt"
	"math"
)

// Paginator handles pagination of generic object lists.
type Paginator[T any] struct {
	ObjectList []T
	PerPage    int
	Count      int
	NumPages   int
}

// NewPaginator creates a new Paginator instance.
func NewPaginator[T any](objectList []T, perPage int) *Paginator[T] {
	if perPage <= 0 {
		perPage = 1 // Prevent division by zero
	}
	count := len(objectList)
	numPages := int(math.Ceil(float64(count) / float64(perPage)))
	if numPages == 0 {
		numPages = 1 // Always have at least one page
	}
	return &Paginator[T]{
		ObjectList: objectList,
		PerPage:    perPage,
		Count:      count,
		NumPages:   numPages,
	}
}

// Page returns a Page object for the given 1-based page number.
func (p *Paginator[T]) Page(number int) (*Page[T], error) {
	if number < 1 {
		return nil, fmt.Errorf("page number less than 1")
	}
	if number > p.NumPages && p.Count > 0 {
		return nil, fmt.Errorf("page number out of range")
	}

	start := (number - 1) * p.PerPage
	end := start + p.PerPage
	if end > p.Count {
		end = p.Count
	}

	var pageList []T
	if p.Count > 0 {
		pageList = p.ObjectList[start:end]
	}

	return &Page[T]{
		ObjectList: pageList,
		Number:     number,
		Paginator:  p,
	}, nil
}

// Page represents a single page of results from a Paginator.
type Page[T any] struct {
	ObjectList []T
	Number     int
	Paginator  *Paginator[T]
}

// HasNext returns true if there is a next page.
func (p *Page[T]) HasNext() bool {
	return p.Number < p.Paginator.NumPages
}

// HasPrevious returns true if there is a previous page.
func (p *Page[T]) HasPrevious() bool {
	return p.Number > 1
}

// NextPageNumber returns the next page number.
func (p *Page[T]) NextPageNumber() int {
	return p.Number + 1
}

// PreviousPageNumber returns the previous page number.
func (p *Page[T]) PreviousPageNumber() int {
	return p.Number - 1
}

// PageRange returns a list of page numbers suitable for pagination links.
func (p *Paginator[T]) PageRange() []int {
	var pr []int
	for i := 1; i <= p.NumPages; i++ {
		pr = append(pr, i)
	}
	return pr
}
