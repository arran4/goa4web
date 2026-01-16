package jmap

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

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
	Client    *http.Client
}

func NewProvider(endpoint, username, password, accountID, identity, from string, client *http.Client) Provider {
	return Provider{
		Endpoint:  endpoint,
		Username:  username,
		Password:  password,
		AccountID: accountID,
		Identity:  identity,
		From:      from,
		Client:    client,
	}
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
	resp, err := j.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}
	var up struct {
		BlobID string `json:"blobId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&up); err != nil {
		return err
	}

	// Resolve a mailbox to import into (Drafts, Outbox, Sent, or Inbox)
	mailboxID, err := j.getBestMailboxID(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve mailbox for import: %w", err)
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
							"mailboxIds": map[string]bool{mailboxID: true},
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
	resp, err = j.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("jmap send failed: %s", resp.Status)
	}

	var res struct {
		MethodResponses [][]interface{} `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	for _, mr := range res.MethodResponses {
		if len(mr) < 2 {
			continue
		}
		methodName, ok := mr[0].(string)
		if !ok {
			continue
		}
		args, ok := mr[1].(map[string]interface{})
		if !ok {
			continue
		}

		if methodName == "Email/import" {
			if notCreated, ok := args["notCreated"].(map[string]interface{}); ok && len(notCreated) > 0 {
				return fmt.Errorf("email import failed (notCreated): %v", notCreated)
			}
			if notImported, ok := args["notImported"].(map[string]interface{}); ok && len(notImported) > 0 {
				return fmt.Errorf("email import failed (notImported): %v", notImported)
			}
		}
		if methodName == "EmailSubmission/set" {
			if notCreated, ok := args["notCreated"].(map[string]interface{}); ok && len(notCreated) > 0 {
				return fmt.Errorf("email submission failed: %v", notCreated)
			}
		}
	}
	return nil
}

func (j Provider) TestConfig(ctx context.Context) error {
	fmt.Printf("Performing JMAP discovery for endpoint: %s\n", j.Endpoint)
	session, err := DiscoverSession(ctx, j.Client, j.Endpoint, j.Username, j.Password)
	if err != nil {
		return fmt.Errorf("failed to discover JMAP session: %w", err)
	}
	acc := SelectAccountID(session)
	id := SelectIdentityID(session)
	if id == "" {
		id, err = DiscoverIdentityID(ctx, j.Client, session.APIURL, j.Username, j.Password, acc)
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
		var session *SessionResponse
		var err error
		// Retry JMAP session discovery as it might be flaky on startup (e.g. 500 errors)
		for i := 0; i < 5; i++ {
			session, err = DiscoverSession(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass)
			if err == nil {
				break
			}
			fmt.Printf("JMAP discovery attempt %d failed: %v\n", i+1, err)
			// Only sleep if we are going to retry
			if i < 4 {
				time.Sleep(2 * time.Second)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("Email disabled: failed to discover JMAP session: %v", err)
		}
		if acc == "" {
			acc = SelectAccountID(session)
		}
		if id == "" {
			id = SelectIdentityID(session)
		}
		// Use session.APIURL if it looks valid, otherwise fallback to ep if using custom path
		apiURL := session.APIURL

		// If an override is provided, force it to be the API URL and the final endpoint
		if override := strings.TrimSpace(cfg.EmailJMAPEndpointOverride); override != "" {
			apiURL = override
			ep = override
		} else if ep != "" {
			u, _ := url.Parse(ep)
			if u != nil && u.Path != "" && u.Path != "/" {
				// If the user provided a custom path endpoint, use it (or prefer it)
				// But we should stick to session.APIURL if it's authoritative.
				// However, observed issue is session.APIURL might be internal or unreachable.
				// So if session discovery succeeded via 'ep', we might trust 'ep' more if it has a path?
				// For now, let's stick to the previous fix which used 'ep' for identity discovery.
				apiURL = ep
			}
		} else if session.APIURL != "" {
			// If no override and no custom path provided, default to session's API URL
			ep = session.APIURL
		}

		if id == "" && acc != "" {
			// Try to fetch identities via API
			fetchedId, err := DiscoverIdentityID(context.Background(), httpClient, apiURL, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
			if err == nil && fetchedId != "" {
				id = fetchedId
			}
		}
	}

	if acc == "" || id == "" {
		return nil, fmt.Errorf("Email disabled: %s or %s not set and could not be discovered", config.EnvJMAPAccount, config.EnvJMAPIdentity)
	}
	// Ensure we use the correct endpoint for the provider calls
	if ep == "" {
		// This case shouldn't happen based on above check, but logical fallback
		// if we started with empty ep and discovered it (not possible with current logic)
	}

	// If the user specified an endpoint with a path (e.g. /jmap/), we should use it.
	// DiscoverSession might have returned a session with an APIURL.
	// If the user explicitly configured an endpoint, we usually trust it.
	// But standard JMAP says use the one from Session.
	// In the failing case, the user likely set the endpoint manually to the API endpoint.
	// So let's prefer 'ep' if it was working for discovery.

	return Provider{
		Endpoint:  ep,
		Username:  cfg.EmailJMAPUser,
		Password:  cfg.EmailJMAPPass,
		AccountID: acc,
		Identity:  id,
		From:      cfg.EmailFrom,
		Client:    httpClient,
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
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("jmap session discovery failed: %s: %s", resp.Status, string(b))
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
	// If the endpoint already has a path (other than root), assume it is the full URL or a custom well-known location substitute
	// Actually, strictly speaking, well-known is a GET. API endpoint is usually POST.
	// But often they can be the same or related.
	// If the user provides `https://host/jmap/`, we probably should try `https://host/jmap/` for the session resource
	// instead of `https://host/.well-known/jmap`.
	if u.Path != "" && u.Path != "/" {
		return endpoint, nil
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

type EmailHeader struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	From    []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"from"`
	ReceivedAt string `json:"receivedAt"`
}

func (j Provider) GetInboxID(ctx context.Context) (string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Mailbox/query",
				map[string]interface{}{
					"accountId": j.AccountID,
					"filter":    map[string]interface{}{"role": "inbox"},
				},
				"c1",
			},
		},
	}
	return j.extractIDFromResponse(ctx, payload, "Mailbox/query")
}

func (j Provider) QueryInbox(ctx context.Context, inboxID string, limit int) ([]string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Email/query",
				map[string]interface{}{
					"accountId": j.AccountID,
					"filter":    map[string]interface{}{"inMailbox": inboxID},
					"sort":      []interface{}{map[string]interface{}{"property": "receivedAt", "isAscending": false}},
					"limit":     limit,
				},
				"c1",
			},
		},
	}
	return j.extractIDsFromResponse(ctx, payload, "Email/query")
}

func (j Provider) GetMessages(ctx context.Context, ids []string) ([]EmailHeader, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Email/get",
				map[string]interface{}{
					"accountId":  j.AccountID,
					"ids":        ids,
					"properties": []string{"id", "subject", "from", "receivedAt"},
				},
				"c1",
			},
		},
	}

	resp, err := j.doCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jmapResp struct {
		MethodResponses [][]interface{} `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jmapResp); err != nil {
		return nil, err
	}

	for _, methodResponse := range jmapResp.MethodResponses {
		if len(methodResponse) >= 2 && methodResponse[0] == "Email/get" {
			args, ok := methodResponse[1].(map[string]interface{})
			if !ok {
				continue
			}
			list, ok := args["list"].([]interface{})
			if !ok {
				continue
			}
			var emails []EmailHeader
			for _, item := range list {
				b, _ := json.Marshal(item)
				var e EmailHeader
				if err := json.Unmarshal(b, &e); err == nil {
					emails = append(emails, e)
				}
			}
			return emails, nil
		}
	}
	return nil, nil
}

func (j Provider) GetAllMessages(ctx context.Context, limit int) ([]string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Email/query",
				map[string]interface{}{
					"accountId": j.AccountID,
					"sort":      []interface{}{map[string]interface{}{"property": "receivedAt", "isAscending": false}},
					"limit":     limit,
				},
				"c1",
			},
		},
	}
	return j.extractIDsFromResponse(ctx, payload, "Email/query")
}

func (j Provider) getBestMailboxID(ctx context.Context) (string, error) {
	// Try Outbox, then Sent first for sending. Fallback to Drafts or Inbox.
	for _, role := range []string{"outbox", "sent", "drafts", "inbox"} {
		id, err := j.getMailboxIDByRole(ctx, role)
		if err == nil && id != "" {
			return id, nil
		}
	}
	// Fallback: Get ANY mailbox
	return j.getAnyMailboxID(ctx)
}

func (j Provider) getMailboxIDByRole(ctx context.Context, role string) (string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Mailbox/query",
				map[string]interface{}{
					"accountId": j.AccountID,
					"filter":    map[string]interface{}{"role": role},
					"limit":     1,
				},
				"c1",
			},
		},
	}
	return j.extractIDFromResponse(ctx, payload, "Mailbox/query")
}

func (j Provider) getAnyMailboxID(ctx context.Context) (string, error) {
	payload := map[string]interface{}{
		"using": []string{coreCapabilityURN, mailCapabilityURN},
		"methodCalls": [][]interface{}{
			{
				"Mailbox/query",
				map[string]interface{}{
					"accountId": j.AccountID,
					"limit":     1,
				},
				"c1",
			},
		},
	}
	return j.extractIDFromResponse(ctx, payload, "Mailbox/query")
}

func (j Provider) extractIDFromResponse(ctx context.Context, payload interface{}, methodName string) (string, error) {
	ids, err := j.extractIDsFromResponse(ctx, payload, methodName)
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}
	return "", nil
}

func (j Provider) extractIDsFromResponse(ctx context.Context, payload interface{}, methodName string) ([]string, error) {
	resp, err := j.doCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jmapResp struct {
		MethodResponses [][]interface{} `json:"methodResponses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jmapResp); err != nil {
		return nil, err
	}
	for _, methodResponse := range jmapResp.MethodResponses {
		if len(methodResponse) >= 2 && methodResponse[0] == methodName {
			args, ok := methodResponse[1].(map[string]interface{})
			if !ok {
				continue
			}
			list, ok := args["ids"].([]interface{})
			if !ok {
				continue
			}
			var ids []string
			for _, item := range list {
				if id, ok := item.(string); ok {
					ids = append(ids, id)
				}
			}
			return ids, nil
		}
	}
	return nil, nil
}

func (j Provider) doCall(ctx context.Context, payload interface{}) (*http.Response, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, j.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(j.Username, j.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := j.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("jmap call failed: %s", resp.Status)
	}
	return resp, nil
}
