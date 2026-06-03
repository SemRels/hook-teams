// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestTeamsNotifierNotifySuccess(t *testing.T) {
	t.Parallel()

	var payload teamsPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewTeamsNotifier(TeamsConfig{WebhookURL: srv.URL})
	require.NoError(t, n.Notify(context.Background(), "v1.2.3", "- feature"))

	require.Equal(t, "MessageCard", payload.Type)
	require.Equal(t, "http://schema.org/extensions", payload.Context)
	require.Equal(t, "0078D7", payload.ThemeColor)
	require.Equal(t, "New release v1.2.3", payload.Summary)
	require.Len(t, payload.Sections, 1)
	require.Equal(t, "🚀 New Release", payload.Sections[0].ActivityTitle)
	require.Equal(t, "v1.2.3", payload.Sections[0].ActivitySubtitle)
	require.Len(t, payload.Sections[0].Facts, 1)
	require.Equal(t, "Version", payload.Sections[0].Facts[0].Name)
	require.Equal(t, "v1.2.3", payload.Sections[0].Facts[0].Value)
	require.Equal(t, "- feature", payload.Sections[0].Text)
}

func TestTeamsNotifierNotifyRequiresWebhook(t *testing.T) {
	t.Parallel()

	err := NewTeamsNotifier(TeamsConfig{}).Notify(context.Background(), "v1.0.0", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "webhook URL")
}

func TestTeamsNotifierNotifyRetriesOnServerError(t *testing.T) {
	var attempts int
	var logs bytes.Buffer
	oldWriter := retryLogWriter
	retryLogWriter = &logs
	defer func() { retryLogWriter = oldWriter }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewTeamsNotifier(TeamsConfig{WebhookURL: srv.URL, MaxRetries: 2, RetryDelay: time.Millisecond})
	require.NoError(t, n.Notify(context.Background(), "v1.0.0", ""))
	require.Equal(t, 3, attempts)
	require.Contains(t, logs.String(), "retry attempt 1/2")
	require.Contains(t, logs.String(), "retry attempt 2/2")
}

func TestTeamsNotifierNotifyDoesNotRetryOnClientError(t *testing.T) {
	t.Parallel()

	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	err := NewTeamsNotifier(TeamsConfig{WebhookURL: srv.URL, MaxRetries: 3, RetryDelay: time.Millisecond}).Notify(context.Background(), "v1.0.0", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected status 400")
	require.Equal(t, 1, attempts)
}

func TestTeamsNotifierNotifyRetriesOnNetworkError(t *testing.T) {
	t.Parallel()

	var attempts int
	n := NewTeamsNotifier(TeamsConfig{WebhookURL: "https://teams.example.test", MaxRetries: 2, RetryDelay: time.Millisecond})
	n.client = &http.Client{
		Timeout: defaultHTTPTimeout,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, &net.DNSError{Err: "temporary failure", IsTemporary: true}
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		}),
	}

	require.NoError(t, n.Notify(context.Background(), "v1.0.0", ""))
	require.Equal(t, 3, attempts)
}

func TestTeamsNotifierCustomTitleThemeAndMention(t *testing.T) {
	t.Parallel()

	var payload teamsPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	n := NewTeamsNotifier(TeamsConfig{
		WebhookURL: srv.URL,
		Title:      "Ship it",
		ThemeColor: "#FFAA00",
		Mention:    "user@example.com",
	})
	require.NoError(t, n.Notify(context.Background(), "v2.0.0", "- bugfix"))

	require.Equal(t, "FFAA00", payload.ThemeColor)
	require.Equal(t, "Ship it", payload.Sections[0].ActivityTitle)
	require.Equal(t, "@user@example.com\n\n- bugfix", payload.Sections[0].Text)
}

func TestTeamsNotifierDefaultsAndHelpers(t *testing.T) {
	t.Parallel()

	n := NewTeamsNotifier(TeamsConfig{WebhookURL: "https://example.test/webhook"})
	require.Equal(t, defaultHTTPTimeout, n.client.Timeout)
	require.Equal(t, "🚀 New Release", n.cfg.Title)
	require.Equal(t, "0078D7", n.cfg.ThemeColor)
	require.Equal(t, DefaultMaxRetries, n.cfg.MaxRetries)
	require.Equal(t, DefaultRetryDelay, n.cfg.RetryDelay)
	require.Equal(t, "notes", buildText("", "  notes  "))
	require.Equal(t, "@user@example.com", buildText(" @user@example.com ", ""))
	require.Equal(t, "@user@example.com\n\nnotes", buildText("user@example.com", "notes"))
}
