package main

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"

	"biome-path-finder/graph"
)

// ---------------------------------------------------------------------------
// Multi-predecessor BFS (guaranteed edges only)
// ---------------------------------------------------------------------------

// allShortestBFS finds ALL shortest paths from src to dst using only the given
// adjacency list. It tracks multiple predecessors per node when ties occur,
// then enumerates every distinct shortest path.
func allShortestBFS(src, dst string, adj map[string][]graph.Edge) [][]string {
	if src == dst {
		return nil // cycles handled separately
	}

	type predInfo struct {
		from string
		edge *graph.Edge
	}

	// depth[node] = BFS depth at which node was first reached
	depth := map[string]int{src: 0}
	// preds[node] = all predecessors that reach node at its optimal depth
	preds := map[string][]predInfo{}

	queue := []string{src}
	found := false
	targetDepth := -1

	for len(queue) > 0 && !found {
		// Process entire current layer
		var nextQueue []string
		curDepth := depth[queue[0]]

		for _, cur := range queue {
			if depth[cur] != curDepth {
				// This node belongs to the next layer; put it back
				nextQueue = append(nextQueue, cur)
				continue
			}

			for i := range adj[cur] {
				e := &adj[cur][i]
				newDepth := curDepth + 1

				if d, visited := depth[e.To]; visited {
					// Already reached at depth d
					if newDepth == d {
						// Same depth: add alternative predecessor
						preds[e.To] = append(preds[e.To], predInfo{from: cur, edge: e})
					}
					// If newDepth > d, skip (worse)
				} else {
					// First time reaching this node
					depth[e.To] = newDepth
					preds[e.To] = append(preds[e.To], predInfo{from: cur, edge: e})
					if e.To == dst {
						found = true
						targetDepth = newDepth
					}
					nextQueue = append(nextQueue, e.To)
				}
			}
		}
		queue = nextQueue
	}

	if !found {
		return nil
	}
	_ = targetDepth

	// Enumerate all paths by backtracking from dst through preds
	var results [][]string
	var backtrack func(node string, path []string)
	backtrack = func(node string, path []string) {
		path = append([]string{node}, path...)
		if node == src {
			cp := make([]string, len(path))
			copy(cp, path)
			results = append(results, cp)
			return
		}
		for _, p := range preds[node] {
			backtrack(p.from, path)
		}
	}
	backtrack(dst, nil)
	return results
}

// allShortestCycleBFS finds ALL shortest cycles from src back to src using
// only the given adjacency list.
func allShortestCycleBFS(src string, adj map[string][]graph.Edge) [][]string {
	type predInfo struct {
		from string
	}

	// Check for self-loops first
	for i := range adj[src] {
		if adj[src][i].To == src {
			return [][]string{{src, src}}
		}
	}

	depth := map[string]int{}
	preds := map[string][]predInfo{}

	// Seed with src's neighbors
	var queue []string
	for i := range adj[src] {
		e := &adj[src][i]
		if e.To == src {
			continue
		}
		if _, ok := depth[e.To]; !ok {
			depth[e.To] = 1
			preds[e.To] = []predInfo{{from: src}}
			queue = append(queue, e.To)
		} else if depth[e.To] == 1 {
			preds[e.To] = append(preds[e.To], predInfo{from: src})
		}
	}

	// Nodes from which we can close the cycle back to src at the best depth
	var closingNodes []string
	found := false

	for len(queue) > 0 && !found {
		var nextQueue []string
		curDepth := depth[queue[0]]

		for _, cur := range queue {
			if depth[cur] != curDepth {
				nextQueue = append(nextQueue, cur)
				continue
			}

			for i := range adj[cur] {
				e := &adj[cur][i]
				newDepth := curDepth + 1

				if e.To == src {
					if !found {
						found = true
					}
					closingNodes = append(closingNodes, cur)
					continue
				}

				if d, visited := depth[e.To]; visited {
					if newDepth == d {
						preds[e.To] = append(preds[e.To], predInfo{from: cur})
					}
				} else {
					depth[e.To] = newDepth
					preds[e.To] = []predInfo{{from: cur}}
					nextQueue = append(nextQueue, e.To)
				}
			}
		}
		queue = nextQueue
	}

	if !found {
		return nil
	}

	// Deduplicate closing nodes
	seen := map[string]bool{}
	var uniqueClosing []string
	for _, n := range closingNodes {
		if !seen[n] {
			seen[n] = true
			uniqueClosing = append(uniqueClosing, n)
		}
	}

	// Backtrack from each closing node to src through preds, building full cycle paths
	var results [][]string
	var backtrack func(node string, path []string)
	backtrack = func(node string, path []string) {
		newPath := make([]string, len(path)+1)
		newPath[0] = node
		copy(newPath[1:], path)

		if node == src {
			results = append(results, newPath)
			return
		}
		for _, p := range preds[node] {
			backtrack(p.from, newPath)
		}
	}

	for _, closer := range uniqueClosing {
		// The cycle is: src -> ... -> closer -> src
		// Backtrack from closer to src, then append src at the end
		backtrack(closer, []string{src})
	}

	return dedup(results)
}

// ---------------------------------------------------------------------------
// Multi-predecessor Dijkstra (all edges, weighted)
// ---------------------------------------------------------------------------

type dijkItem struct {
	biome string
	dist  float64
	index int
}

type dijkPQ []*dijkItem

func (pq dijkPQ) Len() int            { return len(pq) }
func (pq dijkPQ) Less(i, j int) bool   { return pq[i].dist < pq[j].dist }
func (pq dijkPQ) Swap(i, j int)        { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq *dijkPQ) Push(x interface{})   { item := x.(*dijkItem); item.index = len(*pq); *pq = append(*pq, item) }
func (pq *dijkPQ) Pop() interface{} {
	old := *pq; n := len(old); item := old[n-1]; old[n-1] = nil; item.index = -1; *pq = old[:n-1]; return item
}

const eps = 1e-9

type predEntry struct {
	from string
	edge *graph.Edge
}

func allShortestDijkstra(src, dst string, adj map[string][]graph.Edge) [][]string {
	if src == dst {
		return nil
	}

	dist := map[string]float64{src: 0}
	preds := map[string][]predEntry{}

	pq := &dijkPQ{{biome: src, dist: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*dijkItem)
		cur := item.biome

		if item.dist > dist[cur]+eps {
			continue // stale
		}

		// If we've popped dst and its distance is finalized, we can stop
		// expanding nodes with dist > dist[dst]
		if d, ok := dist[dst]; ok && item.dist > d+eps {
			break
		}

		for i := range adj[cur] {
			e := &adj[cur][i]
			newDist := dist[cur] + e.Weight

			if d, ok := dist[e.To]; !ok {
				// First time
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &dijkItem{biome: e.To, dist: newDist})
			} else if newDist < d-eps {
				// Strictly better
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &dijkItem{biome: e.To, dist: newDist})
			} else if math.Abs(newDist-d) < eps {
				// Tied
				preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil
	}

	// Enumerate all shortest paths from dst back to src
	var results [][]string
	var backtrack func(node string, path []string)
	backtrack = func(node string, path []string) {
		path = append([]string{node}, path...)
		if node == src {
			cp := make([]string, len(path))
			copy(cp, path)
			results = append(results, cp)
			return
		}
		for _, p := range preds[node] {
			backtrack(p.from, path)
		}
	}
	backtrack(dst, nil)
	return results
}

func allShortestCycleDijkstra(src string, adj map[string][]graph.Edge) [][]string {
	// Check self-loops
	for i := range adj[src] {
		if adj[src][i].To == src {
			return [][]string{{src, src}}
		}
	}

	dist := map[string]float64{}
	preds := map[string][]predEntry{}

	pq := &dijkPQ{}
	heap.Init(pq)

	// Seed with src's neighbors
	for i := range adj[src] {
		e := &adj[src][i]
		if e.To == src {
			continue
		}
		if d, ok := dist[e.To]; !ok {
			dist[e.To] = e.Weight
			preds[e.To] = []predEntry{{from: src, edge: e}}
			heap.Push(pq, &dijkItem{biome: e.To, dist: e.Weight})
		} else if e.Weight < d-eps {
			dist[e.To] = e.Weight
			preds[e.To] = []predEntry{{from: src, edge: e}}
			heap.Push(pq, &dijkItem{biome: e.To, dist: e.Weight})
		} else if math.Abs(e.Weight-d) < eps {
			preds[e.To] = append(preds[e.To], predEntry{from: src, edge: e})
		}
	}

	// Track how we reach src back (cycle closing edges)
	bestCycleDist := math.Inf(1)
	var cyclePreds []predEntry

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*dijkItem)
		cur := item.biome

		if item.dist > dist[cur]+eps {
			continue
		}

		// If current dist exceeds best cycle dist, we're done
		if item.dist > bestCycleDist+eps {
			break
		}

		for i := range adj[cur] {
			e := &adj[cur][i]
			newDist := dist[cur] + e.Weight

			if e.To == src {
				if newDist < bestCycleDist-eps {
					bestCycleDist = newDist
					cyclePreds = []predEntry{{from: cur, edge: e}}
				} else if math.Abs(newDist-bestCycleDist) < eps {
					cyclePreds = append(cyclePreds, predEntry{from: cur, edge: e})
				}
				continue
			}

			if d, ok := dist[e.To]; !ok {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &dijkItem{biome: e.To, dist: newDist})
			} else if newDist < d-eps {
				dist[e.To] = newDist
				preds[e.To] = []predEntry{{from: cur, edge: e}}
				heap.Push(pq, &dijkItem{biome: e.To, dist: newDist})
			} else if math.Abs(newDist-d) < eps {
				preds[e.To] = append(preds[e.To], predEntry{from: cur, edge: e})
			}
		}
	}

	if math.IsInf(bestCycleDist, 1) {
		return nil
	}

	// Enumerate all shortest cycle paths
	var results [][]string
	var backtrack func(node string, path []string, visited map[string]bool)
	backtrack = func(node string, path []string, visited map[string]bool) {
		if visited[node] {
			return // avoid infinite loops in predecessor graph
		}
		path = append([]string{node}, path...)
		if _, hasPred := preds[node]; !hasPred {
			// node must be a direct neighbor of src (seeded node)
			// check that src is its predecessor
			return
		}
		visited[node] = true
		for _, p := range preds[node] {
			if p.from == src {
				full := append([]string{src}, path...)
				cp := make([]string, len(full))
				copy(cp, full)
				results = append(results, cp)
			} else {
				backtrack(p.from, path, visited)
			}
		}
		delete(visited, node)
	}

	for _, cp := range cyclePreds {
		path := []string{src} // cycle end
		visited := map[string]bool{}
		backtrack(cp.from, path, visited)
	}

	return dedup(results)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func pathKey(p []string) string {
	return strings.Join(p, " -> ")
}

func dedup(paths [][]string) [][]string {
	seen := map[string]bool{}
	var result [][]string
	for _, p := range paths {
		k := pathKey(p)
		if !seen[k] {
			seen[k] = true
			result = append(result, p)
		}
	}
	return result
}

func formatPath(p []string) string {
	return strings.Join(p, " -> ")
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	g := graph.New()

	fmt.Println("=============================================================")
	fmt.Println("  ANALYSIS: Biome Pairs with Multiple Shortest Paths (Ties)")
	fmt.Println("=============================================================")
	fmt.Println()

	// -----------------------------------------------------------------------
	// 1. Guaranteed shortest cycles with ties
	// -----------------------------------------------------------------------
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SECTION 1: Guaranteed Shortest Cycles with Multiple Paths")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println()

	guarCycleTies := 0
	for _, biome := range g.Biomes {
		paths := allShortestCycleBFS(biome, g.AdjGuar)
		if len(paths) >= 2 {
			guarCycleTies++
			fmt.Printf("[GUARANTEED CYCLE TIE] %s  (%d paths, %d hops each)\n",
				biome, len(paths), len(paths[0])-1)
			for i, p := range paths {
				fmt.Printf("  Path %d: %s\n", i+1, formatPath(p))
			}
			fmt.Println()
		}
	}
	if guarCycleTies == 0 {
		fmt.Println("  No guaranteed cycle ties found.")
		fmt.Println()
	}

	// -----------------------------------------------------------------------
	// 2. Weighted shortest cycles with ties
	// -----------------------------------------------------------------------
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SECTION 2: Weighted Shortest Cycles with Multiple Paths")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println()

	weightedCycleTies := 0
	for _, biome := range g.Biomes {
		paths := allShortestCycleDijkstra(biome, g.Adj)
		if len(paths) >= 2 {
			weightedCycleTies++
			fmt.Printf("[WEIGHTED CYCLE TIE] %s  (%d paths)\n", biome, len(paths))
			// Compute the weight for verification
			for i, p := range paths {
				w := pathWeight(p, g.Adj)
				fmt.Printf("  Path %d (weight=%.4f): %s\n", i+1, w, formatPath(p))
			}
			fmt.Println()
		}
	}
	if weightedCycleTies == 0 {
		fmt.Println("  No weighted cycle ties found.")
		fmt.Println()
	}

	// -----------------------------------------------------------------------
	// 3. All src->dst pairs: guaranteed shortest path ties
	// -----------------------------------------------------------------------
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SECTION 3: Guaranteed Shortest Path Ties (src -> dst)")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println()

	guarPathTies := 0
	for _, src := range g.Biomes {
		for _, dst := range g.Biomes {
			if src == dst {
				continue
			}
			paths := allShortestBFS(src, dst, g.AdjGuar)
			if len(paths) >= 2 {
				guarPathTies++
				fmt.Printf("[GUARANTEED PATH TIE] %s -> %s  (%d paths, %d hops each)\n",
					src, dst, len(paths), len(paths[0])-1)
				for i, p := range paths {
					fmt.Printf("  Path %d: %s\n", i+1, formatPath(p))
				}
				fmt.Println()
			}
		}
	}
	if guarPathTies == 0 {
		fmt.Println("  No guaranteed path ties found.")
		fmt.Println()
	}

	// -----------------------------------------------------------------------
	// 4. All src->dst pairs: weighted shortest path ties
	// -----------------------------------------------------------------------
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SECTION 4: Weighted Shortest Path Ties (src -> dst)")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println()

	weightedPathTies := 0
	type tieEntry struct {
		src, dst string
		paths    [][]string
	}
	var weightedTies []tieEntry

	for _, src := range g.Biomes {
		for _, dst := range g.Biomes {
			if src == dst {
				continue
			}
			paths := allShortestDijkstra(src, dst, g.Adj)
			if len(paths) >= 2 {
				weightedPathTies++
				weightedTies = append(weightedTies, tieEntry{src: src, dst: dst, paths: paths})
			}
		}
	}

	// Sort by number of tied paths (descending) for readability
	sort.Slice(weightedTies, func(i, j int) bool {
		return len(weightedTies[i].paths) > len(weightedTies[j].paths)
	})

	for _, te := range weightedTies {
		w := pathWeight(te.paths[0], g.Adj)
		fmt.Printf("[WEIGHTED PATH TIE] %s -> %s  (%d paths, weight=%.4f)\n",
			te.src, te.dst, len(te.paths), w)
		for i, p := range te.paths {
			fmt.Printf("  Path %d: %s\n", i+1, formatPath(p))
		}
		fmt.Println()
	}

	if weightedPathTies == 0 {
		fmt.Println("  No weighted path ties found.")
		fmt.Println()
	}

	// -----------------------------------------------------------------------
	// Summary
	// -----------------------------------------------------------------------
	fmt.Println("=============================================================")
	fmt.Println("  SUMMARY")
	fmt.Println("=============================================================")
	fmt.Printf("  Guaranteed cycle ties:  %d biomes\n", guarCycleTies)
	fmt.Printf("  Weighted cycle ties:    %d biomes\n", weightedCycleTies)
	fmt.Printf("  Guaranteed path ties:   %d src->dst pairs\n", guarPathTies)
	fmt.Printf("  Weighted path ties:     %d src->dst pairs\n", weightedPathTies)
	fmt.Println("=============================================================")
}

// pathWeight computes the sum of edge weights along a path specified as biome names.
func pathWeight(path []string, adj map[string][]graph.Edge) float64 {
	w := 0.0
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		for _, e := range adj[from] {
			if e.To == to {
				w += e.Weight
				break
			}
		}
	}
	return w
}
