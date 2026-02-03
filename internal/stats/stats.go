package stats

import (
	"sync/atomic"
	"time"
)

var (
	AutoSubscribePreferenceFailures atomic.Int64
	// StartTime marks when the server began running.
	StartTime time.Time
)

func IncrementAutoSubscribePreferenceFailures() {
	AutoSubscribePreferenceFailures.Add(1)
}
