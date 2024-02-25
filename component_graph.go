// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

// ComponentGraph represents a graph of components organized in layers.
// Each layer can contain one or more components that can depend on components from the previous layers.
// A layer is represented as a map, mapping from component ID to a component. Layer components are assumed
// to not depend on each other and can be operated on concurrently.
//
// This structure is useful for initializing, starting, and stopping components in the correct order,
// ensuring that dependencies are correctly managed.
type ComponentGraph struct {
	components []map[string]Component
}

// NewComponentGraph creates a new instance of ComponentGraph.
// It initializes an empty graph with no components and returns a pointer to it.
// This function is the starting point for building a graph of components by adding layers.
//
// Example:
//
//	 graph := NewComponentGraph().
//	 	.AddLayer({
//			"component-a": componentA,
//	 	})
//		.AddLayer({
//			"component-b": componentB,
//			"component-c": componentC,
//	 	})
//
// This example creates a new component graph and adds two layers to it.
func NewComponentGraph() *ComponentGraph {
	return &ComponentGraph{}
}

// AddLayer adds a new layer of components to the ComponentGraph.
// Each call to AddLayer represents a new level in the graph, with the given components being added as a single layer.
// Components within the same layer are considered to have no dependencies on each other,
// but depend on components from all previous layers.
//
// Parameters:
//
//	components map[string]Component: A mapping from component ID to a component implementation.
//
// Example:
//
//	 graph := NewComponentGraph().
//	 	.AddLayer({
//			"component-a": componentA,
//	 	})
//		.AddLayer({
//			"component-b": componentB,
//			"component-c": componentC,
//	 	})
//
// This example creates a new component graph and adds two layers to it.
func (c *ComponentGraph) AddLayer(components map[string]Component) *ComponentGraph {
	if len(components) > 0 {
		c.components = append(c.components, components)
	}
	return c
}
