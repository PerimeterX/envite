package envite

type ComponentGraph struct {
	components [][]Component
}

func NewComponentGraph() *ComponentGraph {
	return &ComponentGraph{}
}

func (c *ComponentGraph) AddLayer(layerComponents ...Component) *ComponentGraph {
	if len(layerComponents) > 0 {
		c.components = append(c.components, layerComponents)
	}
	return c
}
