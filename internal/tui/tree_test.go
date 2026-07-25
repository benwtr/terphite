package tui

import (
	"testing"

	"github.com/benwtr/terphite/internal/graphite"
)

func TestFlattenTreeCollapsedByDefault(t *testing.T) {
	root := graphite.BuildTree([]string{"stats.foo.bar", "stats.baz"})
	rows := flattenTree(root, map[string]bool{})
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1 (only top-level 'stats' visible when collapsed)", len(rows))
	}
	if rows[0].node.Name != "stats" {
		t.Errorf("rows[0].node.Name = %q, want stats", rows[0].node.Name)
	}
}

func TestFlattenTreeExpanded(t *testing.T) {
	root := graphite.BuildTree([]string{"stats.foo.bar", "stats.baz"})
	expanded := map[string]bool{"stats": true}
	rows := flattenTree(root, expanded)

	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3 (stats, stats.foo, stats.baz), rows=%+v", len(rows), rows)
	}
	if rows[0].node.Name != "stats" {
		t.Errorf("rows[0] = %q, want stats", rows[0].node.Name)
	}
	// order within a level follows BuildTree's sorted insertion order (baz < foo)
	if rows[1].node.Name != "baz" || rows[2].node.Name != "foo" {
		t.Errorf("children order = [%q, %q], want [baz, foo]", rows[1].node.Name, rows[2].node.Name)
	}
}

func TestFlattenTreeDeepExpansion(t *testing.T) {
	root := graphite.BuildTree([]string{"a.b.c"})
	expanded := map[string]bool{"a": true, "a.b": true}
	rows := flattenTree(root, expanded)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3, rows=%+v", len(rows), rows)
	}
	if rows[2].node.Name != "c" || !rows[2].node.Leaf {
		t.Errorf("rows[2] = %+v, want leaf node 'c'", rows[2])
	}
	if rows[2].depth != 2 {
		t.Errorf("rows[2].depth = %d, want 2", rows[2].depth)
	}
}
