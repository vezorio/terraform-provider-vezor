package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the Vezor API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// Secret represents a secret from the API
type Secret struct {
	ID          string            `json:"id"`
	KeyName     string            `json:"key_name"`
	Value       string            `json:"value,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags"`
	Version     int               `json:"version"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// Group represents a group from the API
type Group struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// GroupSecrets represents the response from pulling group secrets
type GroupSecrets struct {
	Group   string            `json:"group"`
	Tags    map[string]string `json:"tags"`
	Secrets map[string]string `json:"secrets"`
	Count   int               `json:"count"`
}

// SecretsListResponse represents the response from listing secrets
type SecretsListResponse struct {
	Secrets []Secret `json:"secrets"`
	Total   int      `json:"total"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewClient creates a new Vezor API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	if params != nil && len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetSecret retrieves a secret by ID, optionally with a specific version
func (c *Client) GetSecret(secretID string, version *int) (*Secret, error) {
	endpoint := fmt.Sprintf("/api/v1/secrets/%s", secretID)
	params := url.Values{}
	if version != nil {
		params.Set("version", fmt.Sprintf("%d", *version))
	}

	body, err := c.doRequest("GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	var secret Secret
	if err := json.Unmarshal(body, &secret); err != nil {
		return nil, fmt.Errorf("failed to parse secret response: %w", err)
	}

	return &secret, nil
}

// ListSecrets lists secrets with optional tag filters
func (c *Client) ListSecrets(tags map[string]string, search string, limit int) (*SecretsListResponse, error) {
	params := url.Values{}
	for k, v := range tags {
		params.Set(k, v)
	}
	if search != "" {
		params.Set("search", search)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := c.doRequest("GET", "/api/v1/secrets", params)
	if err != nil {
		return nil, err
	}

	var resp SecretsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse secrets list response: %w", err)
	}

	return &resp, nil
}

// FindSecret finds a secret by name and tags
func (c *Client) FindSecret(name string, tags map[string]string) (*Secret, error) {
	// List secrets with the given tags and search for the name
	resp, err := c.ListSecrets(tags, name, 100)
	if err != nil {
		return nil, err
	}

	// Find exact match
	for _, s := range resp.Secrets {
		if strings.EqualFold(s.KeyName, name) {
			// Tags must match exactly
			if tagsMatch(s.Tags, tags) {
				// Get the full secret with value
				return c.GetSecret(s.ID, nil)
			}
		}
	}

	return nil, fmt.Errorf("secret '%s' not found with specified tags", name)
}

// GetGroup retrieves a group by name
func (c *Client) GetGroup(name string) (*Group, error) {
	endpoint := fmt.Sprintf("/api/v1/groups/%s", url.PathEscape(name))

	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("failed to parse group response: %w", err)
	}

	return &group, nil
}

// PullGroupSecrets retrieves all secrets for a group
func (c *Client) PullGroupSecrets(name string) (*GroupSecrets, error) {
	endpoint := fmt.Sprintf("/api/v1/groups/%s/secrets", url.PathEscape(name))
	params := url.Values{}
	params.Set("format", "json")

	body, err := c.doRequest("GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	var secrets GroupSecrets
	if err := json.Unmarshal(body, &secrets); err != nil {
		return nil, fmt.Errorf("failed to parse group secrets response: %w", err)
	}

	return &secrets, nil
}

// tagsMatch checks if secret tags contain all required tags
func tagsMatch(secretTags, requiredTags map[string]string) bool {
	for k, v := range requiredTags {
		if secretTags[k] != v {
			return false
		}
	}
	return true
}
