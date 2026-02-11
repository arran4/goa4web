package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminLinksToolsPage renders the admin link signing utilities page.
func AdminLinksToolsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Link Signing Tools"

	data := struct {
		SignURL       string
		SignDuration  string
		SignNoExpiry  bool
		SignOutputURL string
		SignOutputSig string
		SignOutputTS  string
		SignError     string
		VerifyURL     string
		VerifySig     string
		VerifyTS      string
		VerifyResult  string
		VerifyError   string
	}{
		SignDuration: "24h",
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		if action == "sign" {
			data.SignURL = strings.TrimSpace(r.FormValue("sign_url"))
			data.SignDuration = strings.TrimSpace(r.FormValue("sign_duration"))
			data.SignNoExpiry = r.FormValue("sign_no_expiry") == "on"

			key, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cd.Config.LinkSignSecret, cd.Config.LinkSignSecretFile)
			if err != nil {
				data.SignError = fmt.Sprintf("link sign secret: %v", err)
			} else if data.SignURL == "" {
				data.SignError = "URL is required."
			} else {
				signData := "link:" + data.SignURL
				var opts []sign.SignOption
				if data.SignNoExpiry {
					opts = append(opts, sign.WithOutNonce())
				} else {
					if data.SignDuration == "" {
						data.SignDuration = "24h"
					}
					d, err := time.ParseDuration(data.SignDuration)
					if err != nil {
						data.SignError = fmt.Sprintf("parse duration: %v", err)
					} else {
						expiry := time.Now().Add(d)
						opts = append(opts, sign.WithExpiry(expiry))
						data.SignOutputTS = fmt.Sprintf("%d", expiry.Unix())
					}
				}

				if data.SignError == "" {
					data.SignOutputSig = sign.Sign(signData, key, opts...)
					signedURL, err := sign.AddQuerySig(cd.Config.HTTPHostname+"/goto?u="+data.SignURL, data.SignOutputSig, opts...)
					if err != nil {
						data.SignError = fmt.Sprintf("signed url: %v", err)
					} else {
						data.SignOutputURL = signedURL
					}
				}
			}
		}

		if action == "verify" {
			data.VerifyURL = strings.TrimSpace(r.FormValue("verify_url"))
			data.VerifySig = strings.TrimSpace(r.FormValue("verify_sig"))
			data.VerifyTS = strings.TrimSpace(r.FormValue("verify_ts"))

			key, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cd.Config.LinkSignSecret, cd.Config.LinkSignSecretFile)
			if err != nil {
				data.VerifyError = fmt.Sprintf("link sign secret: %v", err)
			} else if data.VerifyURL == "" || data.VerifySig == "" {
				data.VerifyError = "URL and signature are required."
			} else {
				var opts []sign.SignOption
				validTS := true
				if data.VerifyTS != "" {
					if tsInt, err := strconv.ParseInt(data.VerifyTS, 10, 64); err == nil {
						opts = append(opts, sign.WithExpiryTimeUnix(tsInt))
					} else {
						data.VerifyError = "Invalid timestamp format"
						validTS = false
					}
				}

				if validTS {
					if err := sign.Verify(data.VerifyURL, data.VerifySig, key, opts...); err == nil {
						data.VerifyResult = "valid"
					} else {
						data.VerifyResult = "invalid"
					}
				}
			}
		}
	}

	AdminLinksToolsPageTmpl.Handle(w, r, data)
}

// AdminLinksToolsPageTmpl renders the link signing tools page.
const AdminLinksToolsPageTmpl tasks.Template = "admin/linksToolsPage.gohtml"
