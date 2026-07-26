package graphite

import "testing"

func findChild(n *MetricNode, name string) *MetricNode {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestBuildTree(t *testing.T) {
	root := BuildTree([]string{"stats.foo.bar", "stats.foo.baz", "stats.qux"})

	stats := findChild(root, "stats")
	if stats == nil {
		t.Fatal("expected top-level 'stats' node")
	}
	if stats.Leaf {
		t.Error("'stats' should not be a leaf")
	}
	if stats.Path != "stats" {
		t.Errorf("stats.Path = %q, want %q", stats.Path, "stats")
	}

	foo := findChild(stats, "foo")
	if foo == nil {
		t.Fatal("expected 'stats.foo' node")
	}
	if foo.Leaf {
		t.Error("'stats.foo' should not be a leaf (it has children)")
	}
	if foo.Path != "stats.foo" {
		t.Errorf("foo.Path = %q, want %q", foo.Path, "stats.foo")
	}

	bar := findChild(foo, "bar")
	if bar == nil || !bar.Leaf {
		t.Fatal("expected leaf node 'stats.foo.bar'")
	}
	if bar.Path != "stats.foo.bar" {
		t.Errorf("bar.Path = %q, want %q", bar.Path, "stats.foo.bar")
	}

	baz := findChild(foo, "baz")
	if baz == nil || !baz.Leaf {
		t.Fatal("expected leaf node 'stats.foo.baz'")
	}

	qux := findChild(stats, "qux")
	if qux == nil || !qux.Leaf {
		t.Fatal("expected leaf node 'stats.qux'")
	}
}

func TestBuildTreeSharedPrefixLeafAndBranch(t *testing.T) {
	// A metric name can be both a leaf itself and a prefix of other metrics
	// (e.g. Graphite tags/aggregation nodes) — the node should be marked as
	// a leaf while still holding its children.
	root := BuildTree([]string{"stats.foo", "stats.foo.bar"})
	stats := findChild(root, "stats")
	foo := findChild(stats, "foo")
	if foo == nil {
		t.Fatal("expected 'stats.foo' node")
	}
	if !foo.Leaf {
		t.Error("'stats.foo' should be marked as a leaf since it was seen as a full metric name")
	}
	if len(foo.Children) != 1 || foo.Children[0].Name != "bar" {
		t.Errorf("expected 'stats.foo' to have child 'bar', got %+v", foo.Children)
	}
}

func TestBuildTreeEmpty(t *testing.T) {
	root := BuildTree(nil)
	if len(root.Children) != 0 {
		t.Errorf("expected no children for empty input, got %d", len(root.Children))
	}
}
