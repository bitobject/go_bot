package xui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
func NewClient(url, username, password string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		url:        url,
		username:   username,
		password:   password,
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
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var loginResp loginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return err
	}
	if !loginResp.Success {
		return errors.New(loginResp.Msg)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" {
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	type apiResponse struct {
		Success bool             `json:"success"`
		Msg     string           `json:"msg"`
		Obj     []ClientTraffic `json:"obj"`
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("api error: %s", apiResp.Msg)
	}

	return apiResp.Obj, nil
}
