package queryset

// Lookup is a type alias for keyword arguments in filters.
type Lookup map[string]any

// Connector represents the logical operator joining Q objects.
type Connector string

const (
	AND Connector = "AND"
	OR  Connector = "OR"
)

// QNode represents a node in a boolean query tree.
type QNode struct {
	Connector Connector
	Negated   bool
	Children  []*QNode
	Filters   Lookup
}

// Q creates a new QNode from a set of lookups.
func Q(lookups Lookup) *QNode {
	return &QNode{
		Connector: AND,
		Filters:   lookups,
	}
}

// clone returns a deep copy of the QNode.
func (q *QNode) clone() *QNode {
	if q == nil {
		return nil
	}
	n := &QNode{
		Connector: q.Connector,
		Negated:   q.Negated,
		Filters:   make(Lookup),
	}
	for k, v := range q.Filters {
		n.Filters[k] = v
	}
	for _, c := range q.Children {
		n.Children = append(n.Children, c.clone())
	}
	return n
}

// combine joins two QNodes with a connector.
func (q *QNode) combine(other *QNode, conn Connector) *QNode {
	if q == nil {
		return other.clone()
	}
	if other == nil {
		return q.clone()
	}

	clone := q.clone()

	// We always wrap both in a new node to avoid modifying shared root state
	// (although we cloned it, tree merging can be complex if not handled carefully).
	return &QNode{
		Connector: conn,
		Children:  []*QNode{clone, other.clone()},
	}
}

// And combines this QNode with another using AND.
func (q *QNode) And(other *QNode) *QNode {
	return q.combine(other, AND)
}

// Or combines this QNode with another using OR.
func (q *QNode) Or(other *QNode) *QNode {
	return q.combine(other, OR)
}

// Not negates the QNode.
func (q *QNode) Not() *QNode {
	clone := q.clone()
	clone.Negated = !clone.Negated
	return clone
}
