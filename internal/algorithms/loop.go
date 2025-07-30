package algorithms

// WouldCreateLoop determines if assigning newParent as the parent of id would
// result in a cycle. parents maps id -> parentID for existing items. parentID 0
// denotes a root element. The returned slice contains the ids forming the loop
// in the order they are encountered.
func WouldCreateLoop(parents map[int32]int32, id, newParent int32) ([]int32, bool) {
	if newParent == 0 {
		return nil, false
	}
	if newParent == id {
		return []int32{id}, true
	}
	seen := map[int32]int{}
	path := []int32{}
	p := newParent
	for p != 0 {
		if p == id {
			return append(path, p), true
		}
		if idx, ok := seen[p]; ok {
			return append(path[idx:], p), true
		}
		seen[p] = len(path)
		path = append(path, p)
		np, ok := parents[p]
		if !ok {
			break
		}
		p = np
	}
	return nil, false
}
