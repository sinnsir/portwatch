# portwatch

Lightweight CLI daemon that monitors TCP port availability and triggers webhooks or shell commands when port state changes.

## Features

- Monitor one or more TCP ports on configurable intervals
- Trigger HTTP webhooks on `open` or `closed` state transitions
- Execute shell commands on state changes
- JSON payload sent to webhooks with host, port, state, and timestamp
- Graceful shutdown via SIGINT/SIGTERM

## Installation

```bash
go install github.com/yourorg/portwatch/cmd/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/portwatch.git
cd portwatch
make build
```

## Usage

```bash
portwatch --config config.yaml
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to configuration file |
| `--version` | | Print version and exit |

## Configuration

See [`internal/config/example.yaml`](internal/config/example.yaml) for a full example.

```yaml
ports:
  - host: localhost
    port: 8080
    interval: 10s
    actions:
      - on: open
        webhook: https://hooks.example.com/notify
      - on: closed
        command: "echo 'port 8080 is down'"
```

### Configuration Reference

#### `ports[]`

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `host` | string | No | `localhost` | Hostname or IP to monitor |
| `port` | int | Yes | — | TCP port number (1–65535) |
| `interval` | duration | No | `30s` | How often to check the port |
| `actions` | list | No | — | Actions to trigger on state change |

#### `ports[].actions[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `on` | string | Yes | State that triggers this action: `open` or `closed` |
| `webhook` | string | No* | HTTP URL to POST a JSON payload to |
| `command` | string | No* | Shell command to execute |

\* At least one of `webhook` or `command` must be set per action.

### Webhook Payload

When a webhook is triggered, portwatch sends a POST request with `Content-Type: application/json`:

```json
{
  "host": "localhost",
  "port": 8080,
  "state": "open",
  "previous_state": "closed",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Development

### Prerequisites

- Go 1.22+

### Commands

```bash
make build      # Build binary to ./bin/portwatch
make test       # Run all tests
make lint       # Run golangci-lint
make clean      # Remove build artifacts
```

### Project Structure

```
portwatch/
├── cmd/portwatch/        # CLI entry point
├── internal/
│   ├── config/           # YAML config loading and validation
│   ├── monitor/          # Port polling loop and state tracking
│   ├── notifier/         # Dispatches actions on state change
│   ├── portcheck/        # Low-level TCP dial check
│   └── runner/           # Executes webhooks and shell commands
└── Makefile
```

## License

MIT
