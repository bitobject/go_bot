package xui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a client for the 3x-ui API.
type Client struct {
	httpClient     *http.Client
	url            string
	username       string
	password       string
	sessionCookie  *http.Cookie
	sessionExpires time.Time
	logger         *slog.Logger
}

type loginResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

// ClientTraffic represents the traffic data for a client.
// This structure is based on the 3x-ui API response.
type ClientTraffic struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int    `json:"reset"`
}

// NewClient creates a new 3x-ui API client.
func NewClient(url, username, password string, logger *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		url:        url,
		username:   username,
		password:   password,
		logger:     logger,
	}
}

func (c *Client) login(ctx context.Context) error {
	loginReq := url.Values{"username": {c.username}, "password": {c.password}}
	b := strings.NewReader(loginReq.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/login", b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.logger.Debug("Login response received", "status", resp.Status, "headers", resp.Header, "body", string(body))

	var loginResp loginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return err
	}
	if !loginResp.Success {
		return errors.New(loginResp.Msg)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "3x-ui" {
			c.sessionCookie = cookie
			c.sessionExpires = cookie.Expires
			break // Found the cookie, no need to loop further
		}
	}

	if c.sessionCookie != nil {
		return nil
	}

	return errors.New("session cookie not found")
}

func (c *Client) loginIfNoCookie(ctx context.Context) error {
	if c.sessionCookie != nil && c.sessionExpires.After(time.Now()) {
		return nil
	}
	return c.login(ctx)
}

// GetClientTraffics fetches traffic data for a specific client by email.
func (c *Client) GetClientTraffics(ctx context.Context, email string) ([]ClientTraffic, error) {
	if err := c.loginIfNoCookie(ctx); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/panel/api/inbounds/getClientTraffics/%s", c.url, email)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.AddCookie(c.sessionCookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request to X-UI", "error", err, "email", email)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body from X-UI", "error", err, "email", email, "status", resp.Status)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("X-UI API response", "email", email, "status", resp.Status, "body", string(body))

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("X-UI API returned non-OK status", "email", email, "status_code", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("bad status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("API returned an empty response body, check credentials or User-Agent header")
	}

	type APIResponse struct {
		Success bool            `json:"success"`
		Msg     string          `json:"msg"`
		Obj     json.RawMessage `json:"obj"`
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.logger.Error("Failed to unmarshal X-UI API response", "error", err, "email", email, "body", string(body))
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("api error: %s", apiResp.Msg)
	}

	// Handle null or empty object case
	if string(apiResp.Obj) == "null" || len(apiResp.Obj) == 0 {
		return []ClientTraffic{}, nil
	}

	var traffics []ClientTraffic
	// Try to unmarshal as an array first
	if err := json.Unmarshal(apiResp.Obj, &traffics); err != nil {
		// If it's not an array, try to unmarshal as a single object
		var singleTraffic ClientTraffic
		if err2 := json.Unmarshal(apiResp.Obj, &singleTraffic); err2 != nil {
			// If both fail, return the original array unmarshal error
			return nil, fmt.Errorf("failed to unmarshal 'obj' field from API response: %w", err)
		}
		// If single object unmarshal succeeds, wrap it in a slice
		traffics = []ClientTraffic{singleTraffic}
	}

	return traffics, nil
}
