# hook-teams

A [semrel](https://semrel.dev) plugin that sends release notifications to Microsoft Teams via Incoming Webhooks.

## Usage

```yaml
plugins:
  - uses: hook-teams
    args:
      webhook_url: "https://your-tenant.webhook.office.com/webhookb2/..."
      title: "🚀 New Release"        # optional
      theme_color: "0078D7"           # optional, hex without #
      mention: "user@example.com"     # optional, Teams user to @mention
      max_retries: "3"                # optional
      retry_delay: "2s"               # optional
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SEMREL_PLUGIN_WEBHOOK_URL` | ✅ | Teams Incoming Webhook URL |
| `SEMREL_PLUGIN_TITLE` | ❌ | Notification title (default: "🚀 New Release") |
| `SEMREL_PLUGIN_THEME_COLOR` | ❌ | Card theme color hex (default: `0078D7`) |
| `SEMREL_PLUGIN_MENTION` | ❌ | Teams user email to @mention |
| `SEMREL_PLUGIN_MAX_RETRIES` | ❌ | Retries on transient network failures and HTTP 5xx responses (default: `3`) |
| `SEMREL_PLUGIN_RETRY_DELAY` | ❌ | Delay between retry attempts (default: `2s`) |

## Retry behavior

The plugin retries transient failures caused by network errors or HTTP `5xx` responses. HTTP `2xx` and `4xx` responses are not retried. Each retry attempt is logged to standard error.

## Getting a Webhook URL

1. In Teams, go to the channel → **...** → **Connectors**
2. Search for **Incoming Webhook** → Configure
3. Copy the webhook URL

## License

Apache-2.0
