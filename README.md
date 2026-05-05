# cronwatch

Lightweight daemon that monitors cron job execution and sends alerts on missed or failed runs.

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && make build
```

## Usage

Define your monitored jobs in a config file:

```yaml
# cronwatch.yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert:
      slack: "#alerts"
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap your cron commands to report execution status:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup /usr/local/bin/backup.sh
```

cronwatch will send alerts if a job exceeds its timeout, exits with a non-zero status, or fails to run within the expected schedule window.

## Configuration

| Field | Description |
|-------|-------------|
| `schedule` | Standard cron expression |
| `timeout` | Maximum allowed run duration |
| `alert` | Notification target (email, Slack, webhook) |

## License

MIT