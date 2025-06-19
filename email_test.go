package main

import (
	"os"
	"reflect"
	"testing"
)

func TestGetEmailProviderLog(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "log")
	t.Cleanup(func() { os.Unsetenv("EMAIL_PROVIDER") })
	if p := getEmailProvider(); reflect.TypeOf(p) != reflect.TypeOf(logMailProvider{}) {
		t.Errorf("expected logMailProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "unknown")
	t.Cleanup(func() { os.Unsetenv("EMAIL_PROVIDER") })
	if p := getEmailProvider(); p != nil {
		t.Errorf("expected nil for unknown provider, got %#v", p)
	}
}
