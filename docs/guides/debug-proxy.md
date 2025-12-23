---
title: Debug Proxy
tags:
  - guide
---

# Debug Proxy

HTTP debugging proxy for local development and troubleshooting.

## Overview

Debug proxy captures and inspects HTTP traffic between your application and servers. Lightweight alternative to tools like Burp Suite for local development.

HTTP-only inspection. Designed for local troubleshooting.

## CLI Usage

### Start Proxy

Default port (8888):
```bash
restcli proxy start
```

Custom port:
```bash
restcli proxy start --proxy-port 3128
```

Proxy starts with TUI interface. Press `y` to view captured traffic.

### Configure Application

Set environment variables to route HTTP traffic through proxy:

```bash
export HTTP_PROXY=http://localhost:8888
export http_proxy=http://localhost:8888
```

Then run your application. All HTTP requests will be captured.

### Test with curl

```bash
export HTTP_PROXY=http://localhost:8888
curl http://api.example.com/users
```

Traffic appears in TUI proxy viewer.

## TUI Proxy Management

Press `y` in TUI to open proxy management modal.

### Starting Proxy from TUI

When proxy is stopped:
- Shows configured port (from profile or default 8888)
- Displays environment configuration commands
- Press `s` to start proxy
- Press `ESC` to close

### Traffic Log Display

When proxy is running, displays all captured requests (up to 1000):

```
#ID   METHOD URL                     → STATUS SIZE     DURATION
#1    GET    /api/users              → 200    1.2 KB   45ms
#2    POST   /api/auth/login         → 401    256 B    12ms
#3    GET    /health                 → 200    64 B     5ms
```

**Color Coding**:
- Methods: GET (green), POST (blue), PUT (yellow), DELETE (red)
- Status: 2xx (green), 3xx (yellow), 4xx/5xx (red)

### Request Details

Traffic log shows one line per request with method, URL, status, size, and duration.

Press `Enter` on selected request to open detailed view modal showing:
- Complete request headers
- Full request body (JSON auto-formatted, binary content detection)
- Complete response headers
- Full response body (JSON auto-formatted, binary content detection)
- Binary content (images, PDFs, etc.) displays size and content-type instead of raw bytes
- Scrollable with `j`/`k`, `Ctrl+d`/`u`, `g`/`G` navigation

### Navigation

Traffic list:

| Key         | Action                      |
| ----------- | --------------------------- |
| `s`         | Start/stop proxy            |
| `j`/`k`     | Navigate up/down in list    |
| `Ctrl+d/u`  | Half page down/up           |
| `g`         | Jump to top                 |
| `G`         | Jump to bottom              |
| `PgUp/Dn`   | Page up/down                |
| `Enter`     | Open request detail modal   |
| `c`         | Clear captured logs         |
| `Esc`       | Close viewer                |

Detail modal:

| Key         | Action                      |
| ----------- | --------------------------- |
| `j`/`k`     | Scroll line by line         |
| `Ctrl+d/u`  | Half page down/up           |
| `g`         | Jump to top                 |
| `G`         | Jump to bottom              |
| `PgUp/Dn`   | Page up/down                |
| `Esc`/`q`   | Return to traffic list      |

### Real-Time Updates

Proxy viewer uses **event-based updates** - new requests appear instantly without polling overhead.

## Profile Configuration

Configure proxy port in your profile (`~/.config/restcli/profiles.json`):

```json
{
  "profiles": [
    {
      "name": "default",
      "proxyPort": 9000
    }
  ]
}
```

Without configuration, proxy uses port 8888 by default.

## Use Cases

### Debug API Calls

Inspect exact requests your application sends:

```bash
# Start proxy
restcli proxy start

# In another terminal, configure and run app
export HTTP_PROXY=http://localhost:8888
npm run dev

# Press y in proxy TUI to see all requests
```

### Compare Request/Response

Verify request format and inspect response:

1. Start proxy: `restcli proxy start`
2. Run application with proxy configured
3. Press `y` to view traffic
4. Navigate to request with `j`/`k` or `Ctrl+d`/`u`
5. Press `Enter` to view full request/response details

### Troubleshoot Integration

Debug third-party API integration:

```bash
# Start proxy
restcli proxy start --proxy-port 9000

# Configure application
export HTTP_PROXY=http://localhost:9000
python integration_test.py

# View captured traffic in TUI
```

### Monitor Background Jobs

Track HTTP requests from scheduled tasks:

```bash
# Start proxy
restcli proxy start

# Run cron job with proxy
HTTP_PROXY=http://localhost:8888 ./sync_job.sh

# Check captured requests
```

## Limitations

**HTTP-only** - Captures and inspects HTTP traffic only.

**Memory-Limited** - Stores last 1000 requests. Older requests are discarded.

**Read-only** - Captures traffic without modification.

## Workflow Integration

### With Mock Server

Use proxy to verify requests sent to mock server:

```bash
# Terminal 1: Start mock server
restcli mock start

# Terminal 2: Start proxy
restcli proxy start --proxy-port 8888

# Terminal 3: Send requests through proxy
export HTTP_PROXY=http://localhost:8888
curl http://localhost:8080/api/users

# View in proxy TUI (y)
```

### With Request Files

Test `.http` files and capture traffic:

```bash
# Start proxy in one terminal
restcli proxy start

# In another terminal, execute request through proxy
HTTP_PROXY=http://localhost:8888 restcli exec test.http

# Switch back to proxy TUI and press 'y' to view captured traffic
```

### Development Workflow

1. Start proxy: `restcli proxy start`
2. Configure environment: `export HTTP_PROXY=http://localhost:8888`
3. Run application
4. Monitor traffic: Press `y`
5. Clear logs as needed: Press `c`
6. Stop proxy: `Ctrl+C`

## Browser Configuration

**Firefox Localhost Bypass**

Firefox blocks proxying of localhost connections by default. For local development testing:

Option 1: Use hostname alias
```bash
# Add to /etc/hosts
sudo sh -c 'echo "127.0.0.1 local.dev" >> /etc/hosts'

# Access via local.dev instead of localhost
curl -x http://localhost:8888 http://local.dev:8080/api
```

Option 2: Enable localhost proxying in Firefox
1. Navigate to `about:config`
2. Search for `network.proxy.allow_hijacking_localhost`
3. Set to `true`

**Firefox Proxy Settings**

1. Settings → Network Settings → Manual proxy configuration
2. Set **HTTP Proxy**: `localhost`, Port: `8888`
3. Leave **SOCKS Host** empty (SOCKS overrides HTTP proxy)

## Tips

**Avoid Port Conflicts**

Use custom port if 8888 is taken:
```bash
restcli proxy start --proxy-port 9000
```

**Clear Logs Regularly**

Press `c` in proxy viewer to clear captured logs and reduce memory usage.

**Check Environment Variables**

Verify proxy configuration:
```bash
echo $HTTP_PROXY
```

Should output: `http://localhost:8888`

**Unset After Testing**

Remove proxy configuration when done:
```bash
unset HTTP_PROXY
unset http_proxy
```

**Combine with History**

Use proxy for live capture, history for request replay:
1. Capture traffic with proxy
2. Save request details
3. Replay with history feature

**No-Proxy for Localhost**

Some applications need `NO_PROXY` set:
```bash
export NO_PROXY=localhost,127.0.0.1
```

Prevents localhost requests from routing through proxy.
