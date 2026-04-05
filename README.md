# Stockyard Checkin

**Self-hosted member check-in and attendance tracking**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/checkin/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9807:9807 -v checkin_data:/data ghcr.io/stockyard-dev/stockyard-checkin
```

Open `http://localhost:9807` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9807` | HTTP port |
| `DATA_DIR` | `./checkin-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
