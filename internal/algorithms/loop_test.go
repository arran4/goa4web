package algorithms

import "testing"

func TestWouldCreateLoopSelf(t *testing.T) {
	parents := map[int32]int32{1: 0, 2: 1}
	path, loop := WouldCreateLoop(parents, 1, 1)
	if !loop {
		t.Fatalf("expected loop")
	}
	if len(path) != 1 || path[0] != 1 {
		t.Fatalf("unexpected path %v", path)
	}
}

func TestWouldCreateLoopChain(t *testing.T) {
	parents := map[int32]int32{1: 0, 2: 1, 3: 2}
	path, loop := WouldCreateLoop(parents, 1, 3)
	if !loop {
		t.Fatalf("expected loop")
	}
	expect := []int32{3, 2, 1}
	if len(path) != len(expect) {
		t.Fatalf("unexpected path length %v", path)
	}
	for i, v := range expect {
		if path[i] != v {
			t.Fatalf("unexpected path %v", path)
		}
	}
}

func TestWouldCreateLoopExisting(t *testing.T) {
	parents := map[int32]int32{1: 2, 2: 1}
	path, loop := WouldCreateLoop(parents, 3, 1)
	if !loop {
		t.Fatalf("expected loop")
	}
	expect := []int32{1, 2, 1}
	if len(path) != len(expect) {
		t.Fatalf("unexpected path %v", path)
	}
	for i, v := range expect {
		if path[i] != v {
			t.Fatalf("unexpected path %v", path)
		}
	}
}

func TestWouldCreateLoopNone(t *testing.T) {
	parents := map[int32]int32{1: 0, 2: 1, 3: 1}
	if path, loop := WouldCreateLoop(parents, 3, 1); loop {
		t.Fatalf("unexpected loop %v", path)
	}
}
