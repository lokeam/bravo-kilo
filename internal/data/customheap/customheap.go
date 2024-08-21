package customheap

// Define a heap structure for tag counting
type TagHeap []TagCount

type TagCount struct {
	Tag   string
	Count int
}

func (h TagHeap) Len() int            { return len(h) }
func (h TagHeap) Less(i, j int) bool  { return h[i].Count > h[j].Count } // Max-Heap (Descending)
func (h TagHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *TagHeap) Push(x interface{}) { *h = append(*h, x.(TagCount)) }
func (h *TagHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}