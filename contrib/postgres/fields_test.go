package postgres

import (
	"testing"
	"time"
)

func TestRangeInclusive(t *testing.T) {
	r := NewRange(1, 10)
	if r.Lower == nil || *r.Lower != 1 {
		t.Error("lower should be 1")
	}
	if r.Upper == nil || *r.Upper != 10 {
		t.Error("upper should be 10")
	}
	if !r.LowerIncl || !r.UpperIncl {
		t.Error("bounds should be inclusive")
	}
	s := r.String()
	if s != "[1,10]" {
		t.Errorf("expected [1,10], got %s", s)
	}
}

func TestRangeExclusive(t *testing.T) {
	r := NewRangeExclusive(1, 10)
	if r.LowerIncl || r.UpperIncl {
		t.Error("bounds should be exclusive")
	}
	s := r.String()
	if s != "(1,10)" {
		t.Errorf("expected (1,10), got %s", s)
	}
}

func TestRangeFromOnly(t *testing.T) {
	r := NewRangeFrom(5, true)
	if r.Upper != nil {
		t.Error("upper should be nil")
	}
	s := r.String()
	if s != "[5,)" {
		t.Errorf("expected [5,), got %s", s)
	}
}

func TestRangeToOnly(t *testing.T) {
	r := NewRangeTo(100, false)
	if r.Lower != nil {
		t.Error("lower should be nil")
	}
	s := r.String()
	if s != "(,100)" {
		t.Errorf("expected (,100), got %s", s)
	}
}

func TestRangeEmpty(t *testing.T) {
	var r Range[int]
	if !r.IsEmpty() {
		t.Error("zero-value range should be empty")
	}
	if r.String() != "empty" {
		t.Errorf("expected 'empty', got %s", r.String())
	}
}

func TestIntRangeAlias(t *testing.T) {
	r := IntRange{
		Lower:     ptr(1),
		Upper:     ptr(100),
		LowerIncl: true,
		UpperIncl: true,
	}
	if r.String() != "[1,100]" {
		t.Errorf("expected [1,100], got %s", r.String())
	}
}

func TestBigIntRangeAlias(t *testing.T) {
	r := BigIntRange{
		Lower:     ptr(int64(1)),
		Upper:     ptr(int64(1000000)),
		LowerIncl: true,
		UpperIncl: false,
	}
	if r.String() != "[1,1000000)" {
		t.Errorf("expected [1,1000000), got %s", r.String())
	}
}

func TestNumRangeAlias(t *testing.T) {
	r := NewRange(1.5, 10.5)
	if r.String() != "[1.5,10.5]" {
		t.Errorf("expected [1.5,10.5], got %s", r.String())
	}
}

func TestTsTzRangeAlias(t *testing.T) {
	lower := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	upper := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	r := NewRange(lower, upper)
	s := r.String()
	if s != "[2024-01-01 00:00:00 +0000 UTC,2024-12-31 23:59:59 +0000 UTC]" {
		t.Errorf("unexpected range string: %s", s)
	}
}

func TestHStore(t *testing.T) {
	h := NewHStore("color", "red", "size", "large")
	if h.Get("color") != "red" {
		t.Errorf("expected red, got %s", h.Get("color"))
	}
	if h.Get("size") != "large" {
		t.Errorf("expected large, got %s", h.Get("size"))
	}
	if h.Get("missing") != "" {
		t.Errorf("expected empty string for missing key")
	}

	h.Set("weight", "10kg")
	if h.Get("weight") != "10kg" {
		t.Errorf("expected 10kg, got %s", h.Get("weight"))
	}
}

func TestArrayTypes(t *testing.T) {
	ai := ArrayInt{1, 2, 3}
	if len(ai) != 3 || ai[0] != 1 {
		t.Errorf("unexpected ArrayInt: %v", ai)
	}

	as := ArrayString{"a", "b", "c"}
	if len(as) != 3 || as[1] != "b" {
		t.Errorf("unexpected ArrayString: %v", as)
	}

	ab := ArrayBool{true, false}
	if len(ab) != 2 || !ab[0] {
		t.Errorf("unexpected ArrayBool: %v", ab)
	}

	af := ArrayFloat64{1.1, 2.2}
	if len(af) != 2 || af[0] != 1.1 {
		t.Errorf("unexpected ArrayFloat64: %v", af)
	}

	ai64 := ArrayInt64{1, 2}
	if len(ai64) != 2 || ai64[0] != 1 {
		t.Errorf("unexpected ArrayInt64: %v", ai64)
	}
}

func TestArrayFieldTag(t *testing.T) {
	tag := ArrayFieldTag("IntegerField", ArrayFieldOpts{Size: 3})
	if tag != "ArrayField,base_field=IntegerField,size=3" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = ArrayFieldTag("CharField", ArrayFieldOpts{MaxLength: 50, Null: true})
	if tag != "ArrayField,base_field=CharField,max_length=50,null=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = ArrayFieldTag("DecimalField", ArrayFieldOpts{MaxDigits: 10, DecimalPlaces: 2, Blank: true})
	if tag != "ArrayField,base_field=DecimalField,max_digits=10,decimal_places=2,blank=true" {
		t.Errorf("unexpected tag: %s", tag)
	}
}

func TestRangeFieldTag(t *testing.T) {
	tag := RangeFieldTag("IntegerRangeField", true, false)
	if tag != "IntegerRangeField,null=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = RangeFieldTag("DateTimeRangeField", true, true)
	if tag != "DateTimeRangeField,null=true,blank=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = RangeFieldTag("DateRangeField", false, false)
	if tag != "DateRangeField" {
		t.Errorf("unexpected tag: %s", tag)
	}
}

func TestCIFieldTags(t *testing.T) {
	tag := CICharFieldTag(200, false, false)
	if tag != "CICharField,max_length=200" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = CICharFieldTag(100, true, true)
	if tag != "CICharField,max_length=100,null=true,blank=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = CITextFieldTag(true, false)
	if tag != "CITextField,null=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = CIEmailFieldTag(false, true)
	if tag != "CIEmailField,blank=true" {
		t.Errorf("unexpected tag: %s", tag)
	}
}

func TestHStoreFieldTag(t *testing.T) {
	tag := HStoreFieldTag(true, false)
	if tag != "HStoreField,null=true" {
		t.Errorf("unexpected tag: %s", tag)
	}

	tag = HStoreFieldTag(false, false)
	if tag != "HStoreField" {
		t.Errorf("unexpected tag: %s", tag)
	}
}

func ptr[T any](v T) *T {
	return &v
}
