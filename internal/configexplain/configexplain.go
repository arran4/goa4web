package configexplain

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/arran4/goa4web/config"
)

// SourceKind identifies where a configuration value was resolved.
type SourceKind string

const (
	// SourceFlag indicates a value sourced from command-line flags.
	SourceFlag SourceKind = "flag"
	// SourceFile indicates a value sourced from the config file.
	SourceFile SourceKind = "file"
	// SourceEnv indicates a value sourced from environment variables.
	SourceEnv SourceKind = "env"
	// SourceDefault indicates a value sourced from defaults.
	SourceDefault SourceKind = "default"
)

// Inputs describes the configuration sources to inspect.
type Inputs struct {
	FlagSet       *flag.FlagSet
	FileValues    map[string]string
	Getenv        func(string) string
	ConfigFile    string
	Values        map[string]string
	StringOptions []config.StringOption
	IntOptions    []config.IntOption
	BoolOptions   []config.BoolOption
}

// SourceDetail describes a configuration source and its resolved value.
type SourceDetail struct {
	Kind   SourceKind
	Label  string
	Value  string
	Detail string
	Active bool
}

// OptionInfo captures a resolved configuration option and its sources.
type OptionInfo struct {
	Name         string
	Env          string
	FinalValue   string
	Source       SourceKind
	SourceLabel  string
	SourceDetail string
	Sources      []SourceDetail
}

// Explain returns configuration sources for each option.
func Explain(inputs Inputs) []OptionInfo {
	getenv := inputs.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	fileVals := inputs.FileValues
	if fileVals == nil {
		fileVals = map[string]string{}
	}
	strings := inputs.StringOptions
	if len(strings) == 0 {
		strings = config.StringOptions
	}
	ints := inputs.IntOptions
	if len(ints) == 0 {
		ints = config.IntOptions
	}
	bools := inputs.BoolOptions
	if len(bools) == 0 {
		bools = config.BoolOptions
	}

	setFlags := map[string]bool{}
	if inputs.FlagSet != nil {
		inputs.FlagSet.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	}

	var infos []OptionInfo
	for _, o := range strings {
		flagVal := flagValue(inputs.FlagSet, o.Name)
		envVal := getenv(o.Env)
		fileVal := fileVals[o.Env]
		currentVal := valueOrDefault(inputs.Values, o.Env)
		defaultVal := o.Default
		flagSet := setFlags[o.Name]
		if !flagSet && inputs.FlagSet == nil && shouldInferFlagSource(defaultVal, currentVal, fileVal, envVal) {
			flagSet = true
			flagVal = currentVal
		}
		finalVal, source, sourceLabel, detail := resolveString(defaultVal, flagVal, fileVal, envVal, flagSet, o.Name, o.Env, inputs.ConfigFile)
		infos = append(infos, OptionInfo{
			Name:         o.Name,
			Env:          o.Env,
			FinalValue:   finalVal,
			Source:       source,
			SourceLabel:  sourceLabel,
			SourceDetail: detail,
			Sources:      buildSources(source, o.Name, o.Env, flagVal, fileVal, envVal, defaultVal, flagSet, inputs.ConfigFile),
		})
	}

	for _, o := range ints {
		flagVal := flagValue(inputs.FlagSet, o.Name)
		envVal := getenv(o.Env)
		fileVal := fileVals[o.Env]
		currentVal := valueOrDefault(inputs.Values, o.Env)
		defaultVal := strconv.Itoa(o.Default)
		flagSet := setFlags[o.Name]
		if !flagSet && inputs.FlagSet == nil && shouldInferFlagSource(defaultVal, currentVal, fileVal, envVal) {
			flagSet = true
			flagVal = currentVal
		}
		finalVal, source, sourceLabel, detail := resolveString(defaultVal, flagVal, fileVal, envVal, flagSet, o.Name, o.Env, inputs.ConfigFile)
		infos = append(infos, OptionInfo{
			Name:         o.Name,
			Env:          o.Env,
			FinalValue:   finalVal,
			Source:       source,
			SourceLabel:  sourceLabel,
			SourceDetail: detail,
			Sources:      buildSources(source, o.Name, o.Env, flagVal, fileVal, envVal, defaultVal, flagSet, inputs.ConfigFile),
		})
	}

	for _, o := range bools {
		flagVal := flagValue(inputs.FlagSet, o.Name)
		envVal := getenv(o.Env)
		fileVal := fileVals[o.Env]
		currentVal := valueOrDefault(inputs.Values, o.Env)
		defaultVal := strconv.FormatBool(o.Default)
		flagSet := setFlags[o.Name]
		if !flagSet && inputs.FlagSet == nil && shouldInferFlagSource(defaultVal, currentVal, fileVal, envVal) {
			flagSet = true
			flagVal = currentVal
		}
		finalVal, source, sourceLabel, detail := resolveBool(defaultVal, flagVal, fileVal, envVal, flagSet, o.Name, o.Env, inputs.ConfigFile)
		infos = append(infos, OptionInfo{
			Name:         o.Name,
			Env:          o.Env,
			FinalValue:   finalVal,
			Source:       source,
			SourceLabel:  sourceLabel,
			SourceDetail: detail,
			Sources:      buildSources(source, o.Name, o.Env, flagVal, fileVal, envVal, defaultVal, flagSet, inputs.ConfigFile),
		})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

func flagValue(fs *flag.FlagSet, name string) string {
	if fs == nil {
		return ""
	}
	if f := fs.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

func valueOrDefault(values map[string]string, key string) string {
	if values == nil {
		return ""
	}
	return values[key]
}

func shouldInferFlagSource(defaultVal, currentVal, fileVal, envVal string) bool {
	if currentVal == "" {
		return false
	}
	if fileVal != "" || envVal != "" {
		return false
	}
	return currentVal != defaultVal
}

func resolveString(defaultVal, flagVal, fileVal, envVal string, flagSet bool, name, envKey, cfgFile string) (string, SourceKind, string, string) {
	finalVal := defaultVal
	source := SourceDefault
	sourceLabel := sourceLabelFor(source)
	detail := ""

	if flagSet {
		finalVal = flagVal
		source = SourceFlag
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("--%s", name)
	} else if fileVal != "" {
		finalVal = fileVal
		source = SourceFile
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("%s key: %s", configFileLabel(cfgFile), envKey)
	} else if envVal != "" {
		finalVal = envVal
		source = SourceEnv
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("%s=%s", envKey, envVal)
	}

	return finalVal, source, sourceLabel, detail
}

func resolveBool(defaultVal, flagVal, fileVal, envVal string, flagSet bool, name, envKey, cfgFile string) (string, SourceKind, string, string) {
	finalVal := defaultVal
	source := SourceDefault
	sourceLabel := sourceLabelFor(source)
	detail := ""

	if flagSet && flagVal != "" {
		finalVal = normalizeBool(flagVal)
		source = SourceFlag
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("--%s", name)
	} else if fileVal != "" {
		finalVal = normalizeBool(fileVal)
		source = SourceFile
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("%s key: %s", configFileLabel(cfgFile), envKey)
	} else if envVal != "" {
		finalVal = normalizeBool(envVal)
		source = SourceEnv
		sourceLabel = sourceLabelFor(source)
		detail = fmt.Sprintf("%s=%s", envKey, envVal)
	}

	return finalVal, source, sourceLabel, detail
}

func buildSources(active SourceKind, name, envKey, flagVal, fileVal, envVal, defaultVal string, flagSet bool, cfgFile string) []SourceDetail {
	return []SourceDetail{
		makeSource(SourceFlag, flagVal, flagSet, fmt.Sprintf("--%s", name), active == SourceFlag),
		makeSource(SourceFile, fileVal, fileVal != "", fmt.Sprintf("%s key: %s", configFileLabel(cfgFile), envKey), active == SourceFile),
		makeSource(SourceEnv, envVal, envVal != "", fmt.Sprintf("%s=%s", envKey, envVal), active == SourceEnv),
		makeSource(SourceDefault, defaultVal, true, "Default value", active == SourceDefault),
	}
}

func makeSource(kind SourceKind, value string, hasValue bool, detail string, active bool) SourceDetail {
	if !hasValue {
		detail = "Not set"
		value = ""
	}
	return SourceDetail{
		Kind:   kind,
		Label:  sourceBreakdownLabel(kind),
		Value:  normalizeBool(value),
		Detail: detail,
		Active: active,
	}
}

func normalizeBool(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return value
	}
	return strconv.FormatBool(parsed)
}

func sourceLabelFor(kind SourceKind) string {
	switch kind {
	case SourceFlag:
		return "Arg"
	case SourceFile:
		return "Config File"
	case SourceEnv:
		return "Environment"
	default:
		return "Default"
	}
}

func sourceBreakdownLabel(kind SourceKind) string {
	switch kind {
	case SourceFlag:
		return "Flag"
	case SourceFile:
		return "File"
	case SourceEnv:
		return "Env"
	default:
		return "Default"
	}
}

func configFileLabel(cfgFile string) string {
	if cfgFile == "" {
		return "config file"
	}
	return cfgFile
}
