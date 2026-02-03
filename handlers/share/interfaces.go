package share

import "image"

type ImageGenerator interface {
	// Generate creates an image. implementations should type-switch on options.
	Generate(options ...interface{}) (image.Image, error)
	Name() string
}
