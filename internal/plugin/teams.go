// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The hook-teams Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type TeamsConfig struct {
	WebhookURL string
	Title      string
	ThemeColor string
	Mention    string
}

type TeamsNotifier struct {
	cfg    TeamsConfig
	client *http.Client
}

func NewTeamsNotifier(cfg TeamsConfig) *TeamsNotifier {
	if cfg.Title == "" {
		cfg.Title = "🚀 New Release"
	}
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = "0078D7"
	}
	cfg.ThemeColor = strings.TrimPrefix(cfg.ThemeColor, "#")

	return &TeamsNotifier{
		cfg:    cfg,
		client: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

type teamsPayload struct {
	Type       string         `json:"@type"`
	Context    string         `json:"@context"`
	ThemeColor string         `json:"themeColor"`
	Summary    string         `json:"summary"`
	Sections   []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle    string      `json:"activityTitle"`
	ActivitySubtitle string      `json:"activitySubtitle,omitempty"`
	Facts            []teamsFact `json:"facts,omitempty"`
	Text             string      `json:"text,omitempty"`
}

type teamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (n *TeamsNotifier) Notify(ctx context.Context, version, changelog string) error {
	if n.cfg.WebhookURL == "" {
		return fmt.Errorf("teams: webhook URL is required")
	}

	payload := n.buildPayload(version, changelog)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("teams: send notification: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (n *TeamsNotifier) buildPayload(version, changelog string) teamsPayload {
	return teamsPayload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: n.cfg.ThemeColor,
		Summary:    fmt.Sprintf("New release %s", version),
		Sections: []teamsSection{{
			ActivityTitle:    n.cfg.Title,
			ActivitySubtitle: version,
			Facts: []teamsFact{{
				Name:  "Version",
				Value: version,
			}},
			Text: buildText(n.cfg.Mention, changelog),
		}},
	}
}

func buildText(mention, changelog string) string {
	parts := make([]string, 0, 2)
	if mention = strings.TrimSpace(mention); mention != "" {
		if !strings.HasPrefix(mention, "@") {
			mention = "@" + mention
		}
		parts = append(parts, mention)
	}
	if changelog = strings.TrimSpace(changelog); changelog != "" {
		parts = append(parts, changelog)
	}
	return strings.Join(parts, "\n\n")
}
