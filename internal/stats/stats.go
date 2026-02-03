package stats

import "sync/atomic"

var (
	AutoSubscribePreferenceFailures atomic.Int64
)

func IncrementAutoSubscribePreferenceFailures() {
	AutoSubscribePreferenceFailures.Add(1)
}
