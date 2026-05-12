package graph

// predEntry tracks one way we reached a node during pathfinding.
type predEntry struct {
	from string
	edge *Edge
}

// ShortestPathGuaranteed finds all shortest paths using only guaranteed edges via BFS.
func (g *Graph) ShortestPathGuaranteed(src, dst string) []*PathResult {
	return allShortestBFS(src, dst, g.AdjGuar)
}

// ShortestCycleGuaranteed finds all shortest cycles using only guaranteed edges.
func (g *Graph) ShortestCycleGuaranteed(biome string) []*PathResult {
	return allShortestCycleBFS(biome, g.AdjGuar)
}

func allShortestBFS(src, dst string, adj map[string][]Edge) []*PathResult {
	if src == dst {
		return []*PathResult{{
			Steps:       []PathStep{{Biome: src}},
			TotalHops:   0,
			Probability: 1.0,
		}}
	}

	dist := map[string]int{src: 0}
	preds := map[string][]predEntry{}

	layer := []string{src}
	found := false

	for len(layer) > 0 && !found {
		var next []string
		for _, cur := range layer {
			for i := range adj[cur] {
				e := &adj[cur][i]
				newDist := dist[cur] + 1

				if d, ok := dist[e.To]; !ok {
					dist[e.To] = newDist
					preds[e.To] = []predEntry{{from: cur, edge: e}}
					if e.To == dst {
						found = true
					} else {
						next = append(next, e.To)
					}
				} else if newDist == d {
					preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
				}
			}
		}
		layer = next
	}

	if !found {
		return nil
	}
	return enumeratePaths(src, dst, preds)
}

func allShortestCycleBFS(src string, adj map[string][]Edge) []*PathResult {
	dist := map[string]int{}
	preds := map[string][]predEntry{}

	// Seed from src's neighbors.
	var firstLayer []string
	for i := range adj[src] {
		e := &adj[src][i]
		if e.To == src {
			return []*PathResult{computeStats([]PathStep{
				{Biome: src, Edge: nil},
				{Biome: src, Edge: e},
			})}
		}
		if _, ok := dist[e.To]; !ok {
			dist[e.To] = 1
			preds[e.To] = []predEntry{{from: src, edge: e}}
			firstLayer = append(firstLayer, e.To)
		} else {
			preds[e.To] = append(preds[e.To], predEntry{from: src, edge: e})
		}
	}

	layer := firstLayer
	found := false
	var closers []predEntry

	for len(layer) > 0 {
		var next []string
		for _, cur := range layer {
			for i := range adj[cur] {
				e := &adj[cur][i]

				if e.To == src {
					closers = append(closers, predEntry{from: cur, edge: e})
					found = true
					continue
				}
				if found {
					continue
				}

				newDist := dist[cur] + 1
				if d, ok := dist[e.To]; !ok {
					dist[e.To] = newDist
					preds[e.To] = []predEntry{{from: cur, edge: e}}
					next = append(next, e.To)
				} else if newDist == d {
					preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
				}
			}
		}
		if found {
			break
		}
		layer = next
	}

	if !found {
		return nil
	}

	var results []*PathResult
	for _, cp := range closers {
		for _, steps := range enumeratePathSteps(src, cp.from, preds) {
			full := make([]PathStep, len(steps)+1)
			copy(full, steps)
			full[len(steps)] = PathStep{Biome: src, Edge: cp.edge}
			results = append(results, computeStats(full))
		}
	}
	return results
}

func enumeratePaths(src, dst string, preds map[string][]predEntry) []*PathResult {
	var results []*PathResult
	for _, steps := range enumeratePathSteps(src, dst, preds) {
		results = append(results, computeStats(steps))
	}
	return results
}

func enumeratePathSteps(src, cur string, preds map[string][]predEntry) [][]PathStep {
	if cur == src {
		return [][]PathStep{{{Biome: src, Edge: nil}}}
	}
	entries, ok := preds[cur]
	if !ok {
		return nil
	}
	var all [][]PathStep
	for _, pred := range entries {
		for _, sp := range enumeratePathSteps(src, pred.from, preds) {
			path := make([]PathStep, len(sp)+1)
			copy(path, sp)
			path[len(sp)] = PathStep{Biome: cur, Edge: pred.edge}
			all = append(all, path)
		}
	}
	return all
}

func computeStats(steps []PathStep) *PathResult {
	prob := 1.0
	weight := 0.0
	for _, s := range steps {
		if s.Edge != nil {
			prob *= s.Edge.Probability
			weight += s.Edge.Weight
		}
	}
	return &PathResult{
		Steps:       steps,
		TotalHops:   len(steps) - 1,
		Probability: prob,
		WeightedLen: weight,
	}
}
