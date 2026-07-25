package graphite

import (
	"sort"
	"strings"
)

// MetricNode is one node in the nested tree built from a flat list of
// dot-delimited metric names (e.g. "stats.foo.bar").
type MetricNode struct {
	Name     string
	Path     string
	Leaf     bool
	Children []*MetricNode
}

// BuildTree builds a nested MetricNode tree from a flat list of metric
// names. The root node's Name is "metrics" and has no Path.
func BuildTree(names []string) *MetricNode {
	root := &MetricNode{Name: "metrics"}
	childIndex := make(map[*MetricNode]map[string]*MetricNode)

	sorted := append([]string(nil), names...)
	sort.Strings(sorted)

	for _, name := range sorted {
		if name == "" {
			continue
		}
		parts := strings.Split(name, ".")
		cur := root
		var pathParts []string
		for i, part := range parts {
			pathParts = append(pathParts, part)
			children, ok := childIndex[cur]
			if !ok {
				children = make(map[string]*MetricNode)
				childIndex[cur] = children
			}
			child, ok := children[part]
			if !ok {
				child = &MetricNode{Name: part, Path: strings.Join(pathParts, ".")}
				children[part] = child
				cur.Children = append(cur.Children, child)
			}
			if i == len(parts)-1 {
				child.Leaf = true
			}
			cur = child
		}
	}
	return root
}
