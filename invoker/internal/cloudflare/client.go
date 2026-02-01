package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.cloudflare.com/client/v4"

// Client provides minimal Cloudflare API access for tunnels + DNS.
type Client struct {
	accountID string
	apiToken  string
	zoneName  string
	zoneID    string
	baseURL   string
	client    *http.Client
}

// Tunnel represents a Cloudflare tunnel.
type Tunnel struct {
	ID    string
	Name  string
	Token string
}

// NewClient creates a Cloudflare API client.
func NewClient(accountID, apiToken, zoneName, zoneID string) *Client {
	return &Client{
		accountID: strings.TrimSpace(accountID),
		apiToken:  strings.TrimSpace(apiToken),
		zoneName:  strings.TrimSpace(zoneName),
		zoneID:    strings.TrimSpace(zoneID),
		baseURL:   defaultBaseURL,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// CreateTunnel creates a named tunnel and returns its token.
func (c *Client) CreateTunnel(ctx context.Context, name string) (*Tunnel, error) {
	req := map[string]any{
		"name":       name,
		"config_src": "cloudflare",
	}

	var resp struct {
		Success bool       `json:"success"`
		Errors  []apiError `json:"errors"`
		Result  struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Token string `json:"token"`
		} `json:"result"`
	}

	path := fmt.Sprintf("/accounts/%s/cfd_tunnel", c.accountID)
	if err := c.doJSON(ctx, http.MethodPost, path, req, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("cloudflare create tunnel failed: %s", formatErrors(resp.Errors))
	}
	if resp.Result.ID == "" || resp.Result.Token == "" {
		return nil, fmt.Errorf("cloudflare create tunnel missing id or token")
	}
	return &Tunnel{ID: resp.Result.ID, Name: resp.Result.Name, Token: resp.Result.Token}, nil
}

// EnsureDNSRecord ensures a CNAME exists for <subdomain>.<zone> pointing to the tunnel.
func (c *Client) EnsureDNSRecord(ctx context.Context, subdomain, tunnelID string) error {
	zoneID, err := c.ensureZoneID(ctx)
	if err != nil {
		return err
	}

	fqdn := c.fqdn(subdomain)
	target := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)

	record, err := c.getDNSRecord(ctx, zoneID, fqdn)
	if err != nil {
		return err
	}

	if record != nil {
		if strings.EqualFold(record.Content, target) && record.Proxied {
			return nil
		}
		return c.updateDNSRecord(ctx, zoneID, record.ID, fqdn, target)
	}

	return c.createDNSRecord(ctx, zoneID, fqdn, target)
}

type dnsRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

func (c *Client) getDNSRecord(ctx context.Context, zoneID, fqdn string) (*dnsRecord, error) {
	params := url.Values{}
	params.Set("type", "CNAME")
	params.Set("name", fqdn)
	path := fmt.Sprintf("/zones/%s/dns_records?%s", zoneID, params.Encode())

	var resp struct {
		Success bool        `json:"success"`
		Errors  []apiError  `json:"errors"`
		Result  []dnsRecord `json:"result"`
	}

	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("cloudflare list dns records failed: %s", formatErrors(resp.Errors))
	}
	if len(resp.Result) == 0 {
		return nil, nil
	}
	return &resp.Result[0], nil
}

func (c *Client) createDNSRecord(ctx context.Context, zoneID, fqdn, target string) error {
	req := map[string]any{
		"type":    "CNAME",
		"name":    fqdn,
		"content": target,
		"proxied": true,
		"ttl":     1,
	}

	var resp struct {
		Success bool       `json:"success"`
		Errors  []apiError `json:"errors"`
		Result  dnsRecord  `json:"result"`
	}

	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	if err := c.doJSON(ctx, http.MethodPost, path, req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("cloudflare create dns record failed: %s", formatErrors(resp.Errors))
	}
	return nil
}

func (c *Client) updateDNSRecord(ctx context.Context, zoneID, recordID, fqdn, target string) error {
	req := map[string]any{
		"type":    "CNAME",
		"name":    fqdn,
		"content": target,
		"proxied": true,
		"ttl":     1,
	}

	var resp struct {
		Success bool       `json:"success"`
		Errors  []apiError `json:"errors"`
		Result  dnsRecord  `json:"result"`
	}

	path := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID)
	if err := c.doJSON(ctx, http.MethodPut, path, req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("cloudflare update dns record failed: %s", formatErrors(resp.Errors))
	}
	return nil
}

func (c *Client) ensureZoneID(ctx context.Context) (string, error) {
	if c.zoneID != "" {
		return c.zoneID, nil
	}
	if c.zoneName == "" {
		return "", fmt.Errorf("cloudflare zone name not configured")
	}

	params := url.Values{}
	params.Set("name", c.zoneName)
	path := fmt.Sprintf("/zones?%s", params.Encode())

	var resp struct {
		Success bool       `json:"success"`
		Errors  []apiError `json:"errors"`
		Result  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}

	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("cloudflare list zones failed: %s", formatErrors(resp.Errors))
	}
	if len(resp.Result) == 0 {
		return "", fmt.Errorf("cloudflare zone not found: %s", c.zoneName)
	}
	c.zoneID = resp.Result[0].ID
	return c.zoneID, nil
}

func (c *Client) fqdn(subdomain string) string {
	if strings.HasSuffix(subdomain, "."+c.zoneName) {
		return subdomain
	}
	return subdomain + "." + c.zoneName
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func formatErrors(errs []apiError) string {
	if len(errs) == 0 {
		return "unknown error"
	}
	parts := make([]string, 0, len(errs))
	for _, e := range errs {
		if e.Code != 0 {
			parts = append(parts, fmt.Sprintf("%d: %s", e.Code, e.Message))
			continue
		}
		parts = append(parts, e.Message)
	}
	return strings.Join(parts, "; ")
}

func (c *Client) doJSON(ctx context.Context, method, path string, reqBody any, respBody any) error {
	var body io.Reader
	if reqBody != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return err
		}
		body = buf
	}

	url := strings.TrimRight(c.baseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("cloudflare %s %s failed: %s", method, path, msg)
	}

	if respBody == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(respBody)
}
