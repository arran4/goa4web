# config

## Purpose

Package `config` defines the data structures and parsing logic for the Goa4Web application configuration. It handles reading from environment variables, command-line flags, and configuration files.

## Structure and Components

The primary configuration definitions are found within `runtimeconfig.go` where `RuntimeConfig` is defined. Ensure any new configuration parameters are appended to this struct.

### Exported Types and Interfaces

- **`StringOption`**:
- **`IntOption`**:
- **`BoolOption`**:
- **`ThumbnailSize`**:
- **`RuntimeConfig`**:
  - Methods: `ThumbnailSizes`, `SafeImageDimensions`
- **`Option`**:

### Exported Functions

- `TestLoadOrCreateAdminAPISecretCLI`
- `TestLoadOrCreateAdminAPISecretEnv`
- `TestLoadOrCreateAdminAPISecretFile`
- `TestLoadOrCreateAdminAPISecretGenerate`
- `TestDefaultAdminAPISecretPathDev`
- `TestDefaultAdminAPISecretPathDocker`
- `TestDefaultAdminAPISecretPathUser`
- `DefaultImageSignSecretPath`
- `LoadOrCreateImageSignSecret`
- `DefaultMap`
- `UsageMap`
- `UsageMapWithOptions`
- `NameMap`
- `NameMapWithOptions`
- `ExtendedUsageMap`
- `ExamplesMap`
- `ValuesMap`
- `WithFlagSet`
- `WithFileValues`
- `WithGetenv`
- `WithStringOptions`
- `WithIntOptions`
- `WithBoolOptions`
- `WithRuntimeConfig`
- `NewRuntimeFlagSet`
- `NewRuntimeConfig`
- `UpdatePaginationConfig`
- `TestDBConfigConflicts`
- `TestDBConfigConflictsPass`
- `TestDBConfigConflictsHost`
- `TestDBConfigConflictsPort`
- `TestDBConfigReconstruction`
- `DefaultDataDir`
- `DefaultCacheDir`
- `DefaultAdminAPISecretPath`
- `LoadOrCreateAdminAPISecret`
- `GetAdminEmails`
- `ParseEnvBytes`
- `ToEnvMap`
- `DefaultLinkSignSecretPath`
- `LoadOrCreateLinkSignSecret`
- `TestResolveBoolPrecedence`
- `TestParseBool`
- `DefaultShareSignSecretPath`
- `LoadOrCreateShareSignSecret`
- `ExtendedUsage`
- `TestMerge`
- `TestMergeZeroFields`
- `TestMergePanics`
- `DefaultDataDir`
- `DefaultCacheDir`
- `ApplySMTPFallbacks`
- `LoadAppConfigFile`
- `UpdateConfigKey`
- `AddMissingJSONOptions`
- `Merge`
- `TestLoadOrCreateSecretCLI`
- `TestLoadOrCreateSecretEnv`
- `TestLoadOrCreateSecretFile`
- `TestLoadOrCreateSecretGenerate`
- `TestDefaultSessionSecretPathDev`
- `TestDefaultSessionSecretPathDocker`
- `TestDefaultSessionSecretPathUser`
- `DefaultSessionSecretPath`
- `LoadOrCreateSessionSecret`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/config"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
