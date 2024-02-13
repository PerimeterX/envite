// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

// ComponentGraph represents a graph of components organized in layers.
// Each layer can contain one or more components that can depend on components from the previous layers.
// This structure is useful for initializing, starting, and stopping components in the correct order,
// ensuring that dependencies are correctly managed.
type ComponentGraph struct {
	components [][]Component
}

// NewComponentGraph creates a new instance of ComponentGraph.
// It initializes an empty graph with no components and returns a pointer to it.
// This function is the starting point for building a graph of components by adding layers.
//
// Example:
//
//	 graph := NewComponentGraph().
//	 	.AddLayer(componentA)
//		.AddLayer(componentB, componentC)
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
//	layerComponents ...Component: A variadic parameter that accepts one or more components to be added as a layer.
//
// Example:
//
//	 graph := NewComponentGraph().
//	 	.AddLayer(componentA)
//		.AddLayer(componentB, componentC)
//
// This example creates a new component graph and adds two layers to it.
func (c *ComponentGraph) AddLayer(layerComponents ...Component) *ComponentGraph {
	if len(layerComponents) > 0 {
		c.components = append(c.components, layerComponents)
	}
	return c
}
