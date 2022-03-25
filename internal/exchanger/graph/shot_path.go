package graph

import (
	"container/heap"
	"math"
)

func Dijkstra(g Graph, source, target ID) ([]ID, map[ID]float64, error) {
	// let Q be a priority queue
	minHeap := &nodeDistanceHeap{}

	// distance[source] = 0
	distance := make(map[ID]float64)
	distance[source] = 0.0

	// for each vertex v in G:
	for id := range g.GetNodes() {
		// if v ≠ source:
		if id != source {
			// distance[v] = ∞
			distance[id] = math.MaxFloat64

			// prev[v] = undefined
			// prev[v] = ""
		}

		// Q.add_with_priority(v, distance[v])
		nds := nodeDistance{}
		nds.id = id
		nds.distance = distance[id]

		heap.Push(minHeap, nds)
	}

	heap.Init(minHeap)
	prev := make(map[ID]ID)

	// while Q is not empty:
	for minHeap.Len() != 0 {
		// u = Q.extract_min()
		u := heap.Pop(minHeap).(nodeDistance)

		// if u == target:
		if u.id == target {
			break
		}

		// for each child vertex v of u:
		cmap, err := g.GetTargets(u.id)
		if err != nil {
			return nil, nil, err
		}
		for v := range cmap {
			// alt = distance[u] + weight(u, v)
			weight, err := g.GetWeight(u.id, v)
			if err != nil {
				return nil, nil, err
			}
			alt := distance[u.id] + weight

			// if distance[v] > alt:
			if distance[v] > alt {

				// distance[v] = alt
				distance[v] = alt

				// prev[v] = u
				prev[v] = u.id

				// Q.decrease_priority(v, alt)
				minHeap.updateDistance(v, alt)
			}
		}
		heap.Init(minHeap)
	}

	// path = []
	path := []ID{}

	// u = target
	u := target

	// while prev[u] is defined:
	for {
		if _, ok := prev[u]; !ok {
			break
		}
		// path.push_front(u)
		temp := make([]ID, len(path)+1)
		temp[0] = u
		copy(temp[1:], path)
		path = temp

		// u = prev[u]
		u = prev[u]
	}

	// add the source
	temp := make([]ID, len(path)+1)
	temp[0] = source
	copy(temp[1:], path)
	path = temp

	return path, distance, nil
}

func Prim(g Graph, src ID) (map[Edge]struct{}, error) {

	// let Q be a priority queue
	minHeap := &nodeDistanceHeap{}

	// distance[source] = 0
	distance := make(map[ID]float64)
	distance[src] = 0.0

	// for each vertex v in G:
	for id := range g.GetNodes() {

		// if v ≠ src:
		if id != src {
			// distance[v] = ∞
			distance[id] = math.MaxFloat64

			// prev[v] = undefined
			// prev[v] = ""
		}

		// Q.add_with_priority(v, distance[v])
		nds := nodeDistance{}
		nds.id = id
		nds.distance = distance[id]

		heap.Push(minHeap, nds)
	}

	heap.Init(minHeap)
	prev := make(map[ID]ID)

	// while Q is not empty:
	for minHeap.Len() != 0 {

		// u = Q.extract_min()
		u := heap.Pop(minHeap).(nodeDistance)
		uID := u.id

		// for each adjacent vertex v of u:
		tm, err := g.GetTargets(uID)
		if err != nil {
			return nil, err
		}
		for vID := range tm {

			isExist := false
			for _, one := range *minHeap {
				if vID == one.id {
					isExist = true
					break
				}
			}

			// weight(u, v)
			weight, err := g.GetWeight(uID, vID)
			if err != nil {
				return nil, err
			}

			// if v ∈ Q and distance[v] > weight(u, v):
			if isExist && distance[vID] > weight {

				// distance[v] = weight(u, v)
				distance[vID] = weight

				// prev[v] = u
				prev[vID] = uID

				// Q.decrease_priority(v, weight(u, v))
				minHeap.updateDistance(vID, weight)
				heap.Init(minHeap)
			}
		}

		sm, err := g.GetSources(uID)
		if err != nil {
			return nil, err
		}
		vID := uID
		for uID := range sm {

			isExist := false
			for _, one := range *minHeap {
				if vID == one.id {
					isExist = true
					break
				}
			}

			// weight(u, v)
			weight, err := g.GetWeight(uID, vID)
			if err != nil {
				return nil, err
			}

			// if v ∈ Q and distance[v] > weight(u, v):
			if isExist && distance[vID] > weight {

				// distance[v] = weight(u, v)
				distance[vID] = weight

				// prev[v] = u
				prev[vID] = uID

				// Q.decrease_priority(v, weight(u, v))
				minHeap.updateDistance(vID, weight)
				heap.Init(minHeap)
			}
		}
	}

	tree := make(map[Edge]struct{})
	for k, v := range prev {
		weight, err := g.GetWeight(v, k)
		if err != nil {
			return nil, err
		}
		src, err := g.GetNode(v)
		if err != nil {
			return nil, err
		}
		tgt, err := g.GetNode(k)
		if err != nil {
			return nil, err
		}
		tree[NewEdge(src, tgt, weight)] = struct{}{}
	}
	return tree, nil
}

type nodeDistance struct {
	id       ID
	distance float64
}

type nodeDistanceHeap []nodeDistance

func (h nodeDistanceHeap) Len() int           { return len(h) }
func (h nodeDistanceHeap) Less(i, j int) bool { return h[i].distance < h[j].distance } // Min-Heap
func (h nodeDistanceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nodeDistanceHeap) Push(x interface{}) {
	*h = append(*h, x.(nodeDistance))
}

func (h *nodeDistanceHeap) Pop() interface{} {
	heapSize := len(*h)
	lastNode := (*h)[heapSize-1]
	*h = (*h)[0 : heapSize-1]
	return lastNode
}

func (h *nodeDistanceHeap) updateDistance(id ID, val float64) {
	for i := 0; i < len(*h); i++ {
		if (*h)[i].id == id {
			(*h)[i].distance = val
			break
		}
	}
}
