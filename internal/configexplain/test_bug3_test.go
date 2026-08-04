package configexplain

import (
	"testing"
	"github.com/arran4/goa4web/config"
)

func TestExplainBugStringNormalization(t *testing.T) {
	inputs := Inputs{
		FlagSet: nil,
		StringOptions: []config.StringOption{
			{Name: "str-opt", Env: "STR_OPT", Default: "default"},
		},
		FileValues: map[string]string{
			"STR_OPT": "T",
		},
		Values: map[string]string{
			"STR_OPT": "true",
		},
		Getenv: func(s string) string { return "" },
	}

	infos := Explain(inputs)
	assertOption(t, infos, "str-opt", "true", SourceFlag)
}

func TestExplainBugBoolNormalization(t *testing.T) {
	inputs := Inputs{
		FlagSet: nil,
		BoolOptions: []config.BoolOption{
			{Name: "bool-opt", Env: "BOOL_OPT", Default: false},
		},
		FileValues: map[string]string{
			"BOOL_OPT": "true",
		},
		Values: map[string]string{
			"BOOL_OPT": "1",
		},
		Getenv: func(s string) string { return "" },
	}

	infos := Explain(inputs)
	assertOption(t, infos, "bool-opt", "true", SourceFile)
}
