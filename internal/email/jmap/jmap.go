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

func (j Provider) TestConfig(ctx context.Context) error {
	fmt.Printf("Performing JMAP discovery for endpoint: %s\n", j.Endpoint)
	session, err := DiscoverSession(ctx, j.client, j.Endpoint, j.Username, j.Password)
	if err != nil {
		return fmt.Errorf("failed to discover JMAP session: %w", err)
	}
	acc := SelectAccountID(session)
	id := SelectIdentityID(session)
	if id == "" {
		id, err = DiscoverIdentityID(ctx, j.client, session.APIURL, j.Username, j.Password, acc)
		if err != nil {
			fmt.Printf("failed to discover Identity ID via API: %v\n", err)
		}
	}
	fmt.Printf("Discovered Account ID: %s\n", acc)
	fmt.Printf("Discovered Identity ID: %s\n", id)
	return nil
}

func providerFromConfig(cfg *config.RuntimeConfig) (email.Provider, error) {
	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		return nil, fmt.Errorf("Email disabled: %s not set", config.EnvJMAPEndpoint)
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
		session, err := DiscoverSession(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass)
		if err != nil {
			return nil, fmt.Errorf("Email disabled: failed to discover JMAP session: %v", err)
		}
		if acc == "" {
			acc = SelectAccountID(session)
		}
		if id == "" {
			id = SelectIdentityID(session)
		}
		if ep == "" {
			ep = session.APIURL
		}
		if id == "" && acc != "" {
			// Try to fetch identities via API
			fetchedId, err := DiscoverIdentityID(context.Background(), httpClient, session.APIURL, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
			if err == nil && fetchedId != "" {
				id = fetchedId
			}
		}
	}

	if acc == "" || id == "" {
		return nil, fmt.Errorf("Email disabled: %s or %s not set and could not be discovered", config.EnvJMAPAccount, config.EnvJMAPIdentity)
	}
	return Provider{
		Endpoint:  ep,
		Username:  cfg.EmailJMAPUser,
		Password:  cfg.EmailJMAPPass,
		AccountID: acc,
		Identity:  id,
		From:      cfg.EmailFrom,
		client:    httpClient,
	}, nil
}

// Register registers the JMAP provider.
func Register(r *email.Registry) { r.RegisterProvider("jmap", providerFromConfig) }

// mailCapabilityURN identifies the JMAP mail capability.
const mailCapabilityURN = "urn:ietf:params:jmap:mail"

// submissionCapabilityURN identifies the JMAP submission capability.
const submissionCapabilityURN = "urn:ietf:params:jmap:submission"

// coreCapabilityURN identifies the JMAP core capability.
const coreCapabilityURN = "urn:ietf:params:jmap:core"

// sieveCapabilityURN identifies the JMAP sieve capability.
const sieveCapabilityURN = "urn:ietf:params:jmap:sieve"

type Account struct {
	Name         string                 `json:"name"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

type SessionResponse struct {
	APIURL          string             `json:"apiUrl"`
	PrimaryAccounts map[string]string  `json:"primaryAccounts"`
	DefaultIdentity map[string]string  `json:"defaultIdentity"`
	Accounts        map[string]Account `json:"accounts"`
}

func DiscoverSession(ctx context.Context, client *http.Client, endpoint, username, password string) (*SessionResponse, error) {
	wellKnown, err := JmapWellKnownURL(endpoint)
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
	var session SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}
	return &session, nil
}

func JmapWellKnownURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid JMAP endpoint")
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/.well-known/jmap"}).String(), nil
}

func SelectAccountID(session *SessionResponse) string {
	if session == nil {
		return ""
	}
	if acc := session.PrimaryAccounts[mailCapabilityURN]; acc != "" {
		return acc
	}
	if acc := session.PrimaryAccounts[sieveCapabilityURN]; acc != "" {
		return acc
	}
	// Fallback to checking Accounts for mail capability
	for id, acc := range session.Accounts {
		if _, ok := acc.Capabilities[mailCapabilityURN]; ok {
			return id
		}
	}
	for _, acc := range session.PrimaryAccounts {
		if acc != "" {
			return acc
		}
	}
	return ""
}

func SelectIdentityID(session *SessionResponse) string {
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

func DiscoverIdentityID(ctx context.Context, client *http.Client, apiURL, username, password, accountID string) (string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, submissionCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Identity/get",
				map[string]interface{}{
					"accountId": accountID,
				},
				"c1",
			},
		},
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("jmap identity discovery failed: %s", resp.Status)
	}

	var jmapResp struct {
		MethodResponses [][]interface{} `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jmapResp); err != nil {
		return "", err
	}

	for _, methodResponse := range jmapResp.MethodResponses {
		if len(methodResponse) >= 2 && methodResponse[0] == "Identity/get" {
			args, ok := methodResponse[1].(map[string]interface{})
			if !ok {
				continue
			}
			list, ok := args["list"].([]interface{})
			if !ok {
				continue
			}
			for _, item := range list {
				identity, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if id, ok := identity["id"].(string); ok && id != "" {
					return id, nil
				}
			}
		}
	}

	return "", nil
}
