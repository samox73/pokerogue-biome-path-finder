package graph

import "container/heap"

// ShortestPathWeighted finds the path minimizing expected transitions using Dijkstra.
// Edge weights are 1/probability. Returns nil if no path exists.
func (g *Graph) ShortestPathWeighted(src, dst string) *PathResult {
	if src == dst {
		return &PathResult{
			Steps:       []PathStep{{Biome: src}},
			TotalHops:   0,
			Probability: 1.0,
			WeightedLen: 0,
		}
	}

	dist := map[string]float64{src: 0}
	prev := map[string]*pathPrev{}
	pq := &priorityQueue{{biome: src, dist: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		cur := item.biome

		if cur == dst {
			break
		}

		if item.dist > dist[cur] {
			continue // stale entry
		}

		for i := range g.Adj[cur] {
			e := &g.Adj[cur][i]
			newDist := dist[cur] + e.Weight
			if d, ok := dist[e.To]; !ok || newDist < d {
				dist[e.To] = newDist
				prev[e.To] = &pathPrev{biome: cur, edge: e}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil
	}

	return buildPath(src, dst, prev)
}

// ShortestCycleWeighted finds the shortest cycle from biome back to itself
// using all edges, minimizing expected transitions. Returns nil if no cycle exists.
func (g *Graph) ShortestCycleWeighted(biome string) *PathResult {
	// Seed Dijkstra from biome's neighbors (not biome itself)
	// so we find a real cycle of at least 1 hop.
	dist := map[string]float64{}
	prev := map[string]*pathPrev{}
	pq := &priorityQueue{}
	heap.Init(pq)

	for i := range g.Adj[biome] {
		e := &g.Adj[biome][i]
		if e.To == biome {
			// Self-loop.
			return computeStats([]PathStep{
				{Biome: biome, Edge: nil},
				{Biome: biome, Edge: e},
			})
		}
		if d, ok := dist[e.To]; !ok || e.Weight < d {
			dist[e.To] = e.Weight
			prev[e.To] = &pathPrev{biome: biome, edge: e}
			heap.Push(pq, &pqItem{biome: e.To, dist: e.Weight})
		}
	}

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		cur := item.biome

		if item.dist > dist[cur] {
			continue // stale entry
		}

		for i := range g.Adj[cur] {
			e := &g.Adj[cur][i]
			newDist := dist[cur] + e.Weight

			if e.To == biome {
				// Found a cycle back to start.
				prev[biome] = &pathPrev{biome: cur, edge: e}
				return buildCyclePath(biome, prev)
			}

			if d, ok := dist[e.To]; !ok || newDist < d {
				dist[e.To] = newDist
				prev[e.To] = &pathPrev{biome: cur, edge: e}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			}
		}
	}
	return nil
}

// Priority queue implementation for Dijkstra.

type pqItem struct {
	biome string
	dist  float64
	index int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].dist < pq[j].dist }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*pqItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}
