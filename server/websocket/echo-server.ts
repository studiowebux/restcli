#!/usr/bin/env -S deno run --allow-net

/**
 * WebSocket Echo Server for REST CLI Testing
 *
 * A simple WebSocket server that echoes back any messages it receives.
 * Perfect for testing WebSocket client functionality without external dependencies.
 *
 * Usage:
 *   deno run --allow-net echo-server.ts [port]
 *
 * Default port: 8080
 *
 * Features:
 * - Echoes back all text/JSON messages
 * - Handles ping/pong automatically
 * - Logs all connection events
 * - Graceful shutdown on SIGINT
 */

const DEFAULT_PORT = 8080;
const port = parseInt(Deno.args[0] || String(DEFAULT_PORT));

let connectionCount = 0;
const connections = new Set<WebSocket>();

function formatTimestamp(): string {
  return new Date().toISOString().replace("T", " ").substring(0, 19);
}

function log(message: string, ...args: unknown[]): void {
  console.log(`[${formatTimestamp()}]`, message, ...args);
}

async function handleConnection(ws: WebSocket): Promise<void> {
  const connId = ++connectionCount;
  connections.add(ws);

  log(`🔌 Connection #${connId} opened (total: ${connections.size})`);

  ws.addEventListener("message", (event) => {
    const message = event.data;

    if (typeof message === "string") {
      log(
        `📨 #${connId} received: ${message.substring(0, 100)}${
          message.length > 100 ? "..." : ""
        }`,
      );

      if (message === "hello") {
        ws.send("Bonjour !");
        return;
      }

      // Echo back the message
      try {
        ws.send(message);
        log(`📤 #${connId} echoed back`);
      } catch (error) {
        log(`❌ #${connId} failed to echo:`, error);
      }
    } else {
      log(`📨 #${connId} received binary data (${message.byteLength} bytes)`);
      // Echo back binary data
      try {
        ws.send(message);
        log(`📤 #${connId} echoed back binary`);
      } catch (error) {
        log(`❌ #${connId} failed to echo binary:`, error);
      }
    }
  });

  ws.addEventListener("close", (event) => {
    connections.delete(ws);
    log(
      `🔌 Connection #${connId} closed (code: ${event.code}, total: ${connections.size})`,
    );
  });

  ws.addEventListener("error", (event) => {
    log(`❌ #${connId} error:`, event);
  });

  // Keep connection alive by responding to pings
  ws.addEventListener("ping", () => {
    log(`🏓 #${connId} ping`);
  });
}

async function handler(req: Request): Promise<Response> {
  // Only handle WebSocket upgrade requests
  if (req.headers.get("upgrade") !== "websocket") {
    return new Response(
      "WebSocket Echo Server\n\nConnect using ws://localhost:" + port,
      {
        status: 200,
        headers: { "content-type": "text/plain" },
      },
    );
  }

  const { socket, response } = Deno.upgradeWebSocket(req);
  handleConnection(socket);

  return response;
}

// Start server
log(`🚀 WebSocket Echo Server starting on ws://localhost:${port}`);
log(`   To test: wscat -c ws://localhost:${port}`);
log(`   Press Ctrl+C to stop`);

const ac = new AbortController();

// Graceful shutdown
Deno.addSignalListener("SIGINT", () => {
  log(`\n🛑 Shutting down... (${connections.size} active connections)`);

  // Close all active connections
  for (const ws of connections) {
    try {
      ws.close(1001, "Server shutting down");
    } catch (error) {
      // Ignore errors during shutdown
    }
  }

  ac.abort();
  log(`✅ Server stopped`);
  Deno.exit(0);
});

// Start serving
try {
  await Deno.serve({
    port,
    signal: ac.signal,
    onListen: ({ port, hostname }) => {
      log(`✅ Server listening on ws://${hostname}:${port}`);
    },
  }, handler).finished;
} catch (error) {
  if (error.name === "AddrInUse") {
    log(`❌ Port ${port} is already in use`);
    Deno.exit(1);
  }
  throw error;
}
