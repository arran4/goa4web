package stats

import (
	"log"
	"sort"
	"sync"
)

var (
	mu     sync.Mutex
	counts = make(map[string]int64)
)

func Inc(name string) {
	mu.Lock()
	defer mu.Unlock()
	counts[name]++
}

func Add(name string, delta int64) {
	mu.Lock()
	defer mu.Unlock()
	counts[name] += delta
}

func Dump() {
	mu.Lock()
	defer mu.Unlock()
	if len(counts) == 0 {
		return
	}
	log.Printf("Stats dump:")
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := counts[k]
		log.Printf("  %s: %d", k, v)
		delete(counts, k)
	}
}
