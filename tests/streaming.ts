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

app.get("/json", async (c) => {
  await new Promise((resolve) =>
    setTimeout(resolve, (Math.floor(Math.random() * 5000) + 1) * Math.random())
  );

  return c.json({
    message: "ZZZzzz",
  });
});

app.get("/random", async (c) => {
  await new Promise((resolve) =>
    setTimeout(resolve, (Math.floor(Math.random() * 5000) + 1) * Math.random())
  );

  return c.text("ZZZzzz1");
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

// Generate a very large JSON response for performance testing
app.get("/large-json", (c) => {
  const count = parseInt(c.req.query("count") || "100000");

  const generateRandomString = (length: number): string => {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  };

  const generateRandomObject = (id: number) => ({
    id,
    uuid: crypto.randomUUID(),
    name: generateRandomString(20),
    email: `${generateRandomString(10)}@${generateRandomString(8)}.com`,
    age: Math.floor(Math.random() * 80) + 18,
    salary: Math.floor(Math.random() * 200000) + 30000,
    active: Math.random() > 0.5,
    createdAt: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toISOString(),
    tags: Array.from({ length: Math.floor(Math.random() * 10) + 1 }, () => generateRandomString(8)),
    metadata: {
      department: generateRandomString(15),
      location: generateRandomString(12),
      manager: generateRandomString(18),
      projects: Array.from({ length: Math.floor(Math.random() * 5) + 1 }, (_, i) => ({
        id: `proj-${i}`,
        name: generateRandomString(25),
        budget: Math.floor(Math.random() * 1000000),
        status: ["active", "pending", "completed", "cancelled"][Math.floor(Math.random() * 4)],
      })),
    },
    description: generateRandomString(200),
    notes: Array.from({ length: Math.floor(Math.random() * 5) }, () => generateRandomString(100)),
  });

  const data = {
    total: count,
    generatedAt: new Date().toISOString(),
    records: Array.from({ length: count }, (_, i) => generateRandomObject(i + 1)),
  };

  return c.json(data);
});

Deno.serve(app.fetch);
