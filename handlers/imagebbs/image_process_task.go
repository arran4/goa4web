package imagebbs

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
)

// ProcessImageTask handles background thumbnail generation.
type ProcessImageTask struct {
	tasks.TaskString
	Config *config.RuntimeConfig
	ShaHex string
	Ext    string
}

var _ tasks.Task = (*ProcessImageTask)(nil)
var _ tasks.BackgroundTasker = (*ProcessImageTask)(nil)

func (t *ProcessImageTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if t.Config == nil {
		return nil, fmt.Errorf("config missing")
	}
	return nil, imagesign.WriteImageBoardThumbnail(ctx, upload.ProviderFromConfig(t.Config), t.ShaHex+t.Ext)
}
