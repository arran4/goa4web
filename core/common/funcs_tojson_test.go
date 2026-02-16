package common_test

import (
	"html/template"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestToJSON(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want template.JS
	}{
		{"String", "hello", `"hello"`},
		{"Int", 123, `123`},
		{"Struct", struct {
			A int
		}{A: 1}, `{"A":1}`},
		{"Slice", []int{1, 2}, `[1,2]`},
		{"Map", map[string]int{"a": 1}, `{"a":1}`},
		{"Nil", nil, `null`},
		{"Channel (unsupported type)", make(chan int), `null`},
		{"Cyclic structure (unsupported)", func() any {
			type Cycle struct {
				Next *Cycle
			}
			c := &Cycle{}
			c.Next = c
			return c
		}(), `null`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.ToJSON(tt.in); got != tt.want {
				t.Errorf("ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
