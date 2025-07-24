package jmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Provider sends mail via JMAP.
type Provider struct {
	Endpoint  string
	Username  string
	Password  string
	AccountID string
	Identity  string
	From      string
}

func (j Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	var msg bytes.Buffer
	msg.Write(rawEmailMessage)

	uploadURL := fmt.Sprintf("%s/upload/%s", strings.TrimRight(j.Endpoint, "/"), j.AccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)
	req.Header.Set("Content-Type", "message/rfc822")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}
	var up struct {
		BlobID string `json:"blobId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&up); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
		"methodCalls": [][]interface{}{
			{
				"Email/import",
				map[string]interface{}{
					"accountId": j.AccountID,
					"emails": map[string]interface{}{
						"msg": map[string]interface{}{
							"blobId":     up.BlobID,
							"mailboxIds": map[string]bool{"outbox": true},
						},
					},
				},
				"c1",
			},
			{
				"EmailSubmission/set",
				map[string]interface{}{
					"accountId": j.AccountID,
					"create": map[string]interface{}{
						"sub": map[string]interface{}{
							"emailId":    "#msg",
							"identityId": j.Identity,
						},
					},
				},
				"c2",
			},
		},
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, j.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("jmap send failed: %s", resp.Status)
	}
	return nil
}

func providerFromConfig(cfg config.RuntimeConfig) email.Provider {
	ep := cfg.EmailJMAPEndpoint
	if ep == "" {
		fmt.Printf("Email disabled: %s not set\n", config.EnvJMAPEndpoint)
		return nil
	}
	acc := cfg.EmailJMAPAccount
	id := cfg.EmailJMAPIdentity
	if acc == "" || id == "" {
		fmt.Printf("Email disabled: %s or %s not set\n", config.EnvJMAPAccount, config.EnvJMAPIdentity)
		return nil
	}
	return Provider{
		Endpoint:  ep,
		Username:  cfg.EmailJMAPUser,
		Password:  cfg.EmailJMAPPass,
		AccountID: acc,
		Identity:  id,
		From:      cfg.EmailFrom,
	}
}

// Register registers the JMAP provider.
func Register(r *email.Registry) { r.RegisterProvider("jmap", providerFromConfig) }
