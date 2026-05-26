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
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SEMREL_PLUGIN_WEBHOOK_URL` | ✅ | Teams Incoming Webhook URL |
| `SEMREL_PLUGIN_TITLE` | ❌ | Notification title (default: "🚀 New Release") |
| `SEMREL_PLUGIN_THEME_COLOR` | ❌ | Card theme color hex (default: `0078D7`) |
| `SEMREL_PLUGIN_MENTION` | ❌ | Teams user email to @mention |

## Getting a Webhook URL

1. In Teams, go to the channel → **...** → **Connectors**
2. Search for **Incoming Webhook** → Configure
3. Copy the webhook URL

## License

Apache-2.0