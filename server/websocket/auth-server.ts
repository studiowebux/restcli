// WebSocket Authentication Server Example
// Demonstrates Bearer token authentication
// Run with: deno run --allow-net auth-server.ts

const PORT = 8081;
const VALID_TOKEN = "demo-token-12345";

console.log(`WebSocket Auth Server starting on ws://localhost:${PORT}`);
console.log(`Valid token: ${VALID_TOKEN}`);

Deno.serve({ port: PORT }, (req) => {
  // Check for WebSocket upgrade request
  if (req.headers.get("upgrade") !== "websocket") {
    return new Response("Expected WebSocket upgrade", { status: 400 });
  }

  // Verify Authorization header
  const authHeader = req.headers.get("authorization");
  if (!authHeader) {
    console.log("❌ Connection rejected: Missing Authorization header");
    return new Response("Missing Authorization header", { status: 401 });
  }

  if (!authHeader.startsWith("Bearer ")) {
    console.log("❌ Connection rejected: Invalid auth format:", authHeader);
    return new Response("Invalid Authorization format. Expected: Bearer <token>", {
      status: 401,
    });
  }

  const token = authHeader.substring(7); // Remove "Bearer " prefix
  if (token !== VALID_TOKEN) {
    console.log("❌ Connection rejected: Invalid token:", token);
    return new Response("Invalid token", { status: 403 });
  }

  console.log("✅ Connection accepted with valid token");

  // Upgrade to WebSocket
  const { socket, response } = Deno.upgradeWebSocket(req);

  socket.onopen = () => {
    console.log("🔌 Client connected (authenticated)");
    socket.send(JSON.stringify({
      type: "auth_success",
      message: "Authentication successful",
      timestamp: new Date().toISOString(),
    }));
  };

  socket.onmessage = (event) => {
    console.log("📨 Received:", event.data);

    try {
      const message = JSON.parse(event.data);
      // Echo back with authentication info
      socket.send(JSON.stringify({
        type: "echo",
        original: message,
        authenticated: true,
        timestamp: new Date().toISOString(),
      }));
    } catch {
      // Not JSON, echo as text
      socket.send(`[Authenticated Echo] ${event.data}`);
    }
  };

  socket.onerror = (error) => {
    console.error("❌ WebSocket error:", error);
  };

  socket.onclose = () => {
    console.log("🔌 Client disconnected");
  };

  return response;
});
