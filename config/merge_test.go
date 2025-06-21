package config

import "testing"

type sample struct {
	A string
	B int
	c bool
}

func TestMerge(t *testing.T) {
	dst := sample{A: "x", B: 1, c: true}
	src := sample{B: 2}
	Merge(&dst, src)
	if dst.A != "x" || dst.B != 2 || !dst.c {
		t.Fatalf("merged %#v", dst)
	}
}

func TestMergeZeroFields(t *testing.T) {
	dst := sample{A: "x"}
	src := sample{}
	Merge(&dst, src)
	if dst.A != "x" || dst.B != 0 {
		t.Fatalf("merged %#v", dst)
	}
}

func TestMergePanics(t *testing.T) {
	assertPanic := func(fn func()) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		fn()
	}

	assertPanic(func() { Merge(sample{}, sample{}) })
	assertPanic(func() { Merge(&sample{}, 123) })
	assertPanic(func() { Merge(&sample{}, struct{ A string }{}) })
}
