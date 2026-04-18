package generic

import (
	"testing"
)

func TestPaginator(t *testing.T) {
	objects := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := NewPaginator(objects, 3)

	if p.Count != 10 {
		t.Errorf("Expected count 10, got %d", p.Count)
	}
	if p.NumPages != 4 {
		t.Errorf("Expected num pages 4, got %d", p.NumPages)
	}

	page1, err := p.Page(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.ObjectList) != 3 || page1.ObjectList[0] != 1 {
		t.Errorf("Page 1 error: %v", page1.ObjectList)
	}
	if !page1.HasNext() {
		t.Errorf("Expected page 1 to have next")
	}
	if page1.HasPrevious() {
		t.Errorf("Expected page 1 not to have previous")
	}

	page4, err := p.Page(4)
	if err != nil {
		t.Fatal(err)
	}
	if len(page4.ObjectList) != 1 || page4.ObjectList[0] != 10 {
		t.Errorf("Page 4 error: %v", page4.ObjectList)
	}
	if page4.HasNext() {
		t.Errorf("Expected page 4 not to have next")
	}
	if !page4.HasPrevious() {
		t.Errorf("Expected page 4 to have previous")
	}

	_, err = p.Page(5)
	if err == nil {
		t.Errorf("Expected error for out of range page")
	}

	pr := p.PageRange()
	if len(pr) != 4 || pr[3] != 4 {
		t.Errorf("Page range error: %v", pr)
	}
}
