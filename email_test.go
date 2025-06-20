package main

import (
	"reflect"
	"testing"
)

func TestGetEmailProviderLog(t *testing.T) {
	cfg := EmailConfig{Provider: "log"}
	if p := providerFromConfig(cfg); reflect.TypeOf(p) != reflect.TypeOf(logMailProvider{}) {
		t.Errorf("expected logMailProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	cfg := EmailConfig{Provider: "unknown"}
	if p := providerFromConfig(cfg); p != nil {
		t.Errorf("expected nil for unknown provider, got %#v", p)
	}
}
