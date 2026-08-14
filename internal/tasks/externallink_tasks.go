package tasks

import (
	"github.com/arran4/goa4web/config"
)

type ReloadExternalLinkTask struct {
	TaskString
	URL    string
	Config *config.RuntimeConfig
}

func (t *ReloadExternalLinkTask) Matcher() interface{} { return nil }

type PrefetchExternalLinkTask struct {
	TaskString
	URL    string
	Config *config.RuntimeConfig
}

func (t *PrefetchExternalLinkTask) Matcher() interface{} { return nil }
