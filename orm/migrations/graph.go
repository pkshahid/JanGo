package migrations

import (
	"fmt"
	"sync"
)

var (
	globalMigrationRegistry = make(map[string]map[string]*Migration)
	registryMu              sync.RWMutex
)

// RegisterMigration adds a migration to the global registry.
func RegisterMigration(m *Migration) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := globalMigrationRegistry[m.App]; !ok {
		globalMigrationRegistry[m.App] = make(map[string]*Migration)
	}
	globalMigrationRegistry[m.App][m.Name] = m
}

// GetRegisteredMigrations returns all migrations currently registered.
func GetRegisteredMigrations() map[string]map[string]*Migration {
	registryMu.RLock()
	defer registryMu.RUnlock()
	// Shallow copy outer map for safety
	clone := make(map[string]map[string]*Migration)
	for app, migrations := range globalMigrationRegistry {
		appClone := make(map[string]*Migration)
		for name, m := range migrations {
			appClone[name] = m
		}
		clone[app] = appClone
	}
	return clone
}

// MigrationGraph represents the dependency graph of migrations.
type MigrationGraph struct {
	Nodes map[string]*Node
}

type Node struct {
	Migration *Migration
	Key       string
	Children  []*Node
	Parents   []*Node
}

func NewMigrationGraph() *MigrationGraph {
	return &MigrationGraph{
		Nodes: make(map[string]*Node),
	}
}

func (g *MigrationGraph) AddNode(m *Migration) {
	key := fmt.Sprintf("%s.%s", m.App, m.Name)
	g.Nodes[key] = &Node{
		Migration: m,
		Key:       key,
		Children:  make([]*Node, 0),
		Parents:   make([]*Node, 0),
	}
}

func (g *MigrationGraph) AddDependency(migrationKey, depKey string) error {
	node, ok := g.Nodes[migrationKey]
	if !ok {
		return fmt.Errorf("migration %s not found", migrationKey)
	}
	depNode, ok := g.Nodes[depKey]
	if !ok {
		return fmt.Errorf("dependency %s not found for %s", depKey, migrationKey)
	}

	node.Parents = append(node.Parents, depNode)
	depNode.Children = append(depNode.Children, node)
	return nil
}

// BuildGraph loads all registered migrations and constructs the dependency graph.
func BuildGraph() (*MigrationGraph, error) {
	graph := NewMigrationGraph()
	migrations := GetRegisteredMigrations()

	for _, appMigrations := range migrations {
		for _, m := range appMigrations {
			graph.AddNode(m)
		}
	}

	for _, appMigrations := range migrations {
		for _, m := range appMigrations {
			nodeKey := fmt.Sprintf("%s.%s", m.App, m.Name)
			for _, dep := range m.Dependencies {
				depKey := fmt.Sprintf("%s.%s", dep.App, dep.Name)
				if err := graph.AddDependency(nodeKey, depKey); err != nil {
					return nil, err
				}
			}
		}
	}
	return graph, nil
}

// ForwardsPlan returns a linearly ordered slice of migrations to apply, based on dependencies.
// A topological sort is performed.
func (g *MigrationGraph) ForwardsPlan() ([]*Migration, error) {
	var ordered []*Migration
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(node *Node) error
	visit = func(node *Node) error {
		if visiting[node.Key] {
			return fmt.Errorf("circular dependency detected involving %s", node.Key)
		}
		if visited[node.Key] {
			return nil
		}
		visiting[node.Key] = true

		for _, parent := range node.Parents {
			if err := visit(parent); err != nil {
				return err
			}
		}

		visiting[node.Key] = false
		visited[node.Key] = true
		ordered = append(ordered, node.Migration)
		return nil
	}

	// In Django, there are root nodes. We can iterate over all nodes and ensure they are visited.
	// Since order of maps in Go is random, sorting keys makes the output deterministic if no dependencies dictate order.
	for _, node := range g.Nodes {
		if !visited[node.Key] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	return ordered, nil
}
