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
6. **Categories and filtering** - organize requests with categories, filter by category in TUI

### Execution & Control
7. **Request chaining** with dependency resolution and automatic variable extraction
8. **Streaming support** for SSE and real-time responses
9. **GraphQL & HTTP protocols** with automatic detection
10. **Request cancellation** (ESC to abort in-progress requests)
11. **Confirmation modals** for critical endpoints
12. **Concurrent request blocking** prevents accidental request

### Security & Auth
13. **OAuth 2.0** with PKCE flow and token auto-extraction
14. **mTLS support** with client certificates
15. **Variable interpolation in TLS paths** for dynamic cert loading

### Analysis & Debugging
16. **Inline filter editor** with bookmark system - filter responses while viewing JSON structure
17. **Response filtering** with JMESPath or bash commands
18. **Response pinning and diff** for regression testing
19. **Error detail modals** with full stack traces
20. **Embedded documentation** viewer with collapsible trees
21. **Analytics tracking** with per-endpoint stats and aggregated metrics
22. **Request history** with persistent storage, search, and timestamp tracking
23. **HTTP method color coding** in file list for quick identification

### Performance Testing
24. **Stress testing** with configurable concurrency and load
25. **Ramp-up control** for gradual load increase
26. **Real-time metrics** (latency, RPS, percentiles P50/P95/P99)
27. **Test result persistence** with historical comparison
28. **One-click re-run** for saved test configurations

### Development Tools
29. **Mock server** with YAML-based endpoint definitions
30. **Debug proxy** for inspecting HTTP traffic (localhost only, HTTP-only)
31. **HAR file importer** (convert browser recordings to request files)

### Automation
32. **CLI mode** for scripting (JSON/YAML output)
33. **cURL converter** (convert cURL to request files)
34. **OpenAPI converter** (generate requests from specs)

## License

See [LICENSE](LICENSE)
