package queryset

import (
	"reflect"
	"testing"
)

func TestQ(t *testing.T) {
	q1 := Q(Lookup{"a": 1})
	q2 := Q(Lookup{"b": 2})

	// AND
	qAnd := q1.And(q2)
	if qAnd.Connector != AND {
		t.Errorf("Expected AND connector, got %v", qAnd.Connector)
	}
	if len(qAnd.Children) != 2 {
		t.Errorf("Expected 2 children, got %v", len(qAnd.Children))
	}
	if !reflect.DeepEqual(qAnd.Children[0].Filters, Lookup{"a": 1}) {
		t.Errorf("Expected child 0 to be {a: 1}")
	}

	// OR
	qOr := q1.Or(q2)
	if qOr.Connector != OR {
		t.Errorf("Expected OR connector, got %v", qOr.Connector)
	}

	// NOT
	qNot := q1.Not()
	if !qNot.Negated {
		t.Errorf("Expected Negated to be true")
	}

	// Chaining Q(a=1).Or(Q(b=2)).And(Q(c=3))
	q3 := Q(Lookup{"c": 3})
	chained := q1.Or(q2).And(q3)
	if chained.Connector != AND {
		t.Errorf("Root connector should be AND")
	}
	if chained.Children[0].Connector != OR {
		t.Errorf("First child connector should be OR")
	}
	if len(chained.Children) != 2 {
		t.Errorf("Expected 2 children on root")
	}
}
