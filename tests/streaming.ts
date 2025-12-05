// deno run -A streaming.ts
import { logger } from "npm:hono/logger";
import { stream, streamSSE, streamText } from "npm:hono/streaming";
import { createRoute, OpenAPIHono } from "npm:@hono/zod-openapi";
import { z } from "npm:zod";

export const app = new OpenAPIHono();

// ../bin/restcli openapi2http -f json -o ./tmp http://localhost:8000/doc
app.openapi(
  createRoute({
    method: "get",
    path: "/health",
    tags: ["Admin"],
    request: {},
    responses: {
      200: {
        content: {
          "application/json": {
            schema: z
              .object({
                status: z.string(),
              }).openapi("ServerHealthStatus"),
          },
        },
        description: "Server Health Status",
      },
    },
  }),
  async (c) => {
    return c.json({
      status: "Ok",
    });
  },
);
app.doc("/doc", (c) => ({
  openapi: "3.0.0",
  info: {
    version: "1.0.0",
    title: "Rest CLI Server",
  },
  servers: [
    {
      url: new URL(c.req.url).origin,
      description: "Current environment",
    },
  ],
}));

app.use(logger());

app.get("/", (c) => c.text("Hello Deno!"));

app.get("/slow", async (c) => {
  await new Promise((resolve) => setTimeout(resolve, 5000));
  return c.text("ZZZzzz");
});

app.get("/random", async (c) => {
  await new Promise((resolve) =>
    setTimeout(resolve, (Math.floor(Math.random() * 5000) + 1) * Math.random())
  );
  return c.text("ZZZzzz");
});

let id = 0;

app.get("/sse", async (c) => {
  return streamSSE(c, async (stream) => {
    while (true) {
      const message = `It is ${new Date().toISOString()}`;
      await stream.writeSSE({
        data: message,
        event: "time-update",
        id: String(id++),
      });
      await stream.sleep(1000);
    }
  });
});

app.get("/stream-text", (c) => {
  return streamText(c, async (stream) => {
    // Write a text with a new line ('\n').
    await stream.writeln("Hello");
    // Wait 1 second.
    await stream.sleep(1000);
    // Write a text without a new line.
    await stream.write(`Hono!`);
  });
});

app.get("/stream", (c) => {
  return stream(c, async (stream) => {
    // Write a process to be executed when aborted.
    stream.onAbort(() => {
      console.log("Aborted!");
    });
    // Write a Uint8Array.
    await stream.write(new Uint8Array([0x48, 0x65, 0x6c, 0x6c, 0x6f]));
  });
});

Deno.serve(app.fetch);
