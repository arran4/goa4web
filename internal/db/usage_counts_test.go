package db

import (
	"context"
	"testing"
)

type monthlyUsageCounterStub struct {
	counts map[string]map[[2]int32]int64
	calls  []struct {
		table     string
		column    string
		startYear int32
	}
}

func (s *monthlyUsageCounterStub) monthlyCounts(_ context.Context, table, column string, startYear int32) (map[[2]int32]int64, error) {
	s.calls = append(s.calls, struct {
		table     string
		column    string
		startYear int32
	}{table: table, column: column, startYear: startYear})
	return s.counts[table], nil
}

// TestMonthlyUsageCounts ensures that writing statistics are included in the
// monthly usage aggregation.
func TestMonthlyUsageCounts(t *testing.T) {
	startYear := int32(2024)
	stub := &monthlyUsageCounterStub{
		counts: map[string]map[[2]int32]int64{
			"blogs": {
				{2024, 1}: 2,
			},
			"comments": {
				{2024, 1}: 4,
				{2024, 2}: 1,
			},
			"writing": {
				{2024, 1}: 3,
			},
		},
	}

	rows, err := aggregateMonthlyUsageCounts(context.Background(), stub, startYear)
	if err != nil {
		t.Fatalf("aggregateMonthlyUsageCounts: %v", err)
	}

	if got := len(stub.calls); got != 6 {
		t.Fatalf("expected monthlyCounts to be called for 6 tables, got %d", got)
	}
	for _, c := range stub.calls {
		if c.startYear != startYear {
			t.Fatalf("expected startYear %d, got %d", startYear, c.startYear)
		}
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	find := func(year, month int32) *MonthlyUsageRow {
		for _, r := range rows {
			if r.Year == year && r.Month == month {
				return r
			}
		}
		return nil
	}

	jan := find(2024, 1)
	if jan == nil {
		t.Fatalf("expected row for 2024-01")
	}
	if jan.Blogs != 2 {
		t.Fatalf("expected blogs=2, got %d", jan.Blogs)
	}
	if jan.Comments != 4 {
		t.Fatalf("expected comments=4, got %d", jan.Comments)
	}
	if jan.Writings != 3 {
		t.Fatalf("expected writings=3, got %d", jan.Writings)
	}

	feb := find(2024, 2)
	if feb == nil {
		t.Fatalf("expected row for 2024-02")
	}
	if feb.Comments != 1 {
		t.Fatalf("expected comments=1, got %d", feb.Comments)
	}
	if feb.Writings != 0 {
		t.Fatalf("expected writings=0, got %d", feb.Writings)
	}
}
