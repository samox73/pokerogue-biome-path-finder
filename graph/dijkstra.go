package graph

import (
	"container/heap"
	"math"
)

const distEps = 1e-9

// ShortestPathWeighted finds all paths minimizing expected transitions using Dijkstra.
func (g *Graph) ShortestPathWeighted(src, dst string) []*PathResult {
	if src == dst {
		return []*PathResult{{
			Steps:       []PathStep{{Biome: src}},
			TotalHops:   0,
			Probability: 1.0,
			WeightedLen: 0,
		}}
	}

	dist := map[string]float64{src: 0}
	preds := map[string][]predEntry{}
	pq := &priorityQueue{{biome: src, dist: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		cur := item.biome

		if item.dist > dist[cur]+distEps {
			continue // stale
		}

		// Stop expanding once we exceed the best distance to dst.
		if d, ok := dist[dst]; ok && item.dist > d+distEps {
			break
		}

		for i := range g.Adj[cur] {
			e := &g.Adj[cur][i]
			newDist := dist[cur] + e.Weight

			if d, ok := dist[e.To]; !ok {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			} else if newDist < d-distEps {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			} else if math.Abs(newDist-d) <= distEps {
				preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil
	}
	return enumeratePaths(src, dst, preds)
}

// ShortestCycleWeighted finds all shortest cycles minimizing expected transitions.
func (g *Graph) ShortestCycleWeighted(biome string) []*PathResult {
	dist := map[string]float64{}
	preds := map[string][]predEntry{}
	pq := &priorityQueue{}
	heap.Init(pq)

	for i := range g.Adj[biome] {
		e := &g.Adj[biome][i]
		if e.To == biome {
			return []*PathResult{computeStats([]PathStep{
				{Biome: biome, Edge: nil},
				{Biome: biome, Edge: e},
			})}
		}
		if d, ok := dist[e.To]; !ok {
			dist[e.To] = e.Weight
			preds[e.To] = []predEntry{{from: biome, edge: e}}
			heap.Push(pq, &pqItem{biome: e.To, dist: e.Weight})
		} else if e.Weight < d-distEps {
			dist[e.To] = e.Weight
			preds[e.To] = []predEntry{{from: biome, edge: e}}
			heap.Push(pq, &pqItem{biome: e.To, dist: e.Weight})
		} else if math.Abs(e.Weight-d) <= distEps {
			preds[e.To] = append(preds[e.To], predEntry{from: biome, edge: e})
		}
	}

	bestCycleDist := math.Inf(1)
	var closers []predEntry

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		cur := item.biome

		if item.dist > dist[cur]+distEps {
			continue
		}

		if item.dist > bestCycleDist+distEps {
			break
		}

		for i := range g.Adj[cur] {
			e := &g.Adj[cur][i]
			newDist := dist[cur] + e.Weight

			if e.To == biome {
				if newDist < bestCycleDist-distEps {
					bestCycleDist = newDist
					closers = []predEntry{{from: cur, edge: e}}
				} else if math.Abs(newDist-bestCycleDist) <= distEps {
					closers = append(closers, predEntry{from: cur, edge: e})
				}
				continue
			}

			if d, ok := dist[e.To]; !ok {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			} else if newDist < d-distEps {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &pqItem{biome: e.To, dist: newDist})
			} else if math.Abs(newDist-d) <= distEps {
				preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
			}
		}
	}

	if math.IsInf(bestCycleDist, 1) {
		return nil
	}

	var results []*PathResult
	for _, cp := range closers {
		for _, steps := range enumeratePathSteps(biome, cp.from, preds) {
			full := make([]PathStep, len(steps)+1)
			copy(full, steps)
			full[len(steps)] = PathStep{Biome: biome, Edge: cp.edge}
			results = append(results, computeStats(full))
		}
	}
	return results
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
