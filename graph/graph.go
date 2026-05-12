package graph

// Edge represents a directed edge in the biome graph.
type Edge struct {
	From        string
	To          string
	Probability float64
	Weight      float64 // 1.0 / Probability
}

// PathStep is one step along a found path.
type PathStep struct {
	Biome string
	Edge  *Edge // nil for the starting biome
}

// PathResult holds a complete path and its statistics.
type PathResult struct {
	Steps       []PathStep
	TotalHops   int
	Probability float64 // product of all edge probabilities
	WeightedLen float64 // sum of all edge weights (expected transitions)
}

// Graph holds two adjacency maps: all edges and guaranteed-only edges.
type Graph struct {
	Biomes  []string
	Adj     map[string][]Edge // all edges
	AdjGuar map[string][]Edge // only edges with probability == 1.0
}

// New builds a Graph from the raw biome data.
func New() *Graph {
	g := &Graph{
		Biomes:  Biomes,
		Adj:     make(map[string][]Edge, len(Biomes)),
		AdjGuar: make(map[string][]Edge, len(Biomes)),
	}

	for from, edges := range adjacencyRaw {
		for _, re := range edges {
			e := Edge{
				From:        from,
				To:          re.to,
				Probability: re.prob,
				Weight:      1.0 / re.prob,
			}
			g.Adj[from] = append(g.Adj[from], e)
			if re.prob == 1.0 {
				g.AdjGuar[from] = append(g.AdjGuar[from], e)
			}
		}
	}

	return g
}
