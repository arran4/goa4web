package app

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	intupload "github.com/arran4/goa4web/internal/upload"
)

// test provider that records calls
type testProvider struct {
	checkErr error
}

func (t testProvider) Check(ctx context.Context) error                           { return t.checkErr }
func (t testProvider) Write(ctx context.Context, name string, data []byte) error { return nil }
func (t testProvider) Read(ctx context.Context, name string) ([]byte, error)     { return nil, nil }

func TestCheckUploadTargetOK(t *testing.T) {
	intupload.RegisterProvider("testok", func(*config.RuntimeConfig) intupload.Provider { return testProvider{} })
	cfg := config.RuntimeConfig{ImageUploadProvider: "testok", ImageUploadDir: "ignored", ImageCacheProvider: "testok", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckUploadTargetFail(t *testing.T) {
	intupload.RegisterProvider("testfail", func(*config.RuntimeConfig) intupload.Provider { return testProvider{checkErr: context.Canceled} })
	cfg := config.RuntimeConfig{ImageUploadProvider: "testfail", ImageUploadDir: "ignored", ImageCacheProvider: "testfail", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckUploadTargetNoProvider(t *testing.T) {
	cfg := config.RuntimeConfig{ImageUploadProvider: "missing", ImageUploadDir: "dir", ImageCacheProvider: "missing", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err == nil {
		t.Fatalf("expected error")
	}
}
