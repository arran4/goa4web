package privateforum

import (

	"math"
	"strconv"
	"testing"
)

func TestParsePrivateAPIID(t *testing.T) {
	tests := []struct {
		input       string
		expectedID  int32
		expectedErr bool
	}{
		{"1", 1, false},
		{"2147483647", 2147483647, false}, // Max int32
		{"0", 0, true},
		{"-1", 0, true},
		{"2147483648", 0, true},          // Over max int32
		{"4294967297", 0, true},          // Over max int32, but would alias to 1 if cast from int64 without bounds check
		{"9223372036854775807", 0, true}, // Max int64
		{"invalid", 0, true},
		{"", 0, true},
		{"  1  ", 0, true}, // Trailing spaces usually not parsed by strconv.ParseInt without trimming
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := parsePrivateAPIID(tt.input)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error for input %q, but got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if id != tt.expectedID {
					t.Errorf("Expected ID %d for input %q, but got %d", tt.expectedID, tt.input, id)
				}
			}
		})
	}
}

func TestPrivateAPIPageOffset(t *testing.T) {
	tests := []struct {
		name             string
		pageStr          string
		pageSize         int
		collectionLength int
		expectedStart    int
		expectedEnd      int
		expectedHasMore  bool
		expectedErr      bool
	}{
		{"Valid first page", "1", 10, 25, 0, 10, true, false},
		{"Valid second page", "2", 10, 25, 10, 20, true, false},
		{"Valid last page exactly", "3", 10, 30, 20, 30, false, false},
		{"Valid last page partial", "3", 10, 25, 20, 25, false, false},
		{"Default page (empty string)", "", 10, 25, 0, 10, true, false},
		{"Page out of bounds (too high)", "4", 10, 25, 25, 25, false, false},
		{"Page out of bounds (way too high)", "100", 10, 25, 25, 25, false, false},
		{"Invalid page (zero)", "0", 10, 25, 0, 0, false, true},
		{"Invalid page (negative)", "-1", 10, 25, 0, 0, false, true},
		{"Invalid page (non-numeric)", "abc", 10, 25, 0, 0, false, true},
		{"Invalid page size (zero)", "1", 0, 25, 0, 0, false, true},
		{"Invalid page size (negative)", "1", -10, 25, 0, 0, false, true},
		{"Page number overflow", strconv.Itoa(math.MaxInt64) + "0", 10, 25, 0, 0, false, true},
		{"Valid 32-bit max page without panic", "2147483647", 20, 25, 25, 25, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, hasMore, err := privateAPIPageOffset(tt.pageStr, tt.pageSize, tt.collectionLength)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error for %q, but got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %q: %v", tt.name, err)
				}
				if start != tt.expectedStart {
					t.Errorf("Expected start %d for %q, but got %d", tt.expectedStart, tt.name, start)
				}
				if end != tt.expectedEnd {
					t.Errorf("Expected end %d for %q, but got %d", tt.expectedEnd, tt.name, end)
				}
				if hasMore != tt.expectedHasMore {
					t.Errorf("Expected hasMore %v for %q, but got %v", tt.expectedHasMore, tt.name, hasMore)
				}

				// Validate safe slicing
				slice := make([]int, tt.collectionLength)
				if start > len(slice) || start < 0 {
					t.Errorf("Start index %d out of bounds for slice length %d", start, len(slice))
				}
				if end > len(slice) || end < 0 {
					t.Errorf("End index %d out of bounds for slice length %d", end, len(slice))
				}
				if start > end {
					t.Errorf("Start index %d greater than end index %d", start, end)
				}
				_ = slice[start:end] // This would panic if bounds were violated
			}
		})
	}
}

func TestPrivateAPIPageOffsetEndOverflow(t *testing.T) {
	// Specifically test the case where start + pageSize could overflow int32/int64
	// resulting in a negative end if unchecked.
	start, end, hasMore, err := privateAPIPageOffset("2", math.MaxInt-1, math.MaxInt)
	if err != nil {
		t.Fatalf("Unexpected error for end overflow case: %v", err)
	}

	expectedStart := math.MaxInt - 1
	expectedEnd := math.MaxInt

	if start != expectedStart {
		t.Errorf("Expected start %d, got %d", expectedStart, start)
	}
	if end != expectedEnd {
		t.Errorf("Expected end %d, got %d", expectedEnd, end)
	}
	if hasMore != false {
		t.Errorf("Expected hasMore false, got %v", hasMore)
	}
}
