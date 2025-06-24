package goa4web

import (
	"reflect"
	"testing"
)

func TestGetEmailProviderLog(t *testing.T) {
	cfg := RuntimeConfig{EmailProvider: "log"}
	if p := providerFromConfig(cfg); reflect.TypeOf(p) != reflect.TypeOf(logMailProvider{}) {
		t.Errorf("expected logMailProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	cfg := RuntimeConfig{EmailProvider: "unknown"}
	if p := providerFromConfig(cfg); p != nil {
		t.Errorf("expected nil for unknown provider, got %#v", p)
	}
}
