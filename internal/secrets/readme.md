# internal/secrets

## Purpose

Package `secrets` provides internal, non-exported utilities and service integrations specific to `secrets`.

## Why It Exists

To encapsulate the logic necessary for this specific operational domain, ensuring modularity within the codebase.

## What It Allows

It allows the system to remain decoupled. Code outside this package can rely on its exported API without worrying about its internal implementation details.

## Structure and Components

The primary files and their general responsibilities include:

- `secrets.go`

### Exported Functions

- `DefaultPath`
- `LoadOrCreate`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/internal/secrets"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

## Generated secret files

Development builds default to dot-prefixed files in the working directory, such as `.image_sign_secret`. These files must remain untracked; the repository ignores `.*_secret` and CI rejects tracked files matching that pattern, even if force-added. Release builds normally use the user configuration directory or `/var/lib/goa4web`; explicit configuration takes precedence. Do not commit secrets stored at custom paths either.

If an image signing key has been committed, treat it as public. Check each running deployment's effective `IMAGE_SIGN_SECRET` or `IMAGE_SIGN_SECRET_FILE` (including configuration-file and command-line overrides) without printing the value. Rotate any matching active key. For a local development instance using the default file, stop the application, remove the old file, and restart to generate a fresh key. Rotation can invalidate existing signed image URLs. Removing a tracked key does not erase historical Git copies.
