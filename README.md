# SockRacer

SOCKS5 parallel racing aggregator

## Usage

### Config file mode

Create a config file (default: `config.json`):

```json
{
  "listeners": [
    {
      "listen": "127.0.0.1:1080",
      "socks": [
        {
          "name": "US-West-1",
          "address": "upstream1:1081"
        },
        {
          "name": "US-East-1",
          "address": "upstream2:1082"
        }
      ]
    }
  ]
}
```

Run with default config file:
```bash
sockracer
```

Run with custom config file:
```bash
sockracer --config /path/to/config.json
sockracer -c /path/to/config.json
```

### Command line mode

```bash
sockracer --listen-address 127.0.0.1 --listen-port 1080 --socks upstream1:1081 --socks upstream2:1082
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
cd src && go build -o sockracer
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
2026/01/04 00:58:10 listening on 127.0.0.1:9080 with 3 upstreams
2026/01/04 00:58:18 accepted connection from 127.0.0.1:59075
2026/01/04 00:58:18 socks5: sending connect request (20 bytes) to www.baidu.com:443
2026/01/04 00:58:18 socks5: sending connect request (20 bytes) to www.baidu.com:443
2026/01/04 00:58:18 socks5: sending connect request (20 bytes) to www.baidu.com:443
2026/01/04 00:58:19 socks5: received reply: ver=5, rep=0, rsv=0, atyp=1
2026/01/04 00:58:19 socks5: read bind addr: 0.0.0.0:0
2026/01/04 00:58:19 socks5: connect handshake completed successfully
2026/01/04 00:58:19 ✓ www.baidu.com:443 -> upstream3 (198.18.169.1:1082) (678ms)
2026/01/04 00:58:19 socks5: received reply: ver=5, rep=0, rsv=0, atyp=1
2026/01/04 00:58:19 socks5: read bind addr: 0.0.0.0:0
2026/01/04 00:58:19 socks5: connect handshake completed successfully
2026/01/04 00:58:19 socks5: received reply: ver=5, rep=0, rsv=0, atyp=1
2026/01/04 00:58:19 socks5: read bind addr: 0.0.0.0:0
2026/01/04 00:58:19 socks5: connect handshake completed successfully
```
