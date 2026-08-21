//go:build sqlite || sqlite3

package db

import (
	"database/sql"
	"testing"
	"time"
)

func TestToInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int64
		wantPanic bool
	}{
		{name: "nil", input: nil, expected: 0},
		{name: "int64", input: int64(123456789), expected: 123456789},
		{name: "int64 negative", input: int64(-42), expected: -42},
		{name: "int32", input: int32(42), expected: 42},
		{name: "int", input: int(100), expected: 100},
		{name: "uint64", input: uint64(500), expected: 500},
		{name: "uint32", input: uint32(250), expected: 250},
		{name: "uint", input: uint(75), expected: 75},
		{name: "float64", input: float64(123.45), expected: 123},
		{name: "numeric string", input: "12345", expected: 12345},
		{name: "negative numeric string", input: "-999", expected: -999},
		{name: "empty string", input: "", expected: 0},
		{name: "numeric []byte", input: []byte("98765"), expected: 98765},
		{name: "negative numeric []byte", input: []byte("-123"), expected: -123},
		{name: "empty []byte", input: []byte(""), expected: 0},
		{name: "invalid string", input: "not_a_number", wantPanic: true},
		{name: "invalid []byte", input: []byte("abc"), wantPanic: true},
		{name: "unexpected type struct", input: struct{}{}, wantPanic: true},
		{name: "unexpected type bool", input: true, wantPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for input %v (%T), but did not panic", tc.input, tc.input)
					}
				}()
			}

			res := toInt64(tc.input)
			if !tc.wantPanic && res != tc.expected {
				t.Errorf("toInt64(%v) = %d, expected %d", tc.input, res, tc.expected)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  string
		wantPanic bool
	}{
		{name: "nil", input: nil, expected: ""},
		{name: "string", input: "hello world", expected: "hello world"},
		{name: "empty string", input: "", expected: ""},
		{name: "[]byte", input: []byte("bytes text"), expected: "bytes text"},
		{name: "empty []byte", input: []byte(""), expected: ""},
		{name: "unexpected type int", input: 123, wantPanic: true},
		{name: "unexpected type struct", input: struct{}{}, wantPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for input %v (%T), but did not panic", tc.input, tc.input)
					}
				}()
			}

			res := toString(tc.input)
			if !tc.wantPanic && res != tc.expected {
				t.Errorf("toString(%v) = %q, expected %q", tc.input, res, tc.expected)
			}
		})
	}
}

func TestToNullTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name      string
		input     any
		expected  sql.NullTime
		wantPanic bool
	}{
		{name: "nil", input: nil, expected: sql.NullTime{Valid: false}},
		{name: "time.Time", input: now, expected: sql.NullTime{Time: now, Valid: true}},
		{name: "*time.Time valid", input: &now, expected: sql.NullTime{Time: now, Valid: true}},
		{name: "*time.Time nil", input: (*time.Time)(nil), expected: sql.NullTime{Valid: false}},
		{name: "RFC3339 string", input: "2026-08-20T12:00:00Z", expected: sql.NullTime{Time: time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC), Valid: true}},
		{name: "Standard datetime string", input: "2026-08-20 12:00:00", expected: sql.NullTime{Time: time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC), Valid: true}},
		{name: "Empty string", input: "", expected: sql.NullTime{Valid: false}},
		{name: "Empty []byte", input: []byte(""), expected: sql.NullTime{Valid: false}},
		{name: "[]byte datetime", input: []byte("2026-08-20 12:00:00"), expected: sql.NullTime{Time: time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC), Valid: true}},
		{name: "invalid string", input: "not_a_time", wantPanic: true},
		{name: "unexpected type int", input: 123, wantPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for input %v (%T), but did not panic", tc.input, tc.input)
					}
				}()
			}

			res := toNullTime(tc.input)
			if !tc.wantPanic {
				if res.Valid != tc.expected.Valid {
					t.Errorf("toNullTime(%v).Valid = %v, expected %v", tc.input, res.Valid, tc.expected.Valid)
				}
				if res.Valid && !res.Time.Equal(tc.expected.Time) {
					t.Errorf("toNullTime(%v).Time = %v, expected %v", tc.input, res.Time, tc.expected.Time)
				}
			}
		})
	}
}

func TestNewForDriver(t *testing.T) {
	// Dummy DBTX (nil is valid for constructing wrapper structs if not executing queries)
	querierMySQL := NewForDriver(nil, "mysql")
	if _, ok := querierMySQL.(*Queries); !ok {
		t.Errorf("expected *Queries for mysql driver, got %T", querierMySQL)
	}

	querierSQLite := NewForDriver(nil, "sqlite")
	if _, ok := querierSQLite.(*sqliteQuerier); !ok {
		t.Errorf("expected *sqliteQuerier for sqlite driver, got %T", querierSQLite)
	}

	querierSQLite3 := NewForDriver(nil, "sqlite3")
	if _, ok := querierSQLite3.(*sqliteQuerier); !ok {
		t.Errorf("expected *sqliteQuerier for sqlite3 driver, got %T", querierSQLite3)
	}
}
