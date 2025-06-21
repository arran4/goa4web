package config

import "reflect"

// Merge copies non-zero exported fields from src into dst using reflection.
// dst must be a pointer to a struct and src must be a struct or pointer to the
// same type. Fields are merged by field index.
func Merge(dst, src interface{}) {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.Elem().Kind() != reflect.Struct {
		panic("dst must be pointer to struct")
	}
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	if sv.Kind() != reflect.Struct {
		panic("src must be struct")
	}
	dt := dv.Elem()
	if dt.Type() != sv.Type() {
		panic("src and dst must have the same type")
	}
	for i := 0; i < sv.NumField(); i++ {
		sf := sv.Field(i)
		if sf.IsZero() {
			continue
		}
		df := dt.Field(i)
		if !df.CanSet() {
			continue
		}
		df.Set(sf)
	}
}
