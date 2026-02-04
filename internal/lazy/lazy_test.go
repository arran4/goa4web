package lazy_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/arran4/goa4web/internal/lazy"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
	return v
}

func TestValueLoadOnce(t *testing.T) {
	var v lazy.Value[int]
	calls := 0
	got, err := v.Load(func() (int, error) {
		calls++
		return 42, nil
	})
	if err != nil || got != 42 {
		t.Fatalf("first load got %v %v", got, err)
	}
	got, err = v.Load(func() (int, error) {
		calls++
		return 99, nil
	})
	if err != nil || got != 42 {
		t.Fatalf("second load got %v %v", got, err)
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestValueLoadError(t *testing.T) {
	var v lazy.Value[int]
	firstErr := errors.New("bad")
	if _, err := v.Load(func() (int, error) { return 0, firstErr }); err != firstErr {
		t.Fatalf("err=%v", err)
	}
	if v, err := v.Load(func() (int, error) { return 1, nil }); err != firstErr || v != 0 {
		t.Fatalf("second load v=%d err=%v", v, err)
	}
}

func TestValueSetPeek(t *testing.T) {
	var v lazy.Value[string]
	v.Set("hello")
	if val, ok := v.Peek(); !ok || val != "hello" {
		t.Fatalf("peek got %v %v", val, ok)
	}
	v.Set("world")
	if val, _ := v.Peek(); val != "hello" {
		t.Fatalf("overwrite val=%s", val)
	}
}

func TestMapNilMap(t *testing.T) {
	var mu sync.Mutex
	_, err := lazy.Map[int](nil, &mu, 1, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMapFetchCaching(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	calls := 0
	fetch := func(id int32) (int, error) {
		calls++
		return int(id * 2), nil
	}
	var mu sync.Mutex
	v := must(lazy.Map(&m, &mu, 1, fetch))
	if v != 2 {
		t.Fatalf("got %v", v)
	}
	v = must(lazy.Map(&m, &mu, 1, fetch))
	if v != 2 || calls != 1 {
		t.Fatalf("cached %v calls=%d", v, calls)
	}
}

func TestMapDontFetchMustCached(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	_, err := lazy.Map(&m, &mu, 1, nil, lazy.DontFetch[int](), lazy.MustBeCached[int]())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMapDontFetchDefaultValue(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	v, err := lazy.Map(&m, &mu, 5, nil, lazy.DontFetch[int](), lazy.DefaultValue[int](42))
	if err != nil || v != 42 {
		t.Fatalf("got %v %v", v, err)
	}
	if got, ok := m[5].Peek(); !ok || got != 42 {
		t.Fatalf("cached %v %v", got, ok)
	}
}

func TestMapMustWrapError(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	_, err := lazy.Map(&m, &mu, 1, func(int32) (int, error) { return 0, errors.New("bad") }, lazy.Must[int]())
	if err == nil || err.Error() != "fetch error: bad" {
		t.Fatalf("err=%v", err)
	}
}

func TestMapDefaultValueOnError(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	v, err := lazy.Map(&m, &mu, 1, func(int32) (int, error) { return 0, errors.New("bad") }, lazy.DefaultValue[int](5))
	if err != nil || v != 5 {
		t.Fatalf("got %v %v", v, err)
	}
}

func TestMapClear(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	lazy.Map(&m, &mu, 1, func(int32) (int, error) { return 1, nil })
	lazy.Map(&m, &mu, 1, nil, lazy.Clear[int]())
	if _, ok := m[1]; ok {
		t.Fatal("value not cleared")
	}
}

func TestMapRefresh(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	calls := 0
	fetch := func(int32) (int, error) { calls++; return calls, nil }
	var mu sync.Mutex
	v := must(lazy.Map(&m, &mu, 1, fetch))
	if v != 1 {
		t.Fatalf("first=%d", v)
	}
	v = must(lazy.Map(&m, &mu, 1, fetch, lazy.Refresh[int]()))
	if v != 2 {
		t.Fatalf("refresh=%d", v)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestMapSet(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	v, err := lazy.Map(&m, &mu, 1, nil, lazy.Set[int](7))
	if err != nil || v != 7 {
		t.Fatalf("set %v %v", v, err)
	}
	v, err = lazy.Map(&m, &mu, 1, nil, lazy.DontFetch[int]())
	if err != nil || v != 7 {
		t.Fatalf("cached %v %v", v, err)
	}
}

func TestMapSetID(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var calls int
	fetch := func(id int32) (int, error) { calls++; return int(id), nil }
	var mu sync.Mutex
	v, err := lazy.Map(&m, &mu, 1, fetch, lazy.SetID[int](2))
	if err != nil || v != 2 {
		t.Fatalf("got %v %v", v, err)
	}
	if _, ok := m[2]; !ok {
		t.Fatal("missing id 2")
	}
	if _, ok := m[1]; ok {
		t.Fatal("unexpected id 1")
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestMapConcurrent(t *testing.T) {
	m := make(map[int32]*lazy.Value[int])
	var mu sync.Mutex
	calls := 0
	fetch := func(id int32) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		return int(id), nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := lazy.Map(&m, &mu, 1, fetch); err != nil || v != 1 {
				t.Errorf("%v %v", v, err)
			}
		}()
	}
	wg.Wait()
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}
