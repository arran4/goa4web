package upload

import (
	"context"
	"sort"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	id string
}

func (m *mockProvider) Check(ctx context.Context) error { return nil }
func (m *mockProvider) Write(ctx context.Context, name string, data []byte) error { return nil }
func (m *mockProvider) Read(ctx context.Context, name string) ([]byte, error) { return nil, nil }

func TestRegisterProvider(t *testing.T) {
	testName := "testprovider"

	// Ensure cleanup
	defer func() {
		regMu.Lock()
		delete(registry, testName)
		regMu.Unlock()
	}()

	// 1. Check initial state
	initialNames := ProviderNames()
	assert.NotContains(t, initialNames, testName)

	// 2. Register mock provider
	mockP := &mockProvider{id: "1"}
	factory := func(cfg *config.RuntimeConfig) Provider {
		return mockP
	}

	RegisterProvider(testName, factory)

	// 3. Verify registration
	names := ProviderNames()
	assert.Contains(t, names, testName)
	assert.Equal(t, len(initialNames)+1, len(names))
	assert.True(t, sort.StringsAreSorted(names), "Provider names should be sorted")

	// 4. Verify retrieval via config
	cfg := &config.RuntimeConfig{
		ImageUploadProvider: testName,
	}

	p := ProviderFromConfig(cfg)
	assert.Equal(t, mockP, p)

	// 5. Verify overwrite
	mockP2 := &mockProvider{id: "2"}
	factory2 := func(cfg *config.RuntimeConfig) Provider {
		return mockP2
	}

	// This should log a message about overwriting, but we can't easily assert on logs here without capturing stderr.
	// We mainly care that the value is updated.
	RegisterProvider(testName, factory2)

	p2 := ProviderFromConfig(cfg)
	assert.Equal(t, mockP2, p2)
	assert.NotEqual(t, p, p2)

    // 6. Test case insensitivity (RegisterProvider lowercases the name)
    // The key in registry is lowercased.
    // ProviderNames returns the key.
    // So if we registered "TestProvider", it should be stored as "testprovider".

    testNameCase := "TestCaseProvider"
    defer func() {
        regMu.Lock()
        delete(registry, "testcaseprovider")
        regMu.Unlock()
    }()

    RegisterProvider(testNameCase, factory)
    namesCase := ProviderNames()
    assert.Contains(t, namesCase, "testcaseprovider")

    cfgCase := &config.RuntimeConfig{
        ImageUploadProvider: "TestCaseProvider",
    }
    // ProviderFromConfig lowercases the config string too.
    pCase := ProviderFromConfig(cfgCase)
    assert.Equal(t, mockP, pCase)
}
