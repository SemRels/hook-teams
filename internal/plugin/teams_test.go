// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The hook-teams Authors

package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestTeamsNotifierNotifyHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	err := NewTeamsNotifier(TeamsConfig{WebhookURL: srv.URL}).Notify(context.Background(), "v1.0.0", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected status")
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
	require.Equal(t, "notes", buildText("", "  notes  "))
	require.Equal(t, "@user@example.com", buildText(" @user@example.com ", ""))
	require.Equal(t, "@user@example.com\n\nnotes", buildText("user@example.com", "notes"))
}
