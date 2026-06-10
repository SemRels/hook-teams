// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The hook-teams Authors

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	plugin "github.com/SemRels/hook-teams/internal/plugin"
	"github.com/stretchr/testify/require"
)

type stubNotifier struct {
	cfg       plugin.TeamsConfig
	version   string
	changelog string
	err       error
	called    bool
}

func (s *stubNotifier) Notify(_ context.Context, version, changelog string) error {
	s.called = true
	s.version = version
	s.changelog = changelog
	return s.err
}

func TestRunSuccess(t *testing.T) {
	stub := &stubNotifier{}
	original := newNotifier
	newNotifier = func(cfg plugin.TeamsConfig) notifier {
		stub.cfg = cfg
		return stub
	}
	defer func() { newNotifier = original }()

	stderr := &bytes.Buffer{}
	exitCode := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://example.test/webhook",
		"SEMREL_VERSION":            "v1.2.3",
		"SEMREL_CHANGELOG":          "- feature",
		"SEMREL_PLUGIN_TITLE":       "Shipped",
		"SEMREL_PLUGIN_THEME_COLOR": "00FF00",
		"SEMREL_PLUGIN_MENTION":     "user@example.com",
	}), stderr)

	require.Equal(t, 0, exitCode)
	require.Equal(t, "plugin_schema_version=1\n", stderr.String())
	require.True(t, stub.called)
	require.Equal(t, "v1.2.3", stub.version)
	require.Equal(t, "- feature", stub.changelog)
	require.Equal(t, "Shipped", stub.cfg.Title)
	require.Equal(t, "00FF00", stub.cfg.ThemeColor)
	require.Equal(t, "user@example.com", stub.cfg.Mention)
}

func TestRunDryRun(t *testing.T) {
	stub := &stubNotifier{}
	original := newNotifier
	newNotifier = func(cfg plugin.TeamsConfig) notifier {
		stub.cfg = cfg
		return stub
	}
	defer func() { newNotifier = original }()

	stdout := captureStdout(t, func() {
		stderr := &bytes.Buffer{}
		exitCode := run(context.Background(), env(map[string]string{
			"SEMREL_PLUGIN_WEBHOOK_URL": "https://example.test/webhook",
			"SEMREL_TAG_NAME":           "v2.0.0",
			"SEMREL_DRY_RUN":            "true",
		}), stderr)
		require.Equal(t, 0, exitCode)
		require.Equal(t, "plugin_schema_version=1\n", stderr.String())
	})

	require.False(t, stub.called)
	require.Contains(t, stdout, "[dry-run]")
	require.Contains(t, stdout, "v2.0.0")
}

func TestRunRequiresWebhook(t *testing.T) {
	stderr := &bytes.Buffer{}
	exitCode := run(context.Background(), env(map[string]string{
		"SEMREL_VERSION": "v1.0.0",
	}), stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "SEMREL_PLUGIN_WEBHOOK_URL is required")
}

func TestRunRequiresVersion(t *testing.T) {
	stderr := &bytes.Buffer{}
	exitCode := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://example.test/webhook",
	}), stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "SEMREL_VERSION, SEMREL_TAG_NAME, or SEMREL_NEXT_VERSION is required")
}

func TestRunNotifierError(t *testing.T) {
	stub := &stubNotifier{err: errors.New("boom")}
	original := newNotifier
	newNotifier = func(cfg plugin.TeamsConfig) notifier { return stub }
	defer func() { newNotifier = original }()

	stderr := &bytes.Buffer{}
	exitCode := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://example.test/webhook",
		"SEMREL_NEXT_VERSION":       "v3.0.0",
	}), stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "boom")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = original }()

	fn()

	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return string(data)
}

func env(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}
