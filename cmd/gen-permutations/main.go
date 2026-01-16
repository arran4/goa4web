package main

import (
	"fmt"
	"os"
)

func main() {
	if err := os.MkdirAll("examples/og-images", 0755); err != nil {
		panic(err)
	}

	patterns := []string{
		"SierpinskiTriangle",
		"Checker",
		"Circle",
		"Crosshatch",
		"Stripe",
		"Polka",
		"Null",
	}

	colors := []string{
		"#FF0000",
		"#00FF00",
		"#0000FF",
		"#FFFF00",
		"#00FFFF",
		"#FF00FF",
	}

	for _, pattern := range patterns {
		for _, fg := range colors {
			for _, bg := range colors {
				if fg == bg {
					continue
				}
				for _, rpg := range []bool{false, true} {
					filename := fmt.Sprintf("examples/og-images/%s-%s-%s-%t.png", pattern, fg[1:], bg[1:], rpg)
					fmt.Printf("./goa4web gen-og-image --pattern %s --fg-color %s --bg-color %s --rpg-theme=%t --output %s\n", pattern, fg, bg, rpg, filename)
				}
			}
		}
	}
}
