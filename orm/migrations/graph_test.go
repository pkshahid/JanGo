package migrations

import (
	"reflect"
	"testing"
)

func TestGraph(t *testing.T) {
	// Clear registry
	globalMigrationRegistry = make(map[string]map[string]*Migration)

	m1 := &Migration{App: "app", Name: "0001"}
	m2 := &Migration{App: "app", Name: "0002", Dependencies: []Dep{{"app", "0001"}}}
	m3 := &Migration{App: "other", Name: "0001", Dependencies: []Dep{{"app", "0001"}}}
	m4 := &Migration{App: "app", Name: "0003", Dependencies: []Dep{{"app", "0002"}, {"other", "0001"}}}

	RegisterMigration(m1)
	RegisterMigration(m2)
	RegisterMigration(m3)
	RegisterMigration(m4)

	graph, err := BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph error: %v", err)
	}

	plan, err := graph.ForwardsPlan()
	if err != nil {
		t.Fatalf("ForwardsPlan error: %v", err)
	}

	if len(plan) != 4 {
		t.Errorf("Expected 4 migrations in plan, got %d", len(plan))
	}

	if plan[0].Name != "0001" || plan[0].App != "app" {
		t.Errorf("First migration should be app 0001, got %s %s", plan[0].App, plan[0].Name)
	}

	// Verify m4 is last because it depends on all others
	if plan[3].Name != "0003" || plan[3].App != "app" {
		t.Errorf("Last migration should be app 0003, got %s %s", plan[3].App, plan[3].Name)
	}

	// Circular dependency test
	globalMigrationRegistry = make(map[string]map[string]*Migration)
	mc1 := &Migration{App: "app", Name: "0001", Dependencies: []Dep{{"app", "0002"}}}
	mc2 := &Migration{App: "app", Name: "0002", Dependencies: []Dep{{"app", "0001"}}}

	RegisterMigration(mc1)
	RegisterMigration(mc2)

	cGraph, err := BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph error for circular deps: %v", err)
	}
	_, err = cGraph.ForwardsPlan()
	if err == nil {
		t.Errorf("Expected circular dependency error")
	}
}
