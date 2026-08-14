package tasks

import (
	"github.com/arran4/goa4web/config"
	"net/http"
)

type StartRemoteImageCacheFetchTask struct {
	TaskString
	Config     *config.RuntimeConfig
	ID         string
	SourceURL  string
	HTTPClient *http.Client
}

func (t *StartRemoteImageCacheFetchTask) Matcher() interface{} { return nil }

type ProcessImageTask struct {
	TaskString
	Config *config.RuntimeConfig
	ShaHex string
	Ext    string
}

func (t *ProcessImageTask) Matcher() interface{} { return nil }
