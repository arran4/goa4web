package startup

import (
	"context"
	"testing"

	intupload "github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/runtimeconfig"
)

// test provider that records calls
type testProvider struct {
	checkErr error
}

func (t testProvider) Check(ctx context.Context) error                           { return t.checkErr }
func (t testProvider) Write(ctx context.Context, name string, data []byte) error { return nil }
func (t testProvider) Read(ctx context.Context, name string) ([]byte, error)     { return nil, nil }

func TestCheckUploadTargetOK(t *testing.T) {
	intupload.RegisterProvider("testok", func(runtimeconfig.RuntimeConfig) intupload.Provider { return testProvider{} })
	cfg := runtimeconfig.RuntimeConfig{ImageUploadProvider: "testok", ImageUploadDir: "ignored", ImageCacheProvider: "testok", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckUploadTargetFail(t *testing.T) {
	intupload.RegisterProvider("testfail", func(runtimeconfig.RuntimeConfig) intupload.Provider { return testProvider{checkErr: context.Canceled} })
	cfg := runtimeconfig.RuntimeConfig{ImageUploadProvider: "testfail", ImageUploadDir: "ignored", ImageCacheProvider: "testfail", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckUploadTargetNoProvider(t *testing.T) {
	cfg := runtimeconfig.RuntimeConfig{ImageUploadProvider: "missing", ImageUploadDir: "dir", ImageCacheProvider: "missing", ImageCacheDir: "cache"}
	if err := CheckUploadTarget(cfg); err == nil {
		t.Fatalf("expected error")
	}
}
