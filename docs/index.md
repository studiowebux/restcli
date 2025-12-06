---
title: Welcome to RestCLI
description: Keyboard-driven TUI for testing HTTP endpoints.
---

# REST CLI

Keyboard-driven TUI for testing HTTP endpoints.

## Why?

Manage API calls using directory and file structure instead of massive JSON files.

No more merge conflicts. No more outdated collections. Each endpoint is a file. Git-friendly. Team-friendly.

## Features

### Core
1. **File-based requests** (`.http`, `.yaml`, `.json`, `.jsonc`, `openapi`) - one endpoint per file
2. **Variable substitution** with `{{varName}}` and shell commands `$(cmd)`
3. **Multi-value variables** with aliases (e.g., `dev`, `staging`, `prod`)
4. **Profile system** for environment-specific headers and variables
5. **Request history** with split view and live preview

### Execution & Control
6. **Streaming support** for SSE and real-time responses
7. **GraphQL & HTTP protocols** with automatic detection
8. **Request cancellation** (ESC to abort in-progress requests)
9. **Confirmation modals** for critical endpoints
10. **Concurrent request blocking** prevents accidental request

### Security & Auth
11. **OAuth 2.0** with PKCE flow and token auto-extraction
12. **mTLS support** with client certificates
13. **Variable interpolation in TLS paths** for dynamic cert loading

### Analysis & Debugging
14. **Response filtering** with JMESPath or bash commands
15. **Response pinning and diff** for regression testing
16. **Error detail modals** with full stack traces
17. **Embedded documentation** viewer with collapsible trees
18. **Analytics tracking** with per-endpoint stats and aggregated metrics
19. **Request history** with persistent storage and search

### Performance Testing
20. **Stress testing** with configurable concurrency and load
21. **Ramp-up control** for gradual load increase
22. **Real-time metrics** (latency, RPS, percentiles P50/P95/P99)
23. **Test result persistence** with historical comparison
24. **One-click re-run** for saved test configurations

### Automation
25. **CLI mode** for scripting (JSON/YAML output)
26. **cURL converter** (convert cURL to request files)
27. **OpenAPI converter** (generate requests from specs)

## License

See [LICENSE](LICENSE)
