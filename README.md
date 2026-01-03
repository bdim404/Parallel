# Parallel

SOCKS5 parallel racing aggregator

## Usage

### Config file mode

Create a config file (default: `config.json`):

```json
{
  "listeners": [
    {
      "listen": "127.0.0.1:1080",
      "socks": ["upstream1:1081", "upstream2:1082"]
    }
  ]
}
```

Run with default config file:
```bash
parallel
```

Run with custom config file:
```bash
parallel --config /path/to/config.json
parallel -c /path/to/config.json
```

### Command line mode

```bash
parallel --listen-address 127.0.0.1 --listen-port 1080 --socks upstream1:1081 --socks upstream2:1082
```

## Command Line Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--config` | `-c` | `config.json` | Path to config file |
| `--listen-address` | | `127.0.0.1` | Listen address for SOCKS5 server |
| `--listen-port` | | | Listen port (required for command line mode) |
| `--socks` | | | Upstream SOCKS5 proxy (can be specified multiple times) |
| `--help` | `-h` | | Show help message |

### Notes

- Config file mode: If `--listen-port` is not specified, the program will load configuration from the config file
- Command line mode: If `--listen-port` is specified, at least one `--socks` upstream must be provided
- The `--socks` option can be used multiple times to specify multiple upstream proxies

## Build

```bash
go build -o parallel
```

## Testing

Test the SOCKS5 proxy with curl:

```bash
curl --socks5 127.0.0.1:1080 http://ipinfo.io
```

Check your IP address:

```bash
curl --socks5 127.0.0.1:1080 https://api.ipify.org
```

Test connection speed:

```bash
curl --socks5 127.0.0.1:1080 -w "\nTime: %{time_total}s\n" -o /dev/null -s http://www.google.com
```

## Example Output

```
2026/01/03 17:25:36 listening on 127.0.0.1:1080 with 6 upstreams
2026/01/03 17:26:39 → new connection from 127.0.0.1:58992
2026/01/03 17:26:39 request from 127.0.0.1:58992 to 104.26.13.205:443
2026/01/03 17:26:39 racing 6 upstreams for 104.26.13.205:443
2026/01/03 17:26:39 ✓ winner: upstream1:1080 (114ms) - received 1/6 responses
2026/01/03 17:26:39   closed: upstream4:1083 (233ms) - slower than winner
2026/01/03 17:26:40   closed: upstream6:1085 (367ms) - slower than winner
2026/01/03 17:26:40   closed: upstream3:1082 (468ms) - slower than winner
2026/01/03 17:26:40   closed: upstream5:1084 (585ms) - slower than winner
2026/01/03 17:26:40   closed: upstream2:1081 (667ms) - slower than winner
2026/01/03 17:26:40 race completed for 104.26.13.205:443: winner=upstream1:1080, duration=114ms, total=670ms
2026/01/03 17:26:40 relaying data for 127.0.0.1:58992 -> 104.26.13.205:443
2026/01/03 17:26:40 ← connection closed from 127.0.0.1:58992
```
