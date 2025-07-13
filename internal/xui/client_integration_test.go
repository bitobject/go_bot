//go:build integration

package xui

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestClient_Integration is an integration test that verifies the XUI client can
// successfully authenticate and fetch data from a live 3x-UI service.
func TestClient_Integration(t *testing.T) {
	// Get configuration from environment variables
	url := os.Getenv("XUI_URL")
	username := os.Getenv("XUI_USERNAME")
	password := os.Getenv("XUI_PASSWORD")
	testEmail := os.Getenv("XUI_TEST_EMAIL")

	// Skip the test if the required environment variables are not set
	if url == "" || username == "" || password == "" {
		t.Skip("XUI_URL, XUI_USERNAME, and XUI_PASSWORD environment variables must be set for integration tests")
	}

	// Create a new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create a new client
	client := NewClient(url, username, password, logger)

	t.Run("Login and get session", func(t *testing.T) {
		// The client should automatically log in on the first request
		// We'll test this by making a request that requires authentication
		_, err := client.GetClientTraffics(context.Background(), "test@example.com")

		// If we get an "unauthorized" error, the login failed
		require.NoError(t, err, "Failed to authenticate with 3x-UI API")
	})

	if testEmail == "" {
		t.Skip("XUI_TEST_EMAIL not set, skipping client traffic tests")
	}

	t.Run("Get existing client", func(t *testing.T) {
		// Fetch the client's traffic data
		traffics, err := client.GetClientTraffics(context.Background(), testEmail)
		require.NoError(t, err)
		require.NotEmpty(t, traffics, "Expected to find at least one client traffic record, but got none")

		// Log the results for debugging
		t.Logf("Found %d traffic records for email %s", len(traffics), testEmail)
		for i, traffic := range traffics {
			t.Logf("Traffic %d: %+v", i+1, traffic)
		}

		// If we found the client, log some details
		if len(traffics) > 0 {
			traffic := traffics[0]
			t.Logf("Client %s: Up: %d MB, Down: %d MB, Total: %d MB, Expires: %s",
				traffic.Email,
				traffic.Up/(1024*1024),
				traffic.Down/(1024*1024),
				traffic.Total/(1024*1024),
				time.Unix(traffic.ExpiryTime/1000, 0).Format(time.RFC3339),
			)
		}
	})

	t.Run("Get non-existent client", func(t *testing.T) {
		// Test with a non-existent email that should return an empty result
		nonExistentEmail := "nonexistent-email-12345@example.com"
		traffics, err := client.GetClientTraffics(context.Background(), nonExistentEmail)
		require.NoError(t, err)
		require.Empty(t, traffics, "Expected no traffic records for a non-existent client")
	})
}
