package subscriptions

import (
	"testing"
)

func BenchmarkMatchDefinition(b *testing.B) {
	// Sample patterns that match various definitions
	patterns := []string{
		"create thread:/forum/topic/123",
		"reply:/forum/topic/123/thread/456",
		"post:/blog/some-blog-post",
		"register:/auth/register",
		"unknown:/some/random/pattern", // Non-matching case
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range patterns {
			MatchDefinition(p)
		}
	}
}
