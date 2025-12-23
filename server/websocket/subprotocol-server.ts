// WebSocket Subprotocol Server Example
// Demonstrates subprotocol negotiation and protocol-specific handling
// Run with: deno run --allow-net subprotocol-server.ts

const PORT = 8082;
const SUPPORTED_PROTOCOLS = ["chat", "json-rpc"];

console.log(`WebSocket Subprotocol Server starting on ws://localhost:${PORT}`);
console.log(`Supported subprotocols: ${SUPPORTED_PROTOCOLS.join(", ")}`);

Deno.serve({ port: PORT }, (req) => {
  // Check for WebSocket upgrade request
  if (req.headers.get("upgrade") !== "websocket") {
    return new Response("Expected WebSocket upgrade", { status: 400 });
  }

  // Check requested subprotocols
  const requestedProtocols = req.headers.get("sec-websocket-protocol");
  let selectedProtocol = "";

  if (requestedProtocols) {
    const protocols = requestedProtocols.split(",").map((p) => p.trim());
    // Find first supported protocol
    selectedProtocol = protocols.find((p) => SUPPORTED_PROTOCOLS.includes(p)) || "";
  }

  if (!selectedProtocol) {
    console.log(`❌ No compatible subprotocol found. Requested: ${requestedProtocols}`);
    console.log(`   Supported: ${SUPPORTED_PROTOCOLS.join(", ")}`);
  } else {
    console.log(`✅ Negotiated subprotocol: ${selectedProtocol}`);
  }

  // Upgrade to WebSocket with selected protocol
  const { socket, response } = Deno.upgradeWebSocket(req, {
    protocol: selectedProtocol || undefined,
  });

  socket.onopen = () => {
    console.log(`🔌 Client connected using protocol: ${selectedProtocol || "none"}`);

    // Send protocol-specific welcome message
    if (selectedProtocol === "chat") {
      socket.send(JSON.stringify({
        type: "system",
        message: "Welcome to chat protocol!",
        protocol: "chat",
        timestamp: new Date().toISOString(),
      }));
    } else if (selectedProtocol === "json-rpc") {
      socket.send(JSON.stringify({
        jsonrpc: "2.0",
        method: "server.ready",
        params: {
          message: "JSON-RPC server ready",
          supportedMethods: ["echo", "ping", "time"],
        },
      }));
    } else {
      socket.send("Connected without subprotocol");
    }
  };

  socket.onmessage = (event) => {
    console.log(`📨 [${selectedProtocol || "none"}] Received:`, event.data);

    if (selectedProtocol === "chat") {
      handleChatProtocol(socket, event.data);
    } else if (selectedProtocol === "json-rpc") {
      handleJsonRpcProtocol(socket, event.data);
    } else {
      // Default echo
      socket.send(`[Echo] ${event.data}`);
    }
  };

  socket.onerror = (error) => {
    console.error("❌ WebSocket error:", error);
  };

  socket.onclose = () => {
    console.log(`🔌 Client disconnected [${selectedProtocol || "none"}]`);
  };

  return response;
});

function handleChatProtocol(socket: WebSocket, data: string) {
  try {
    const message = JSON.parse(data);

    if (message.type === "message") {
      // Echo chat message
      socket.send(JSON.stringify({
        type: "message",
        from: "server",
        text: `Echo: ${message.text}`,
        timestamp: new Date().toISOString(),
      }));
    } else if (message.type === "ping") {
      socket.send(JSON.stringify({
        type: "pong",
        timestamp: new Date().toISOString(),
      }));
    }
  } catch {
    socket.send(JSON.stringify({
      type: "error",
      message: "Invalid chat message format",
    }));
  }
}

function handleJsonRpcProtocol(socket: WebSocket, data: string) {
  try {
    const request = JSON.parse(data);

    if (!request.jsonrpc || request.jsonrpc !== "2.0") {
      socket.send(JSON.stringify({
        jsonrpc: "2.0",
        error: {
          code: -32600,
          message: "Invalid JSON-RPC version",
        },
        id: request.id || null,
      }));
      return;
    }

    // Handle different methods
    let result: unknown;
    switch (request.method) {
      case "echo":
        result = { echo: request.params };
        break;
      case "ping":
        result = { pong: new Date().toISOString() };
        break;
      case "time":
        result = { time: new Date().toISOString() };
        break;
      default:
        socket.send(JSON.stringify({
          jsonrpc: "2.0",
          error: {
            code: -32601,
            message: `Method not found: ${request.method}`,
          },
          id: request.id || null,
        }));
        return;
    }

    socket.send(JSON.stringify({
      jsonrpc: "2.0",
      result,
      id: request.id || null,
    }));
  } catch {
    socket.send(JSON.stringify({
      jsonrpc: "2.0",
      error: {
        code: -32700,
        message: "Parse error",
      },
      id: null,
    }));
  }
}
