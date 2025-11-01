import * as path from "@std/path";
import { exists } from "@std/fs";

export interface HistoryEntry {
  timestamp: string;
  requestFile: string;
  requestName?: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  body?: string;
  responseStatus: number;
  responseStatusText: string;
  responseHeaders: Record<string, string>;
  responseBody: string;
  duration: number;
  error?: string;
}

export class HistoryManager {
  private historyDir: string;

  constructor(baseDir?: string) {
    // If baseDir not provided, use current directory for backward compatibility
    const dir = baseDir ?? ".";
    this.historyDir = path.join(dir, "history");
  }

  /**
   * Initialize history directory
   */
  async init(): Promise<void> {
    if (!await exists(this.historyDir)) {
      await Deno.mkdir(this.historyDir, { recursive: true });
    }
  }

  /**
   * Generate timestamp for filename (ISO 8601 format, safe for filenames)
   */
  private generateTimestamp(): string {
    const now = new Date();
    return now.toISOString()
      .replace(/:/g, "-")  // Replace colons for filename safety
      .replace(/\./g, "-") // Replace dots for filename safety
      .slice(0, -1);       // Remove trailing 'Z'
  }

  /**
   * Save a history entry to file
   */
  async save(entry: HistoryEntry): Promise<string> {
    await this.init();

    const filenameSafeTimestamp = this.generateTimestamp();

    // Extract base filename from request file path
    const requestBaseName = path.basename(entry.requestFile, ".http");
    const historyFileName = `${requestBaseName}_${filenameSafeTimestamp}.json`;
    const historyFilePath = path.join(this.historyDir, historyFileName);

    // Save entry with original ISO timestamp (keep it parseable)
    await Deno.writeTextFile(
      historyFilePath,
      JSON.stringify(entry, null, 2)
    );

    return historyFilePath;
  }

  /**
   * Get all history entries for a specific request file
   */
  async getHistory(requestFile: string): Promise<HistoryEntry[]> {
    if (!await exists(this.historyDir)) {
      return [];
    }

    const requestBaseName = path.basename(requestFile, ".http");
    const entries: HistoryEntry[] = [];

    for await (const dirEntry of Deno.readDir(this.historyDir)) {
      if (dirEntry.isFile && dirEntry.name.startsWith(requestBaseName) && dirEntry.name.endsWith(".json")) {
        const filePath = path.join(this.historyDir, dirEntry.name);
        const content = await Deno.readTextFile(filePath);
        const entry: HistoryEntry = JSON.parse(content);
        entries.push(entry);
      }
    }

    // Sort by timestamp (newest first)
    entries.sort((a, b) => b.timestamp.localeCompare(a.timestamp));

    return entries;
  }

  /**
   * Clear history for a specific request file
   */
  async clearHistory(requestFile: string): Promise<number> {
    if (!await exists(this.historyDir)) {
      return 0;
    }

    const requestBaseName = path.basename(requestFile, ".http");
    let count = 0;

    for await (const dirEntry of Deno.readDir(this.historyDir)) {
      if (dirEntry.isFile && dirEntry.name.startsWith(requestBaseName) && dirEntry.name.endsWith(".json")) {
        const filePath = path.join(this.historyDir, dirEntry.name);
        await Deno.remove(filePath);
        count++;
      }
    }

    return count;
  }

  /**
   * Clear all history
   */
  async clearAllHistory(): Promise<number> {
    if (!await exists(this.historyDir)) {
      return 0;
    }

    let count = 0;
    for await (const dirEntry of Deno.readDir(this.historyDir)) {
      if (dirEntry.isFile && dirEntry.name.endsWith(".json")) {
        const filePath = path.join(this.historyDir, dirEntry.name);
        await Deno.remove(filePath);
        count++;
      }
    }

    return count;
  }

  /**
   * Get the latest history entry for a request file
   */
  async getLatest(requestFile: string): Promise<HistoryEntry | null> {
    const history = await this.getHistory(requestFile);
    return history.length > 0 ? history[0] : null;
  }
}
