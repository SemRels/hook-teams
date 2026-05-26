// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The hook-teams Authors

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	plugin "github.com/SemRels/hook-teams/internal/plugin"
)

type notifier interface {
	Notify(context.Context, string, string) error
}

var newNotifier = func(cfg plugin.TeamsConfig) notifier {
	return plugin.NewTeamsNotifier(cfg)
}

func run(ctx context.Context, getenv func(string) string, stderr io.Writer) int {
	webhookURL := getenv("SEMREL_PLUGIN_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Fprintln(stderr, "hook-teams: SEMREL_PLUGIN_WEBHOOK_URL is required")
		return 1
	}
	version := firstNonEmpty(getenv("SEMREL_VERSION"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_NEXT_VERSION"))
	if version == "" {
		fmt.Fprintln(stderr, "hook-teams: SEMREL_VERSION, SEMREL_TAG_NAME, or SEMREL_NEXT_VERSION is required")
		return 1
	}
	if getenv("SEMREL_DRY_RUN") == "true" {
		fmt.Printf("hook-teams: [dry-run] would send Teams notification for %s\n", version)
		return 0
	}

	cfg := plugin.TeamsConfig{
		WebhookURL: webhookURL,
		Title:      getenv("SEMREL_PLUGIN_TITLE"),
		ThemeColor: getenv("SEMREL_PLUGIN_THEME_COLOR"),
		Mention:    getenv("SEMREL_PLUGIN_MENTION"),
	}

	if err := newNotifier(cfg).Notify(ctx, version, getenv("SEMREL_CHANGELOG")); err != nil {
		fmt.Fprintln(stderr, "hook-teams:", err)
		return 1
	}
	return 0
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	os.Exit(run(ctx, os.Getenv, os.Stderr))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
