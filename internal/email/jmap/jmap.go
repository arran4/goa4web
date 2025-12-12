package jmap

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
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
	client    *http.Client
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
	resp, err := j.client.Do(req)
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
	resp, err = j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("jmap send failed: %s", resp.Status)
	}
	return nil
}

func providerFromConfig(cfg *config.RuntimeConfig) email.Provider {
	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		fmt.Printf("Email disabled: %s not set\n", config.EnvJMAPEndpoint)
		return nil
	}
	acc := strings.TrimSpace(cfg.EmailJMAPAccount)
	id := strings.TrimSpace(cfg.EmailJMAPIdentity)

	httpClient := http.DefaultClient
	if cfg.EmailJMAPInsecure {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: tr}
	}

	if acc == "" || id == "" {
		session, err := discoverSession(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass)
		if err != nil {
			fmt.Printf("Email disabled: failed to discover JMAP session: %v\n", err)
			return nil
		}
		if acc == "" {
			acc = selectAccountID(session)
		}
		if id == "" {
			id = selectIdentityID(session)
		}
		if ep == "" {
			ep = session.APIURL
		}
	}

	if acc == "" || id == "" {
		fmt.Printf("Email disabled: %s or %s not set and could not be discovered\n", config.EnvJMAPAccount, config.EnvJMAPIdentity)
		return nil
	}
	return Provider{
		Endpoint:  ep,
		Username:  cfg.EmailJMAPUser,
		Password:  cfg.EmailJMAPPass,
		AccountID: acc,
		Identity:  id,
		From:      cfg.EmailFrom,
		client:    httpClient,
	}
}

// Register registers the JMAP provider.
func Register(r *email.Registry) { r.RegisterProvider("jmap", providerFromConfig) }

// mailCapabilityURN identifies the JMAP mail capability.
const mailCapabilityURN = "urn:ietf:params:jmap:mail"

// sieveCapabilityURN identifies the JMAP sieve capability.
const sieveCapabilityURN = "urn:ietf:params:jmap:sieve"

type sessionResponse struct {
	APIURL          string            `json:"apiUrl"`
	PrimaryAccounts map[string]string `json:"primaryAccounts"`
	DefaultIdentity map[string]string `json:"defaultIdentity"`
}

func discoverSession(ctx context.Context, client *http.Client, endpoint, username, password string) (*sessionResponse, error) {
	wellKnown, err := jmapWellKnownURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return nil, err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jmap session discovery failed: %s", resp.Status)
	}
	var session sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}
	return &session, nil
}

func jmapWellKnownURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid JMAP endpoint")
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/.well-known/jmap"}).String(), nil
}

func selectAccountID(session *sessionResponse) string {
	if session == nil {
		return ""
	}
	if acc := session.PrimaryAccounts[mailCapabilityURN]; acc != "" {
		return acc
	}
	if acc := session.PrimaryAccounts[sieveCapabilityURN]; acc != "" {
		return acc
	}
	for _, acc := range session.PrimaryAccounts {
		if acc != "" {
			return acc
		}
	}
	return ""
}

func selectIdentityID(session *sessionResponse) string {
	if session == nil {
		return ""
	}
	if id := session.DefaultIdentity[mailCapabilityURN]; id != "" {
		return id
	}
	for _, id := range session.DefaultIdentity {
		if id != "" {
			return id
		}
	}
	return ""
}
