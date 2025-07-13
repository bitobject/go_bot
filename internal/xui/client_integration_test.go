//go:build integration

package xui

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	url := os.Getenv("XUI_URL")
	username := os.Getenv("XUI_USERNAME")
	password := os.Getenv("XUI_PASSWORD")
	testEmail := os.Getenv("XUI_TEST_EMAIL")

	if url == "" || username == "" || password == "" {
		t.Skip("XUI_URL, XUI_USERNAME, or XUI_PASSWORD not set, skipping integration test.")
	}

	// Use a discard logger to keep test output clean from client's internal logs
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	client := NewClient(url, username, password, logger)

	t.Run("Login and get session", func(t *testing.T) {
		traffics, err := client.GetClientTraffics(context.Background(), "test@example.com")
		if err != nil {
			require.ErrorContains(t, err, "API returned an empty response body", "On login failure, a specific error is expected")
		} else {
			require.Empty(t, traffics, "On successful login, expected empty result for a non-existent user")
		}
	})

	if testEmail == "" {
		t.Skip("XUI_TEST_EMAIL not set, skipping client traffic tests")
	}

	t.Run("Get existing client", func(t *testing.T) {
		traffics, err := client.GetClientTraffics(context.Background(), testEmail)
		require.NoError(t, err)
		require.NotEmpty(t, traffics, "Expected to find at least one client traffic record, but got none")

		// If we found the client, log some details in a readable format
		if len(traffics) > 0 {
			traffic := traffics[0]
			t.Logf("Client %s: Up: %.2f MB, Down: %.2f MB, Total: %.2f GB, Expires: %s",
				traffic.Email,
				float64(traffic.Up)/(1024*1024),
				float64(traffic.Down)/(1024*1024),
				float64(traffic.Total)/(1024*1024*1024),
				time.Unix(traffic.ExpiryTime/1000, 0).Format("2006-01-02"),
			)
		}
	})

	t.Run("Get non-existent client", func(t *testing.T) {
		traffics, err := client.GetClientTraffics(context.Background(), "nonexistent-email-12345@example.com")
		require.NoError(t, err)
		require.Empty(t, traffics, "Expected to find no traffic records for a non-existent client")
	})
}
