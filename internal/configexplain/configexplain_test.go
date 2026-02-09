package configexplain

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/assert"
)

func TestExplain(t *testing.T) {
	t.Run("Standard Resolution", func(t *testing.T) {
		// Define a custom FlagSet for testing
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("flag-opt", "default", "usage")
		fs.Int("int-opt", 0, "usage")
		fs.Bool("bool-opt", false, "usage")
		// Simulate parsing to mark flags as set
		_ = fs.Parse([]string{"-flag-opt", "flag-value", "-int-opt", "20", "-bool-opt", "true"})

		// Helper to create Inputs
		baseInputs := Inputs{
			FlagSet: fs,
			StringOptions: []config.StringOption{
				{Name: "flag-opt", Env: "FLAG_OPT", Default: "default"},
				{Name: "file-opt", Env: "FILE_OPT", Default: "default"},
				{Name: "env-opt", Env: "ENV_OPT", Default: "default"},
				{Name: "default-opt", Env: "DEFAULT_OPT", Default: "default"},
			},
			IntOptions: []config.IntOption{
				{Name: "int-opt", Env: "INT_OPT", Default: 10},
			},
			BoolOptions: []config.BoolOption{
				{Name: "bool-opt", Env: "BOOL_OPT", Default: false},
			},
			Getenv: func(key string) string {
				if key == "ENV_OPT" {
					return "env-value"
				}
				return ""
			},
			FileValues: map[string]string{
				"FILE_OPT": "file-value",
			},
			Values: map[string]string{
				"DEFAULT_OPT": "default",
			},
		}

		infos := Explain(baseInputs)

		// Verify String Options
		assertOption(t, infos, "flag-opt", "flag-value", SourceFlag)
		assertOption(t, infos, "file-opt", "file-value", SourceFile)
		assertOption(t, infos, "env-opt", "env-value", SourceEnv)
		assertOption(t, infos, "default-opt", "default", SourceDefault)

		// Verify Int Options
		assertOption(t, infos, "int-opt", "20", SourceFlag)

		// Verify Bool Options
		assertOption(t, infos, "bool-opt", "true", SourceFlag)
	})

	t.Run("Inferred Flag (No FlagSet)", func(t *testing.T) {
		inferredInputs := Inputs{
			FlagSet: nil,
			StringOptions: []config.StringOption{
				{Name: "inferred-opt", Env: "INFERRED_OPT", Default: "default"},
			},
			Values: map[string]string{
				"INFERRED_OPT": "inferred-value",
			},
			Getenv: func(s string) string { return "" },
		}

		infosInferred := Explain(inferredInputs)
		assertOption(t, infosInferred, "inferred-opt", "inferred-value", SourceFlag)
	})

	t.Run("Boolean Normalization", func(t *testing.T) {
		boolInputs := Inputs{
			BoolOptions: []config.BoolOption{
				{Name: "bool-norm", Env: "BOOL_NORM", Default: false},
			},
			Getenv: func(key string) string {
				if key == "BOOL_NORM" {
					return "TRUE" // Should normalize to "true"
				}
				return ""
			},
		}
		infosBool := Explain(boolInputs)
		assertOption(t, infosBool, "bool-norm", "true", SourceEnv)
	})

	t.Run("Inferred Flag Overrides File", func(t *testing.T) {
		// Reproducing the bug:
		// File has value, but Current value is different (implies Flag).
		// Should infer SourceFlag, not SourceFile.
		inputs := Inputs{
			FlagSet: nil,
			StringOptions: []config.StringOption{
				{Name: "override-opt", Env: "OVERRIDE_OPT", Default: "default"},
			},
			FileValues: map[string]string{
				"OVERRIDE_OPT": "from-file",
			},
			Values: map[string]string{
				"OVERRIDE_OPT": "from-flag",
			},
			Getenv: func(s string) string { return "" },
		}

		infos := Explain(inputs)
		assertOption(t, infos, "override-opt", "from-flag", SourceFlag)
	})
}

func assertOption(t *testing.T, infos []OptionInfo, name, expectedVal string, expectedSource SourceKind) {
	t.Helper()
	found := false
	for _, info := range infos {
		if info.Name == name {
			assert.Equal(t, expectedVal, info.FinalValue, "Value mismatch for %s", name)
			assert.Equal(t, expectedSource, info.Source, "Source mismatch for %s", name)
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Option %s not found", name)
	}
}
