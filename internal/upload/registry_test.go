package upload

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
)

type mockProvider struct {
	id string
}

func (m *mockProvider) Check(ctx context.Context) error {
	return nil
}

func (m *mockProvider) Write(ctx context.Context, name string, data []byte) error {
	return nil
}

func (m *mockProvider) Read(ctx context.Context, name string) ([]byte, error) {
	return nil, nil
}

func TestRegisterProvider(t *testing.T) {
	// Backup registry
	regMu.Lock()
	oldRegistry := make(map[string]ProviderFactory)
	for k, v := range registry {
		oldRegistry[k] = v
	}
	// Clear registry for test
	registry = make(map[string]ProviderFactory)
	regMu.Unlock()

	defer func() {
		regMu.Lock()
		registry = oldRegistry
		regMu.Unlock()
	}()

	// Register "test1"
	t1Factory := func(cfg *config.RuntimeConfig) Provider {
		return &mockProvider{id: "test1"}
	}
	RegisterProvider("test1", t1Factory)

	// Verify "test1" is in ProviderNames
	names := ProviderNames()
	if len(names) != 1 || names[0] != "test1" {
		t.Errorf("expected [test1], got %v", names)
	}

	// Register "TEST2"
	t2Factory := func(cfg *config.RuntimeConfig) Provider {
		return &mockProvider{id: "test2"}
	}
	RegisterProvider("TEST2", t2Factory)

	// Verify "test1", "test2" are in ProviderNames (sorted)
	names = ProviderNames()
	if len(names) != 2 || names[0] != "test1" || names[1] != "test2" {
		t.Errorf("expected [test1 test2], got %v", names)
	}

	// Verify providerFactory retrieves correct factory
	f1 := providerFactory("test1")
	if f1 == nil {
		t.Error("expected providerFactory to return factory for test1")
	} else {
		p1 := f1(nil)
		if mp, ok := p1.(*mockProvider); !ok || mp.id != "test1" {
			t.Errorf("expected providerFactory to return factory creating test1 provider, got %v", p1)
		}
	}

	f2 := providerFactory("test2")
	if f2 == nil {
		t.Error("expected providerFactory to return factory for test2")
	} else {
		p2 := f2(nil)
		if mp, ok := p2.(*mockProvider); !ok || mp.id != "test2" {
			t.Errorf("expected providerFactory to return factory creating test2 provider, got %v", p2)
		}
	}

	// Test overwriting "test1"
	t1NewFactory := func(cfg *config.RuntimeConfig) Provider {
		return &mockProvider{id: "test1_new"}
	}
	RegisterProvider("test1", t1NewFactory)

	f1New := providerFactory("test1")
	if f1New == nil {
		t.Error("expected providerFactory to return factory for test1")
	} else {
		p1New := f1New(nil)
		if mp, ok := p1New.(*mockProvider); !ok || mp.id != "test1_new" {
			t.Errorf("expected providerFactory to return new factory, got %v", p1New)
		}
	}
}
