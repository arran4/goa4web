package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestConvertToLiteExpr(t *testing.T) {
	dbStructs := map[string]StructInfo{
		"UserParam": {
			Name: "UserParam",
			Fields: []FieldInfo{
				{Name: "ID", Type: "int32"},
				{Name: "Name", Type: "string"},
			},
		},
	}
	liteStructs := map[string]StructInfo{
		"UserParam": {
			Name: "UserParam",
			Fields: []FieldInfo{
				{Name: "ID", Type: "int64"},
				{Name: "Name", Type: "string"},
			},
		},
	}
	dbAliases := map[string]string{"UserID": "int32"}
	liteAliases := map[string]string{"UserID": "int64"}

	tests := []struct {
		name      string
		valName   string
		dbType    string
		liteType  string
		wantExpr  string
		wantError bool
	}{
		{name: "identical primitive", valName: "id", dbType: "int64", liteType: "int64", wantExpr: "id"},
		{name: "identical slice", valName: "names", dbType: "[]string", liteType: "[]string", wantExpr: "names"},
		{name: "int32 to int64", valName: "count", dbType: "int32", liteType: "int64", wantExpr: "int64(count)"},
		{name: "int64 to int32", valName: "count", dbType: "int64", liteType: "int32", wantExpr: "int32(count)"},
		{name: "sql.NullInt32 to sql.NullInt64", valName: "n", dbType: "sql.NullInt32", liteType: "sql.NullInt64", wantExpr: "sql.NullInt64{Int64: int64(n.Int32), Valid: n.Valid}"},
		{name: "sql.NullInt64 to sql.NullInt32", valName: "n", dbType: "sql.NullInt64", liteType: "sql.NullInt32", wantExpr: "sql.NullInt32{Int32: int32(n.Int64), Valid: n.Valid}"},
		{name: "int32 to sql.NullInt64", valName: "v", dbType: "int32", liteType: "sql.NullInt64", wantExpr: "sql.NullInt64{Int64: int64(v), Valid: true}"},
		{name: "bool to int64", valName: "flag", dbType: "bool", liteType: "int64", wantExpr: "func(b bool) int64 { if b { return 1 }; return 0 }(flag)"},
		{name: "string to sql.NullString", valName: "str", dbType: "string", liteType: "sql.NullString", wantExpr: `sql.NullString{String: str, Valid: str != ""}`},
		{name: "any to interface{}", valName: "val", dbType: "sql.NullTime", liteType: "interface{}", wantExpr: "val"},
		{name: "[]int32 to []int64", valName: "ids", dbType: "[]int32", liteType: "[]int64", wantExpr: "func(s []int32) []int64 { if s == nil { return nil }; out := make([]int64, len(s)); for i, v := range s { out[i] = int64(v) }; return out }(ids)"},
		{name: "type alias", valName: "u", dbType: "UserID", liteType: "UserID", wantExpr: "dbsqlite.UserID(u)"},
		{name: "struct mapping", valName: "arg", dbType: "UserParam", liteType: "UserParam", wantExpr: "dbsqlite.UserParam{\n\t\tID: int64(arg.ID),\n\t\tName: arg.Name,\n\t}"},
		{name: "unsupported mapping", valName: "unknown", dbType: "complex128", liteType: "chan int", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := convertToLiteExpr(tc.valName, tc.dbType, tc.liteType, dbStructs, liteStructs, dbAliases, liteAliases)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error for %s -> %s, got nil (expr: %s)", tc.dbType, tc.liteType, expr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr != tc.wantExpr {
				t.Errorf("got %q, want %q", expr, tc.wantExpr)
			}
		})
	}
}

func TestConvertFromLiteExpr(t *testing.T) {
	dbStructs := map[string]StructInfo{
		"UserRow": {
			Name: "UserRow",
			Fields: []FieldInfo{
				{Name: "ID", Type: "int32"},
				{Name: "Name", Type: "string"},
			},
		},
	}
	liteStructs := map[string]StructInfo{
		"UserRow": {
			Name: "UserRow",
			Fields: []FieldInfo{
				{Name: "ID", Type: "int64"},
				{Name: "Name", Type: "string"},
			},
		},
	}
	dbAliases := map[string]string{"UserID": "int32"}
	liteAliases := map[string]string{"UserID": "int64"}

	tests := []struct {
		name      string
		valName   string
		dbType    string
		liteType  string
		wantExpr  string
		wantError bool
	}{
		{name: "identical primitive", valName: "res", dbType: "int64", liteType: "int64", wantExpr: "res"},
		{name: "interface{} to int32", valName: "res", dbType: "int32", liteType: "interface{}", wantExpr: "int32(toInt64(res))"},
		{name: "interface{} to int64", valName: "res", dbType: "int64", liteType: "interface{}", wantExpr: "toInt64(res)"},
		{name: "interface{} to string", valName: "res", dbType: "string", liteType: "interface{}", wantExpr: "toString(res)"},
		{name: "interface{} to sql.NullTime", valName: "res", dbType: "sql.NullTime", liteType: "interface{}", wantExpr: "toNullTime(res)"},
		{name: "int64 to int32", valName: "res", dbType: "int32", liteType: "int64", wantExpr: "int32(res)"},
		{name: "int64 to bool", valName: "res", dbType: "bool", liteType: "int64", wantExpr: "(res != 0)"},
		{name: "sql.NullInt64 to sql.NullBool", valName: "res", dbType: "sql.NullBool", liteType: "sql.NullInt64", wantExpr: "sql.NullBool{Bool: res.Int64 != 0, Valid: res.Valid}"},
		{name: "[]byte to sql.NullString", valName: "res", dbType: "sql.NullString", liteType: "[]byte", wantExpr: "sql.NullString{String: string(res), Valid: res != nil}"},
		{name: "[]int64 to []int32", valName: "res", dbType: "[]int32", liteType: "[]int64", wantExpr: "func(s []int64) []int32 { if s == nil { return nil }; out := make([]int32, len(s)); for i, v := range s { out[i] = int32(v) }; return out }(res)"},
		{name: "pointer struct mapping", valName: "res", dbType: "*UserRow", liteType: "*UserRow", wantExpr: "func(v *dbsqlite.UserRow) *UserRow {\n\t\tif v == nil { return nil }\n\t\treturn &UserRow{\n\t\t\tID: int32(v.ID),\n\t\t\tName: v.Name,\n\t\t}\n\t}(res)"},
		{name: "unsupported mapping", valName: "res", dbType: "chan int", liteType: "complex128", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := convertFromLiteExpr(tc.valName, tc.dbType, tc.liteType, dbStructs, liteStructs, dbAliases, liteAliases)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error for %s <- %s, got nil (expr: %s)", tc.dbType, tc.liteType, expr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr != tc.wantExpr {
				t.Errorf("got %q, want %q", expr, tc.wantExpr)
			}
		})
	}
}

func TestGenMethodValidation(t *testing.T) {
	var buf bytes.Buffer
	dbM := MethodInfo{
		Name:       "TestFunc",
		Params:     []ParamInfo{{Name: "ctx", Type: "context.Context"}, {Name: "id", Type: "int32"}},
		ReturnList: []string{"error"},
	}

	// Case 1: Parameter count mismatch
	liteMismatch := MethodInfo{
		Name:       "TestFunc",
		Params:     []ParamInfo{{Name: "ctx", Type: "context.Context"}},
		ReturnList: []string{"error"},
	}
	err := genMethod(&buf, dbM, liteMismatch, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "parameter count mismatch") {
		t.Errorf("expected parameter count mismatch error, got %v", err)
	}

	// Case 2: Successful method generation
	liteMatching := MethodInfo{
		Name:       "TestFunc",
		Params:     []ParamInfo{{Name: "ctx", Type: "context.Context"}, {Name: "id", Type: "int64"}},
		ReturnList: []string{"error"},
	}
	buf.Reset()
	err = genMethod(&buf, dbM, liteMatching, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "func (s *sqliteQuerier) TestFunc(ctx context.Context, id int32) error") {
		t.Errorf("generated signature unexpected: %s", output)
	}
	if !strings.Contains(output, "return s.q.TestFunc(ctx, int64(id))") {
		t.Errorf("generated body unexpected: %s", output)
	}
}
