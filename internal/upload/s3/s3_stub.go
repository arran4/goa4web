//go:build !s3

package s3

// Built indicates whether the S3 provider is compiled in.
const Built = false

// Register is a no-op when the s3 build tag is not present.
func Register() {}
