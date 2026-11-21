package externallinkworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	domainexternallink "github.com/arran4/goa4web/internal/externallink"
	"github.com/arran4/goa4web/internal/opengraph"
)

// Worker listens for new content and fetches metadata for external links.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier, cfg *config.RuntimeConfig) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)

	for {
		select {
		case env, ok := <-ch:
			if !ok {
				return
			}
			func() {
				defer env.Ack()
				evt, ok := env.Msg.(eventbus.TaskEvent)
				if !ok {
					return
				}
				if evt.Outcome != eventbus.TaskOutcomeSuccess {
					return
				}
				body, ok := evt.Data["Body"].(string)
				if !ok || body == "" {
					return
				}

				root, err := a4code.ParseString(body)
				if err != nil {
					return
				}

				_ = ast.Walk(root, func(n ast.Node) error {
					link, ok := n.(*ast.Link)
					if !ok {
						return nil
					}
					url := link.Href
					if url == "" {
						return nil
					}
					canonicalURL := common.CanonicalizeExternalURL(url)
					if canonicalURL == "" {
						return nil
					}

					linkID, err := domainexternallink.ResolveRecord(ctx, q, canonicalURL)
					if err != nil {
						log.Printf("resolve external link %s: %v", canonicalURL, err)
						return nil
					}
					if linkID == 0 {
						return nil
					}

					_, err = domainexternallink.RefreshMissingMetadata(ctx, q, linkID, canonicalURL,
						func(ctx context.Context, url string) (*opengraph.Info, error) {
							return domainexternallink.FetchMetadata(ctx, url, func(ctx context.Context, url string) (*opengraph.Info, error) {
								return opengraph.Fetch(ctx, url, nil)
							}, domainexternallink.Wait)
						},
						func(ctx context.Context, image string) (string, error) {
							return common.NewCoreData(ctx, q, cfg).DownloadAndCacheImage(image)
						})
					if err != nil {
						log.Printf("refresh external-link metadata %d: %v", linkID, err)
					}

					return nil
				})
			}()

		case <-ctx.Done():
			return
		}
	}
}
