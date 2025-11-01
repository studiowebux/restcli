import { walk } from "@std/fs";
import * as path from "@std/path";
import { stringify as yamlStringify } from "@std/yaml";
import { applyVariables, parseHttpFile, type Documentation } from "./parser.ts";
import { RequestExecutor, type RequestResult } from "./executor.ts";
import { SessionManager } from "./session.ts";
import { ConfigManager } from "./config.ts";
import { type HistoryEntry, HistoryManager } from "./history.ts";

interface FileEntry {
  path: string;
  name: string;
  isDirectory: boolean;
}

class TUI {
  private files: FileEntry[] = [];
  private selectedIndex = 0;
  private response: RequestResult | null = null;
  private sessionManager: SessionManager;
  private historyManager: HistoryManager;
  private executor: RequestExecutor;
  private requestsDir = "./requests";
  private baseDir = ".";
  private running = false;
  private statusMessage = "";
  private searchMode = false;
  private historyMode = false;
  private historyEntries: HistoryEntry[] = [];
  private historyIndex = 0;
  private searchQuery = "";
  private searchResults: number[] = [];
  private searchResultIndex = 0;
  private gotoMode = false;
  private gotoQuery = "";
  private pageSize = 10; // Will be calculated dynamically
  private fullscreenMode = false;
  private variableMode = false;
  private variableIndex = 0;
  private variableEditMode: "list" | "add" | "edit" | "delete" = "list";
  private variableEditKey = "";
  private variableEditValue = "";
  private variableEditField: "key" | "value" = "key";
  private headerMode = false;
  private headerIndex = 0;
  private headerEditMode: "list" | "add" | "edit" | "delete" = "list";
  private headerEditKey = "";
  private headerEditValue = "";
  private headerEditField: "key" | "value" = "key";
  private responseScrollOffset = 0; // For scrolling through response body
  private maxResponseScrollOffset = 0; // Maximum scroll offset for response body
  private showResponseHeaders = true; // Toggle response headers visibility
  private showResponseBody = true; // Toggle response body visibility
  private helpMode = false; // Show help modal
  private helpScrollOffset = 0; // Scroll offset for help modal
  private maxHelpScrollOffset = 0; // Maximum scroll offset for help modal
  private documentationMode = false; // Show documentation panel
  private documentationScrollOffset = 0; // Scroll offset for documentation panel
  private maxDocumentationScrollOffset = 0; // Maximum scroll offset for documentation panel
  private documentationCursorIndex = 0; // Current cursor position in documentation
  private documentationCollapsedFields = new Set<string>(); // Set of collapsed field paths
  private documentationMaxCursorIndex = 0; // Maximum cursor index

  constructor() {
    this.sessionManager = new SessionManager();
    this.historyManager = new HistoryManager();
    this.executor = new RequestExecutor();
  }

  async init(): Promise<void> {
    // Check if config directory exists, use it if available
    const configManager = new ConfigManager();
    const isInitialized = await configManager.isInitialized();

    if (isInitialized) {
      this.baseDir = configManager.getConfigDir();
      this.sessionManager = new SessionManager(this.baseDir);
      this.historyManager = new HistoryManager(this.baseDir);
      this.requestsDir = path.join(this.baseDir, "requests");
    } else {
      // Check if running in a directory with requests/
      const hasLocalRequests = await Deno.stat("./requests").then(() => true)
        .catch(() => false);

      if (hasLocalRequests) {
        // Use current directory (backward compatibility for local development)
        this.baseDir = ".";
        this.sessionManager = new SessionManager();
        this.historyManager = new HistoryManager();
        this.requestsDir = "./requests";

        // Show helpful message
        console.log(
          "\n💡 Tip: Run 'deno task init --migrate' to migrate to ~/.restcli/",
        );
        console.log("   This allows you to use restcli from any directory!\n");
        await new Promise((resolve) => setTimeout(resolve, 2000)); // Show for 2 seconds
      } else {
        // No local requests/ and no ~/.restcli/, auto-initialize
        console.log("\n🚀 First time setup: Initializing restcli...\n");
        await configManager.init();
        await configManager.createExamples();

        console.log(`\n✅ Initialized at: ${configManager.getConfigDir()}`);
        console.log("\n📝 Example files created. Edit them to get started!");
        console.log(
          "   Config: ~/.restcli/.profiles.json and ~/.restcli/.session.json",
        );
        console.log("   Requests: ~/.restcli/requests/\n");
        await new Promise((resolve) => setTimeout(resolve, 3000)); // Show for 3 seconds

        this.baseDir = configManager.getConfigDir();
        this.sessionManager = new SessionManager(this.baseDir);
        this.historyManager = new HistoryManager(this.baseDir);
        this.requestsDir = path.join(this.baseDir, "requests");
      }
    }

    await this.sessionManager.load();

    // Update requestsDir based on active profile's workdir
    this.requestsDir = this.sessionManager.getWorkdir();

    await this.loadFiles();
  }

  async loadFiles(): Promise<void> {
    this.files = [];
    try {
      for await (
        const entry of walk(this.requestsDir, {
          exts: [".http", ".yaml", ".yml"],
        })
      ) {
        if (entry.isFile) {
          const relativePath = path.relative(this.requestsDir, entry.path);
          this.files.push({
            path: entry.path,
            name: relativePath,
            isDirectory: false,
          });
        }
      }
      this.files.sort((a, b) => a.name.localeCompare(b.name));
    } catch {
      // Requests directory doesn't exist yet
    }
  }

  private clear(): void {
    // Don't clear entire screen - just move cursor to top
    // We use \x1b[K on each line to clear to end of line
    this.moveCursor(1, 1);
  }

  private moveCursor(row: number, col: number): void {
    Deno.stdout.writeSync(new TextEncoder().encode(`\x1b[${row};${col}H`));
  }

  private write(text: string): void {
    Deno.stdout.writeSync(new TextEncoder().encode(text));
  }

  /**
   * Delete the last word from a string (used for Option+Delete / Alt+Backspace)
   * Removes the last sequence of non-whitespace characters and any trailing whitespace
   */
  private deleteLastWord(text: string): string {
    if (!text) return text;

    // Remove trailing whitespace first
    let trimmed = text.trimEnd();
    if (trimmed === "") return "";

    // Find the last word boundary (space, dot, slash, dash, etc.)
    const wordBoundaryRegex = /[\s.\-_/\\]+[^\s.\-_/\\]*$/;
    const match = trimmed.match(wordBoundaryRegex);

    if (match) {
      return trimmed.slice(0, match.index);
    }

    // If no word boundary found, clear everything
    return "";
  }

  /**
   * Beautify JSON response body if it's valid JSON
   */
  private beautifyJson(body: string): string {
    try {
      const json = JSON.parse(body);
      return JSON.stringify(json, null, 2);
    } catch {
      // Not JSON, return as-is
      return body;
    }
  }

  /**
   * Format bytes to human-readable size
   */
  private formatBytes(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 10) / 10 + " " + sizes[i];
  }

  /**
   * Check if a line contains a URL
   * @param text The text to check
   * @returns true if the line contains a URL pattern
   */
  private containsUrl(text: string): boolean {
    const urlPattern = /https?:\/\/|www\./i;
    return urlPattern.test(text);
  }

  /**
   * Wrap a long line into multiple lines at word boundaries
   * Handles URLs differently based on fullscreen mode
   * @param text The text to wrap
   * @param maxWidth Maximum width per line
   * @param allowFullUrls If true (fullscreen), show full URLs; if false, truncate them
   * @returns Array of wrapped lines
   */
  private wrapLine(
    text: string,
    maxWidth: number,
    allowFullUrls = false,
  ): string[] {
    if (text.length <= maxWidth) {
      return [text];
    }

    // Handle URLs based on mode
    if (this.containsUrl(text)) {
      if (allowFullUrls) {
        // Fullscreen: keep full URL on one line for terminal link detection
        return [text];
      } else {
        // Non-fullscreen: truncate to fit width and prevent sidebar clash
        return [text.slice(0, maxWidth - 3) + "..."];
      }
    }

    const wrapped: string[] = [];
    let currentLine = "";

    // Split by spaces to preserve word boundaries
    const words = text.split(" ");

    for (const word of words) {
      // If the word itself is longer than maxWidth, we need to hard-break it
      if (word.length > maxWidth) {
        // Flush current line if it has content
        if (currentLine) {
          wrapped.push(currentLine);
          currentLine = "";
        }
        // Break the long word into chunks
        for (let i = 0; i < word.length; i += maxWidth) {
          wrapped.push(word.slice(i, i + maxWidth));
        }
        continue;
      }

      // Try adding the word to the current line
      const testLine = currentLine ? `${currentLine} ${word}` : word;

      if (testLine.length <= maxWidth) {
        currentLine = testLine;
      } else {
        // Word doesn't fit, push current line and start new one
        if (currentLine) {
          wrapped.push(currentLine);
        }
        currentLine = word;
      }
    }

    // Don't forget the last line
    if (currentLine) {
      wrapped.push(currentLine);
    }

    return wrapped.length > 0 ? wrapped : [""];
  }

  private draw(): void {
    this.clear();

    const width = Deno.consoleSize().columns;
    const height = Deno.consoleSize().rows;

    let sidebarWidth: number;
    let separatorCol: number;
    let mainStartCol: number;
    let mainWidth: number;

    if (this.fullscreenMode) {
      // Fullscreen: no sidebar, use full width
      sidebarWidth = 0;
      separatorCol = 0;
      mainStartCol = 1;
      mainWidth = width;
    } else {
      // Normal mode: show sidebar
      sidebarWidth = Math.min(60, Math.floor(width * 0.4));
      separatorCol = sidebarWidth + 1;
      mainStartCol = separatorCol + 2;
      mainWidth = width - mainStartCol;
    }

    // Calculate page size for page up/down
    this.pageSize = Math.max(1, height - 7); // Reserve space for header, title, scroll indicator, status

    // Header
    this.drawHeader(width);

    if (!this.fullscreenMode) {
      // Sidebar (only in normal mode)
      this.drawSidebar(sidebarWidth, height - 1);

      // Vertical separator (only in normal mode)
      this.drawSeparator(separatorCol, height);
    }

    // Main content
    this.drawMain(mainStartCol, mainWidth, height - 1);

    // Status bar
    this.drawStatusBar(width, height);

    // Position cursor at bottom
    this.moveCursor(height, 1);
  }

  private drawSeparator(col: number, height: number): void {
    // Build the entire separator in one buffer to reduce cursor movements
    const encoder = new TextEncoder();
    const parts: Uint8Array[] = [];

    for (let row = 2; row <= height; row++) {
      parts.push(encoder.encode(`\x1b[${row};${col}H\x1b[2m│\x1b[0m`));
    }

    // Write all at once
    const totalLength = parts.reduce((sum, part) => sum + part.length, 0);
    const buffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const part of parts) {
      buffer.set(part, offset);
      offset += part.length;
    }
    Deno.stdout.writeSync(buffer);
  }

  private drawHeader(width: number): void {
    const activeProfile = this.sessionManager.getActiveProfile();
    const profileName = activeProfile ? activeProfile.name : "No Profile";
    const fullscreenTag = this.fullscreenMode ? " [FULLSCREEN]" : "";
    const header = ` HTTP TUI | Profile: ${profileName}${fullscreenTag} `;
    const padding = " ".repeat(Math.max(0, width - header.length));

    this.moveCursor(1, 1);
    this.write(`\x1b[7m${header}${padding}\x1b[0m`);
  }

  private drawSidebar(width: number, height: number): void {
    // Title with count
    this.moveCursor(2, 1);
    const totalFiles = this.files.length;
    const visibleCount = Math.min(height - 3, totalFiles);
    const title = ` Files (${totalFiles}) `;
    const titlePadding = " ".repeat(Math.max(0, width - title.length));
    this.write(`\x1b[1m${title}\x1b[0m${titlePadding}`);

    const maxVisibleLines = height - 4; // Reserve one line for scroll indicator
    const startIdx = Math.max(
      0,
      Math.min(
        this.selectedIndex - Math.floor(maxVisibleLines / 2),
        totalFiles - maxVisibleLines,
      ),
    );
    const endIdx = Math.min(startIdx + maxVisibleLines, totalFiles);
    const visibleFiles = this.files.slice(startIdx, endIdx);

    // Calculate number width (for alignment) - using hex
    const maxNum = this.files.length;
    const numWidth = maxNum.toString(16).length;

    for (let i = 0; i < maxVisibleLines; i++) {
      this.moveCursor(3 + i, 1);
      if (i < visibleFiles.length) {
        const file = visibleFiles[i];
        const globalIdx = startIdx + i;
        const isSelected = globalIdx === this.selectedIndex;

        // Line number in hex
        const lineNum = (globalIdx + 1).toString(16).toUpperCase().padStart(
          numWidth,
          " ",
        );
        const lineNumDisplay = `\x1b[2m${lineNum}\x1b[0m`;

        // Prefix
        const prefixVisible = isSelected ? ">" : " ";

        // Calculate max display name length (width - numWidth - 2 for prefix and space)
        const maxNameWidth = width - numWidth - 2;
        let displayName = file.name;
        if (displayName.length > maxNameWidth) {
          displayName = displayName.slice(0, maxNameWidth - 3) + "...";
        }

        // Pad to exactly fill the sidebar width
        const totalContentLength = numWidth + 1 + displayName.length; // num + prefix + name
        const padding = " ".repeat(Math.max(0, width - totalContentLength));

        // Apply styling
        if (isSelected) {
          this.write(
            `${lineNumDisplay}\x1b[7m${prefixVisible}${displayName}\x1b[0m${padding}`,
          );
        } else {
          this.write(
            `${lineNumDisplay}${prefixVisible}${displayName}${padding}`,
          );
        }
      } else {
        const clearLine = " ".repeat(width);
        this.write(clearLine);
      }
    }

    // Scroll indicator
    this.moveCursor(3 + maxVisibleLines, 1);
    if (totalFiles > maxVisibleLines) {
      const scrollPercent = Math.round(
        (this.selectedIndex / (totalFiles - 1)) * 100,
      );
      const hasMore = endIdx < totalFiles;
      const hasPrevious = startIdx > 0;

      let indicator = "\x1b[2m";
      if (hasPrevious && hasMore) {
        indicator += `↕ ${scrollPercent}% (${endIdx - startIdx}/${totalFiles})`;
      } else if (hasPrevious) {
        indicator += `↑ Bottom (${endIdx - startIdx}/${totalFiles})`;
      } else if (hasMore) {
        indicator += `↓ More below (${endIdx - startIdx}/${totalFiles})`;
      } else {
        indicator += `All ${totalFiles} files`;
      }
      indicator += "\x1b[0m";

      const indicatorPadding = " ".repeat(Math.max(0, width - 20)); // Rough estimate
      this.write(indicator + indicatorPadding);
    } else {
      const clearLine = " ".repeat(width);
      this.write(clearLine);
    }
  }

  private drawMain(startCol: number, width: number, height: number): void {
    // Clear the main content area (since we no longer clear the entire screen)
    // Batch all clear operations into one write to avoid flickering
    const encoder = new TextEncoder();
    const parts: Uint8Array[] = [];
    for (let row = 2; row <= height; row++) {
      parts.push(encoder.encode(`\x1b[${row};${startCol}H\x1b[K`));
    }
    const totalLength = parts.reduce((sum, part) => sum + part.length, 0);
    const buffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const part of parts) {
      buffer.set(part, offset);
      offset += part.length;
    }
    Deno.stdout.writeSync(buffer);

    // Check if in help mode
    if (this.helpMode) {
      this.drawHelpModal(startCol, width, height);
      return;
    }

    // Check if in documentation mode
    if (this.documentationMode) {
      this.drawDocumentation(startCol, width, height);
      return;
    }

    // Check if in variable mode
    if (this.variableMode) {
      this.drawVariableEditor(startCol, width, height);
      return;
    }

    // Check if in header mode
    if (this.headerMode) {
      this.drawHeaderEditor(startCol, width, height);
      return;
    }

    // Check if in history mode
    if (this.historyMode) {
      this.drawHistoryViewer(startCol, width, height);
      return;
    }

    this.moveCursor(2, startCol);
    const title = " Response ";
    const titleTruncated = title.slice(0, width);
    this.write(`\x1b[1m${titleTruncated}\x1b[0m\x1b[K`);

    if (!this.response) {
      this.moveCursor(4, startCol);
      const noResponseText = " No request executed yet ".slice(0, width);
      this.write(`\x1b[2m${noResponseText}\x1b[0m\x1b[K`);
      return;
    }

    // Check if in inspection mode
    const inspectionMode = (this.response as any).inspectionMode;
    if (inspectionMode) {
      this.drawInspection(startCol, width, height);
      return;
    }

    let line = 3;
    this.moveCursor(line++, startCol);

    // Status line
    const statusColor =
      this.response.status >= 200 && this.response.status < 300
        ? "\x1b[32m" // Green
        : this.response.status >= 400
        ? "\x1b[31m" // Red
        : "\x1b[33m"; // Yellow

    if (this.response.error) {
      // Wrap error message across multiple lines
      const errorPrefix = "Error: ";
      const errorMessage = this.response.error;
      const fullError = errorPrefix + errorMessage;
      const maxLineWidth = width - 2;

      // Split into chunks that fit
      const errorLines: string[] = [];
      for (let i = 0; i < fullError.length; i += maxLineWidth) {
        errorLines.push(fullError.slice(i, i + maxLineWidth));
      }

      // Display first line
      this.write(`\x1b[31m${errorLines[0]}\x1b[0m\x1b[K`);

      // Display remaining lines
      for (let i = 1; i < Math.min(errorLines.length, 3); i++) { // Limit to 3 lines total
        line++;
        this.moveCursor(line, startCol);
        this.write(`\x1b[31m${errorLines[i]}\x1b[0m\x1b[K`);
      }

      line++;
    } else {
      const statusText =
        `${this.response.status} ${this.response.statusText} | ${
          Math.round(this.response.duration)
        }ms | Req: ${this.formatBytes(this.response.requestSize)} | Res: ${this.formatBytes(this.response.responseSize)}`.slice(0, width - 2);
      this.write(`${statusColor}${statusText}\x1b[0m\x1b[K`);
      line++;
    }

    // Headers
    if (Object.keys(this.response.headers).length > 0) {
      this.moveCursor(line++, startCol);
      const headerCount = Object.keys(this.response.headers).length;
      const headerToggle = this.showResponseHeaders ? "[-]" : "[+]";
      this.write(
        `\x1b[2mHeaders (${headerCount}) ${headerToggle} Press Shift+H to toggle\x1b[0m\x1b[K`,
      );

      if (this.showResponseHeaders) {
        for (
          const [key, value] of Object.entries(this.response.headers).slice(
            0,
            5,
          )
        ) {
          this.moveCursor(line++, startCol);
          const display = `${key}: ${value}`.slice(0, width - 2);
          this.write(`  ${display}\x1b[K`);
        }
        this.moveCursor(line++, startCol);
        this.write("\x1b[K");
      }
    }

    // Body
    const bodyLines = this.response.body.split("\n");
    const maxWidth = Math.max(1, width - 2); // Ensure at least 1 char width

    // Wrap long lines and flatten into a single array
    // Process ALL lines to enable scrolling through long responses
    const wrappedLines: string[] = [];
    for (const bodyLine of bodyLines) {
      const wrapped = this.wrapLine(bodyLine, maxWidth, this.fullscreenMode);
      wrappedLines.push(...wrapped);
    }

    // Show "Body:" with scroll indicator and toggle
    this.moveCursor(line++, startCol);
    const bodyToggle = this.showResponseBody ? "[-]" : "[+]";
    this.write(`\x1b[2mBody ${bodyToggle} Press Shift+B to toggle\x1b[0m\x1b[K`);

    // Calculate maxLines AFTER writing the Body: header
    const maxLines = height - line;

    // Calculate scroll position
    const totalWrappedLines = wrappedLines.length;
    this.maxResponseScrollOffset = Math.max(0, totalWrappedLines - maxLines);
    const startIndex = Math.max(
      0,
      Math.min(
        this.responseScrollOffset,
        this.maxResponseScrollOffset,
      ),
    );

    // Show scroll indicator on separate line if needed
    if (this.showResponseBody && totalWrappedLines > maxLines) {
      this.moveCursor(line++, startCol);
      const scrollProgress = `\x1b[2m[${startIndex + 1}-${
        Math.min(startIndex + maxLines, totalWrappedLines)
      }/${totalWrappedLines}] j/k to scroll\x1b[0m`;
      this.write(`${scrollProgress}\x1b[K`);
    }

    // Display wrapped lines with scrolling
    if (this.showResponseBody) {
      for (
        let i = startIndex; i < wrappedLines.length && line <= height; i++
      ) {
        const wrappedLine = wrappedLines[i];

        // Calculate how many visual lines this will take
        // In fullscreen, URLs can wrap naturally; in non-fullscreen they're already truncated
        const visualLinesNeeded = this.fullscreenMode
          ? Math.max(1, Math.ceil(wrappedLine.length / maxWidth))
          : 1; // Non-fullscreen: already truncated, takes 1 line

        // Stop if we don't have enough space
        if (line + visualLinesNeeded > height + 1) break;

        this.moveCursor(line, startCol);
        this.write(`${wrappedLine}\x1b[K`);

        // Increment by the number of visual lines this took
        line += visualLinesNeeded;
      }
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawInspection(
    startCol: number,
    width: number,
    height: number,
  ): void {
    const inspectionData = (this.response as any).inspectionData;
    let line = 3;

    // Request line
    this.moveCursor(line++, startCol);
    const requestLine = `${inspectionData.method} ${inspectionData.url}`.slice(
      0,
      width - 2,
    );
    this.write(`\x1b[1;36m${requestLine}\x1b[0m\x1b[K`);

    // Headers
    if (this.response && Object.keys(this.response.headers).length > 0) {
      this.moveCursor(line++, startCol);
      const headerCount = Object.keys(this.response.headers).length;
      const headerToggle = this.showResponseHeaders ? "[-]" : "[+]";
      this.write(
        `\x1b[2mHeaders (${headerCount}, merged with profile) ${headerToggle} Press Shift+H to toggle\x1b[0m\x1b[K`,
      );

      if (this.showResponseHeaders) {
        for (const [key, value] of Object.entries(this.response.headers)) {
          this.moveCursor(line++, startCol);
          const display = `  ${key}: ${value}`.slice(0, width - 2);
          this.write(`${display}\x1b[K`);

          if (line >= height - 5) break; // Leave room for body
        }

        this.moveCursor(line++, startCol);
        this.write("\x1b[K");
      }
    }

    // Body
    if (this.response && this.response.body) {
      const bodyLines = this.response.body.split("\n");
      const maxWidth = Math.max(1, width - 2); // Ensure at least 1 char width

      // Wrap long lines and flatten into a single array
      // Process ALL lines to enable scrolling through long responses
      const wrappedLines: string[] = [];
      for (const bodyLine of bodyLines) {
        const wrapped = this.wrapLine(bodyLine, maxWidth, this.fullscreenMode);
        wrappedLines.push(...wrapped);
      }

      // Show "Body:" with scroll indicator and toggle
      this.moveCursor(line++, startCol);
      const bodyToggle = this.showResponseBody ? "[-]" : "[+]";
      this.write(`\x1b[2mBody ${bodyToggle} Press Shift+B to toggle\x1b[0m\x1b[K`);

      // Calculate maxLines AFTER writing the Body: header
      const maxLines = height - line;

      // Calculate scroll position
      const totalWrappedLines = wrappedLines.length;
      this.maxResponseScrollOffset = Math.max(0, totalWrappedLines - maxLines);
      const startIndex = Math.max(
        0,
        Math.min(
          this.responseScrollOffset,
          this.maxResponseScrollOffset,
        ),
      );

      // Show scroll indicator on separate line if needed
      if (this.showResponseBody && totalWrappedLines > maxLines) {
        this.moveCursor(line++, startCol);
        const scrollProgress = `\x1b[2m[${startIndex + 1}-${
          Math.min(startIndex + maxLines, totalWrappedLines)
        }/${totalWrappedLines}] j/k to scroll\x1b[0m`;
        this.write(`${scrollProgress}\x1b[K`);
      }

      // Display wrapped lines with scrolling
      if (this.showResponseBody) {
        for (
          let i = startIndex; i < wrappedLines.length && line <= height; i++
        ) {
          const wrappedLine = wrappedLines[i];

          // Calculate how many visual lines this will take
          // In fullscreen, URLs can wrap naturally; in non-fullscreen they're already truncated
          const visualLinesNeeded = this.fullscreenMode
            ? Math.max(1, Math.ceil(wrappedLine.length / maxWidth))
            : 1; // Non-fullscreen: already truncated, takes 1 line

          // Stop if we don't have enough space
          if (line + visualLinesNeeded > height + 1) break;

          this.moveCursor(line, startCol);
          this.write(`${wrappedLine}\x1b[K`);

          // Increment by the number of visual lines this took
          line += visualLinesNeeded;
        }
      }
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawVariableEditor(
    startCol: number,
    width: number,
    height: number,
  ): void {
    this.moveCursor(2, startCol);
    const activeProfile = this.sessionManager.getActiveProfile();
    const profileName = activeProfile ? activeProfile.name : "No Profile";
    const title = ` Variables (${profileName}) `;
    this.write(`\x1b[1m${title}\x1b[0m\x1b[K`);

    let line = 3;
    const variables = this.sessionManager.getProfileVariables();
    const varEntries = Object.entries(variables);

    // List mode
    if (this.variableEditMode === "list") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[2mTotal: ${varEntries.length} variables\x1b[0m\x1b[K`);
      line++;

      const maxVisibleLines = height - line - 5;
      for (let i = 0; i < Math.min(varEntries.length, maxVisibleLines); i++) {
        this.moveCursor(line++, startCol);
        const [key, value] = varEntries[i];
        const isSelected = i === this.variableIndex;

        const truncatedValue = value.length > width - key.length - 8
          ? value.slice(0, width - key.length - 11) + "..."
          : value;
        const display = `${key}: ${truncatedValue}`;
        const displayTruncated = display.slice(0, width - 4);

        if (isSelected) {
          this.write(`\x1b[7m> ${displayTruncated}\x1b[0m\x1b[K`);
        } else {
          this.write(`  ${displayTruncated}\x1b[K`);
        }
      }
    } // Add mode
    else if (this.variableEditMode === "add") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[1mAdd New Variable\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      const keyLabel = "Key: ";
      const maxKeyWidth = width - keyLabel.length - 2;
      const keyDisplay =
        (this.variableEditKey + (this.variableEditField === "key" ? "_" : ""))
          .slice(0, maxKeyWidth);
      const keyLine = keyLabel + keyDisplay;
      if (this.variableEditField === "key") {
        this.write(`${keyLabel}\x1b[7m${keyDisplay}\x1b[0m\x1b[K`);
      } else {
        this.write(`${keyLine}\x1b[K`);
      }

      this.moveCursor(line++, startCol);
      const valueLabel = "Value: ";
      const maxValueWidth = width - valueLabel.length - 2;
      const valueDisplay =
        (this.variableEditValue +
          (this.variableEditField === "value" ? "_" : "")).slice(
            0,
            maxValueWidth,
          );
      const valueLine = valueLabel + valueDisplay;
      if (this.variableEditField === "value") {
        this.write(`${valueLabel}\x1b[7m${valueDisplay}\x1b[0m\x1b[K`);
      } else {
        this.write(`${valueLine}\x1b[K`);
      }
    } // Edit mode
    else if (this.variableEditMode === "edit") {
      this.moveCursor(line++, startCol);
      const editTitle = `Edit Variable: ${this.variableEditKey}`.slice(
        0,
        width - 2,
      );
      this.write(`\x1b[1m${editTitle}\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      const valueLabel = "Value: ";
      const maxValueWidth = width - valueLabel.length - 2; // -2 for cursor and padding
      const valueWithCursor = this.variableEditValue + "_";
      // Show the END of the value so user can see what they're typing (important for long values like JWTs)
      const valueDisplay = valueWithCursor.length > maxValueWidth
        ? valueWithCursor.slice(-maxValueWidth)
        : valueWithCursor;
      this.write(`${valueLabel}\x1b[7m${valueDisplay}\x1b[0m\x1b[K`);
    } // Delete confirmation
    else if (this.variableEditMode === "delete") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[1;31mDelete Variable?\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      this.write(`Key: ${this.variableEditKey}\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      this.write(`\x1b[2mPress [Y] to confirm, [N] to cancel\x1b[0m\x1b[K`);
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawHeaderEditor(
    startCol: number,
    width: number,
    height: number,
  ): void {
    this.moveCursor(2, startCol);
    const activeProfile = this.sessionManager.getActiveProfile();
    const profileName = activeProfile ? activeProfile.name : "No Profile";
    const title = ` Headers (${profileName}) `;
    this.write(`\x1b[1m${title}\x1b[0m\x1b[K`);

    let line = 3;
    const headers = this.sessionManager.getProfileHeaders();
    const headerEntries = Object.entries(headers);

    // List mode
    if (this.headerEditMode === "list") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[2mTotal: ${headerEntries.length} headers\x1b[0m\x1b[K`);
      line++;

      const maxVisibleLines = height - line - 5;
      for (
        let i = 0; i < Math.min(headerEntries.length, maxVisibleLines); i++
      ) {
        this.moveCursor(line++, startCol);
        const [key, value] = headerEntries[i];
        const isSelected = i === this.headerIndex;

        const truncatedValue = value.length > width - key.length - 8
          ? value.slice(0, width - key.length - 11) + "..."
          : value;
        const display = `${key}: ${truncatedValue}`;
        const displayTruncated = display.slice(0, width - 4);

        if (isSelected) {
          this.write(`\x1b[7m> ${displayTruncated}\x1b[0m\x1b[K`);
        } else {
          this.write(`  ${displayTruncated}\x1b[K`);
        }
      }
    } // Add mode
    else if (this.headerEditMode === "add") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[1mAdd New Header\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      const keyLabel = "Key: ";
      const maxKeyWidth = width - keyLabel.length - 2;
      const keyDisplay =
        (this.headerEditKey + (this.headerEditField === "key" ? "_" : ""))
          .slice(0, maxKeyWidth);
      const keyLine = keyLabel + keyDisplay;
      if (this.headerEditField === "key") {
        this.write(`${keyLabel}\x1b[7m${keyDisplay}\x1b[0m\x1b[K`);
      } else {
        this.write(`${keyLine}\x1b[K`);
      }

      this.moveCursor(line++, startCol);
      const valueLabel = "Value: ";
      const maxValueWidth = width - valueLabel.length - 2;
      const valueDisplay =
        (this.headerEditValue + (this.headerEditField === "value" ? "_" : ""))
          .slice(0, maxValueWidth);
      const valueLine = valueLabel + valueDisplay;
      if (this.headerEditField === "value") {
        this.write(`${valueLabel}\x1b[7m${valueDisplay}\x1b[0m\x1b[K`);
      } else {
        this.write(`${valueLine}\x1b[K`);
      }
    } // Edit mode
    else if (this.headerEditMode === "edit") {
      this.moveCursor(line++, startCol);
      const editTitle = `Edit Header: ${this.headerEditKey}`.slice(
        0,
        width - 2,
      );
      this.write(`\x1b[1m${editTitle}\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      const valueLabel = "Value: ";
      const maxValueWidth = width - valueLabel.length - 2; // -2 for cursor and padding
      const valueWithCursor = this.headerEditValue + "_";
      // Show the END of the value so user can see what they're typing (important for long values like JWTs)
      const valueDisplay = valueWithCursor.length > maxValueWidth
        ? valueWithCursor.slice(-maxValueWidth)
        : valueWithCursor;
      this.write(`${valueLabel}\x1b[7m${valueDisplay}\x1b[0m\x1b[K`);
    } // Delete confirmation
    else if (this.headerEditMode === "delete") {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[1;31mDelete Header?\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      this.write(`Key: ${this.headerEditKey}\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      this.write(`\x1b[2mPress [Y] to confirm, [N] to cancel\x1b[0m\x1b[K`);
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawHistoryViewer(
    startCol: number,
    width: number,
    height: number,
  ): void {
    const file = this.files[this.selectedIndex];
    this.moveCursor(2, startCol);
    const title = ` History for ${file.name} `;
    this.write(`\x1b[1m${title.slice(0, width)}\x1b[0m\x1b[K`);

    let line = 3;

    if (this.historyEntries.length === 0) {
      this.moveCursor(line++, startCol);
      this.write(`\x1b[2mNo history entries found\x1b[0m\x1b[K`);
      line++;

      this.moveCursor(line++, startCol);
      this.write(
        `\x1b[2mExecute this request to create history entries\x1b[0m\x1b[K`,
      );
    } else {
      this.moveCursor(line++, startCol);
      this.write(
        `\x1b[2mTotal: ${this.historyEntries.length} entries (newest first)\x1b[0m\x1b[K`,
      );
      line++;

      const maxVisibleLines = height - line - 2;
      let entriesShown = 0;

      for (let i = 0; i < this.historyEntries.length; i++) {
        const entry = this.historyEntries[i];
        const isSelected = i === this.historyIndex;

        // Format timestamp
        const date = new Date(entry.timestamp);
        const timeStr = date.toLocaleString("en-US", {
          month: "short",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        });

        // Status color
        const statusColor =
          entry.responseStatus >= 200 && entry.responseStatus < 300
            ? "\x1b[32m" // Green
            : entry.responseStatus >= 400
            ? "\x1b[31m" // Red
            : "\x1b[33m"; // Yellow

        // Format display line
        const status = entry.error ? "ERR" : `${entry.responseStatus}`;
        const duration = `${Math.round(entry.duration)}ms`;
        const prefix =
          `${timeStr} | ${statusColor}${status}\x1b[0m | ${duration} | ${entry.method} `;

        let display: string;
        let linesNeeded: number;

        if (this.fullscreenMode) {
          // Fullscreen: show full URL, let it wrap
          display = prefix + entry.url;
          const visibleLength = timeStr.length + 3 + status.length + 3 +
            duration.length + 3 + entry.method.length + 1 + entry.url.length +
            2; // +2 for "> "
          linesNeeded = Math.max(
            1,
            Math.ceil(visibleLength / Math.max(1, width)),
          );
        } else {
          // Non-fullscreen: truncate URL to prevent sidebar clash
          const maxUrlLength = width -
            (timeStr.length + 3 + status.length + 3 + duration.length + 3 +
              entry.method.length + 1 + 2 + 10); // +10 for safety margin
          const truncatedUrl = entry.url.length > maxUrlLength
            ? entry.url.slice(0, Math.max(10, maxUrlLength)) + "..."
            : entry.url;
          display = prefix + truncatedUrl;
          linesNeeded = 1; // Truncated, fits on one line
        }

        // Stop if we don't have enough space left
        if (line + linesNeeded > height - 2) break;

        this.moveCursor(line, startCol);
        if (isSelected) {
          this.write(`\x1b[7m> ${display}\x1b[0m\x1b[K`);
        } else {
          this.write(`  ${display}\x1b[K`);
        }

        line += linesNeeded;
        entriesShown++;
      }
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawHelpModal(startCol: number, width: number, height: number): void {
    this.moveCursor(2, startCol);
    const title = " Keyboard Shortcuts ";
    this.write(`\x1b[1m${title}\x1b[0m\x1b[K`);

    const shortcuts = [
      {
        category: "Navigation",
        items: [
          { key: "↑/↓", desc: "Navigate files (circular wrapping)" },
          { key: "Page Up/Down", desc: "Fast scroll through file list" },
          { key: ":", desc: "Goto line in hex (e.g., :64 → file #100)" },
          {
            key: "Ctrl+R",
            desc: "Search files by name (Ctrl+R again to cycle)",
          },
        ],
      },
      {
        category: "Actions",
        items: [
          { key: "Enter", desc: "Execute selected request" },
          { key: "i", desc: "Inspect request (preview without executing)" },
          { key: "d", desc: "Duplicate current file" },
          { key: "s", desc: "Save response to file (timestamp)" },
          { key: "c", desc: "Copy response body to clipboard" },
          { key: "r", desc: "Refresh file list" },
        ],
      },
      {
        category: "Response View",
        items: [
          { key: "j/k", desc: "Scroll response down/up" },
          { key: "Shift+H", desc: "Toggle response headers visibility" },
          { key: "Shift+B", desc: "Toggle response body visibility" },
          { key: "f", desc: "Toggle fullscreen mode" },
        ],
      },
      {
        category: "Profiles & Variables",
        items: [
          { key: "p", desc: "Switch profile (cycles through profiles)" },
          { key: "v", desc: "Open variable editor" },
          { key: "h", desc: "Open header editor" },
          { key: "Ctrl+H", desc: "View request history" },
        ],
      },
      {
        category: "Other",
        items: [
          { key: "m", desc: "View request documentation" },
          { key: "?", desc: "Show this help (you are here!)" },
          { key: "ESC", desc: "Clear status / Cancel search or goto" },
          { key: "q", desc: "Quit" },
        ],
      },
    ];

    // Build all content lines first
    const contentLines: string[] = [];
    for (const section of shortcuts) {
      contentLines.push(`\x1b[1;36m${section.category}:\x1b[0m`);
      contentLines.push(""); // Empty line

      for (const item of section.items) {
        const keyPart = `  \x1b[33m${item.key.padEnd(15)}\x1b[0m`;
        const descPart = item.desc.slice(0, width - 20);
        contentLines.push(`${keyPart} ${descPart}`);
      }

      contentLines.push(""); // Empty line between sections
    }

    // Add footer
    contentLines.push("");
    contentLines.push("\x1b[2mPress ESC or ? to close | ↑/↓ to scroll\x1b[0m");

    // Calculate scrolling
    const maxLines = height - 4; // Reserve space for title and scroll indicator
    const totalLines = contentLines.length;
    this.maxHelpScrollOffset = Math.max(0, totalLines - maxLines);
    const startIndex = Math.max(0, Math.min(this.helpScrollOffset, this.maxHelpScrollOffset));

    // Display content with scrolling
    let line = 4;
    for (let i = startIndex; i < contentLines.length && line < height - 1; i++) {
      this.moveCursor(line++, startCol);
      this.write(`${contentLines[i]}\x1b[K`);
    }

    // Show scroll indicator if needed
    if (totalLines > maxLines) {
      const scrollProgress = `[${startIndex + 1}-${Math.min(startIndex + maxLines, totalLines)}/${totalLines}]`;
      this.moveCursor(height - 1, width - scrollProgress.length - 1);
      this.write(`\x1b[2m${scrollProgress}\x1b[0m`);
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  /**
   * Get current documentation navigation items (cached during render)
   * This is a workaround since we build navItems during rendering
   */
  private currentDocNavItems: any[] = [];

  private getCurrentDocNavItems(): any[] {
    return this.currentDocNavItems;
  }

  /**
   * Initialize collapsed fields - collapse ALL fields with children by default
   * Optimized to O(n) instead of O(n²)
   */
  private initializeCollapsedFields(): void {
    this.documentationCollapsedFields.clear();

    const selectedFile = this.files[this.selectedIndex];
    if (!selectedFile) return;

    try {
      const filePath = selectedFile.path;
      const content = Deno.readTextFileSync(filePath);
      const parsed = parseHttpFile(content);

      if (parsed.requests.length > 0 && parsed.requests[0].documentation) {
        const doc = parsed.requests[0].documentation;

        // Collapse ALL response fields that have nested children (not just top-level)
        if (doc.responses) {
          for (const response of doc.responses) {
            if (response.fields && response.fields.length > 0) {
              // Build a set of ALL parent paths in the hierarchy (O(n))
              const parentPaths = new Set<string>();

              for (const field of response.fields) {
                // Extract ALL parent paths (not just immediate parent)
                // e.g., "account.characters[].inventory.items[].name" should add:
                //   - account
                //   - account.characters[]
                //   - account.characters[].inventory
                //   - account.characters[].inventory.items[]
                const parts = field.name.split('.');
                for (let i = 1; i < parts.length; i++) {
                  const parentPath = parts.slice(0, i).join('.');
                  parentPaths.add(parentPath);
                }
              }

              // Collapse all parent paths
              for (const parent of parentPaths) {
                this.documentationCollapsedFields.add(parent);
              }
            }
          }
        }
      }
    } catch {
      // Ignore errors
    }
  }

  /**
   * Helper to add response fields to navigation items with collapse support
   */
  private addResponseFieldsToNav(
    fields: any[],
    navItems: any[],
    width: number
  ): void {
    // Build a complete list of all paths (including intermediate parents)
    const allPaths = new Set<string>();
    const fieldMap = new Map<string, any>();

    // Add all actual fields
    for (const field of fields) {
      allPaths.add(field.name);
      fieldMap.set(field.name, field);
    }

    // Add all intermediate parent paths
    for (const field of fields) {
      const parts = field.name.split('.');
      for (let i = 1; i < parts.length; i++) {
        const parentPath = parts.slice(0, i).join('.');
        if (!allPaths.has(parentPath)) {
          allPaths.add(parentPath);
          // Create a virtual parent node
          fieldMap.set(parentPath, {
            name: parentPath,
            type: 'object',
            required: false,
            isVirtual: true  // Mark as virtual parent
          });
        }
      }
    }

    // Convert to array for processing
    const allFields = Array.from(allPaths).map(path => fieldMap.get(path)!);

    // Pre-compute which fields have children (O(n) instead of O(n²))
    const hasChildrenCache = new Map<string, boolean>();
    for (const field of allFields) {
      hasChildrenCache.set(field.name, false);
    }
    for (const field of allFields) {
      const parts = field.name.split('.');
      if (parts.length > 1) {
        const parent = parts.slice(0, -1).join('.');
        hasChildrenCache.set(parent, true);
      }
    }

    // Recursively add fields starting from root
    this.addFieldsRecursive('', allFields, navItems, width, 0, hasChildrenCache);
  }

  /**
   * Recursively add fields with proper indentation and collapse support
   */
  private addFieldsRecursive(
    parentPath: string,
    allFields: any[],
    navItems: any[],
    width: number,
    depth: number,
    hasChildrenCache: Map<string, boolean>
  ): void {
    // Prevent excessive depth
    if (depth > 100) {
      return;
    }

    // Get direct children of this parent
    const children = allFields.filter(f => {
      const parts = f.name.split('.');
      const fieldParent = parts.slice(0, -1).join('.');
      return fieldParent === parentPath;
    });

    for (const field of children) {
      const displayName = field.name.split('.').pop() || field.name;
      const baseIndent = 6 + (depth * 2);
      const indent = " ".repeat(baseIndent);

      // Check if this field has children (use cache)
      const hasChildren = hasChildrenCache.get(field.name) || false;
      const isCollapsed = this.documentationCollapsedFields.has(field.name);

      // Collapse indicator
      const collapseIndicator = hasChildren ? (isCollapsed ? '▶ ' : '▼ ') : '  ';

      // For virtual parent nodes (auto-generated from dot notation), show simpler format
      let fieldText: string;
      if (field.isVirtual) {
        fieldText = `${indent}${collapseIndicator}\x1b[1m${displayName}\x1b[0m`;
      } else {
        const requiredBadge = field.required ? "\x1b[31m[required]\x1b[0m" : "\x1b[33m[optional]\x1b[0m";
        const deprecatedBadge = field.deprecated ? " \x1b[33m[deprecated]\x1b[0m" : "";
        fieldText = `${indent}${collapseIndicator}\x1b[1m${displayName}\x1b[0m \x1b[2m{${field.type}}\x1b[0m ${requiredBadge}${deprecatedBadge}`;
      }

      navItems.push({
        type: 'field',
        text: fieldText,
        fieldPath: field.name,
        hasChildren,
        isCollapsible: hasChildren,
        depth
      });

      // Add description and example if not collapsed (skip for virtual nodes)
      if (!isCollapsed && !field.isVirtual) {
        // Add extra indentation for description/example to align after the collapse indicator (2 chars)
        const textIndent = `${indent}    `; // 4 extra spaces to align after "▶ " or "▼ "

        if (field.description) {
          navItems.push({
            type: 'text',
            text: `${textIndent}${field.description.slice(0, width - 10 - textIndent.length)}`,
            parentField: field.name
          });
        }
        if (field.example !== undefined) {
          const exampleStr = typeof field.example === 'string' ? `"${field.example}"` : JSON.stringify(field.example);
          navItems.push({
            type: 'text',
            text: `${textIndent}\x1b[2mExample: ${exampleStr.slice(0, width - 20 - textIndent.length)}\x1b[0m`,
            parentField: field.name
          });
        }
      }

      // Recursively add children if not collapsed
      if (!isCollapsed && hasChildren) {
        this.addFieldsRecursive(field.name, allFields, navItems, width, depth + 1, hasChildrenCache);
      }
    }
  }

  private drawDocumentation(startCol: number, width: number, height: number): void {
    this.moveCursor(2, startCol);
    const title = " Documentation ";
    this.write(`\x1b[1m${title}\x1b[0m\x1b[K`);

    // Get currently selected file and parse it
    const selectedFile = this.files[this.selectedIndex];
    if (!selectedFile) {
      this.moveCursor(4, startCol);
      this.write(`\x1b[2mNo file selected\x1b[0m\x1b[K`);
      return;
    }

    // Parse the file to get documentation
    let documentation: Documentation | undefined;
    try {
      const filePath = selectedFile.path;
      const content = Deno.readTextFileSync(filePath);
      const parsed = parseHttpFile(content);

      if (parsed.requests.length > 0) {
        documentation = parsed.requests[0].documentation;
      }
    } catch (error) {
      this.moveCursor(4, startCol);
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.write(`\x1b[31mError loading documentation: ${errorMsg}\x1b[0m\x1b[K`);
      return;
    }

    // Check if documentation has any content
    const hasContent = documentation && (
      documentation.description ||
      (documentation.tags && documentation.tags.length > 0) ||
      (documentation.parameters && documentation.parameters.length > 0) ||
      (documentation.responses && documentation.responses.length > 0)
    );

    if (!hasContent) {
      this.moveCursor(4, startCol);
      this.write(`\x1b[2mNo documentation available for this request\x1b[0m\x1b[K`);
      this.moveCursor(6, startCol);
      this.write(`\x1b[2mAdd documentation using:\x1b[0m\x1b[K`);
      this.moveCursor(7, startCol);
      this.write(`\x1b[2m  • # @description ... in .http files\x1b[0m\x1b[K`);
      this.moveCursor(8, startCol);
      this.write(`\x1b[2m  • documentation: section in .yaml files\x1b[0m\x1b[K`);
      this.moveCursor(10, startCol);
      this.write(`\x1b[2mSee docs/DOCUMENTATION.md for details\x1b[0m\x1b[K`);
      return;
    }

    // Build navigable items with collapse support
    interface NavItem {
      type: 'text' | 'field' | 'header';
      text: string;
      fieldPath?: string;  // For collapsible fields
      hasChildren?: boolean;  // Whether this field has nested children
      isCollapsible?: boolean;  // Whether this item can be collapsed
    }

    const navItems: NavItem[] = [];
    const doc = documentation!;

    // Description
    if (doc.description) {
      navItems.push({ type: 'header', text: `\x1b[1;36mDescription:\x1b[0m` });
      navItems.push({ type: 'text', text: "" });
      const maxDescWidth = width - 4;
      const words = doc.description.split(' ');
      let currentLine = "  ";
      for (const word of words) {
        if (currentLine.length + word.length + 1 > maxDescWidth) {
          navItems.push({ type: 'text', text: currentLine });
          currentLine = "  " + word;
        } else {
          currentLine += (currentLine.length > 2 ? " " : "") + word;
        }
      }
      if (currentLine.length > 2) {
        navItems.push({ type: 'text', text: currentLine });
      }
      navItems.push({ type: 'text', text: "" });
    }

    // Tags
    if (doc.tags && doc.tags.length > 0) {
      navItems.push({ type: 'header', text: `\x1b[1;36mTags:\x1b[0m` });
      navItems.push({ type: 'text', text: "" });
      navItems.push({ type: 'text', text: `  ${doc.tags.map(t => `\x1b[35m#${t}\x1b[0m`).join("  ")}` });
      navItems.push({ type: 'text', text: "" });
    }

    // Parameters
    if (doc.parameters && doc.parameters.length > 0) {
      navItems.push({ type: 'header', text: `\x1b[1;36mParameters:\x1b[0m` });
      navItems.push({ type: 'text', text: "" });
      for (const param of doc.parameters) {
        const requiredBadge = param.required ? "\x1b[31m[required]\x1b[0m" : "\x1b[33m[optional]\x1b[0m";
        navItems.push({ type: 'field', text: `  \x1b[1m${param.name}\x1b[0m \x1b[2m{${param.type}}\x1b[0m ${requiredBadge}`, fieldPath: `param.${param.name}`, isCollapsible: false });
        if (param.description) {
          navItems.push({ type: 'text', text: `    ${param.description.slice(0, width - 6)}` });
        }
        if (param.example !== undefined) {
          const exampleStr = typeof param.example === 'string' ? `"${param.example}"` : String(param.example);
          navItems.push({ type: 'text', text: `    \x1b[2mExample: ${exampleStr.slice(0, width - 16)}\x1b[0m` });
        }
        navItems.push({ type: 'text', text: "" });
      }
    }

    // Responses with collapsible fields
    if (doc.responses && doc.responses.length > 0) {
      navItems.push({ type: 'header', text: `\x1b[1;36mResponses:\x1b[0m` });
      navItems.push({ type: 'text', text: "" });
      for (const response of doc.responses) {
        const codeColor = response.code.startsWith('2') ? '\x1b[32m' :
                          response.code.startsWith('4') || response.code.startsWith('5') ? '\x1b[31m' : '\x1b[33m';
        navItems.push({ type: 'field', text: `  ${codeColor}${response.code}\x1b[0m  ${response.description.slice(0, width - 10)}`, fieldPath: `response.${response.code}`, isCollapsible: false });

        if (response.fields && response.fields.length > 0) {
          navItems.push({ type: 'text', text: "" });
          navItems.push({ type: 'text', text: `    \x1b[2mResponse Body:\x1b[0m` });

          // Build field tree and collapse by default
          this.addResponseFieldsToNav(response.fields, navItems, width);
        }
        navItems.push({ type: 'text', text: "" });
      }
    }

    // Footer
    navItems.push({ type: 'text', text: "" });
    navItems.push({ type: 'text', text: "\x1b[2mPress ESC/m to close | ↑/↓/PgUp/PgDn to navigate | Space to expand/collapse\x1b[0m" });

    // Cache navItems for keyboard handlers
    this.currentDocNavItems = navItems;

    this.documentationMaxCursorIndex = navItems.filter(item => item.type === 'field').length - 1;

    // Ensure cursor is in valid range
    if (this.documentationCursorIndex > this.documentationMaxCursorIndex) {
      this.documentationCursorIndex = Math.max(0, this.documentationMaxCursorIndex);
    }

    // Calculate which field is under the cursor
    const fieldItems = navItems.filter(item => item.type === 'field');
    const cursorField = fieldItems[this.documentationCursorIndex];

    // Auto-scroll to keep cursor visible
    const maxLines = height - 4;
    const cursorItemIndex = navItems.indexOf(cursorField);

    if (cursorItemIndex !== -1) {
      // Scroll down if cursor is below visible area
      if (cursorItemIndex >= this.documentationScrollOffset + maxLines) {
        this.documentationScrollOffset = cursorItemIndex - maxLines + 1;
      }
      // Scroll up if cursor is above visible area
      if (cursorItemIndex < this.documentationScrollOffset) {
        this.documentationScrollOffset = cursorItemIndex;
      }
    }

    const totalLines = navItems.length;
    this.maxDocumentationScrollOffset = Math.max(0, totalLines - maxLines);
    const startIndex = Math.max(0, Math.min(this.documentationScrollOffset, this.maxDocumentationScrollOffset));

    // Display content with scrolling and cursor highlighting
    let line = 4;
    for (let i = startIndex; i < navItems.length && line < height - 1; i++) {
      this.moveCursor(line++, startCol);

      const item = navItems[i];
      const isCursor = item === cursorField;

      // Highlight cursor line
      if (isCursor) {
        this.write(`\x1b[7m${item.text}\x1b[0m\x1b[K`); // Reverse video
      } else {
        this.write(`${item.text}\x1b[K`);
      }
    }

    // Show scroll indicator if needed
    if (totalLines > maxLines) {
      const scrollProgress = `[${startIndex + 1}-${Math.min(startIndex + maxLines, totalLines)}/${totalLines}]`;
      this.moveCursor(height - 1, startCol + width - scrollProgress.length - 1);
      this.write(`\x1b[2m${scrollProgress}\x1b[0m`);
    }

    // Clear remaining lines
    while (line <= height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawStatusBar(width: number, row: number): void {
    this.moveCursor(row, 1);

    let statusText: string;

    if (this.variableMode) {
      if (this.variableEditMode === "list") {
        statusText =
          " [↑↓] Navigate | [A] Add | [E/Enter] Edit | [D] Delete | [ESC] Exit ";
      } else if (this.variableEditMode === "add") {
        statusText =
          " [Tab] Switch field | [Ctrl+K] Clear | [Opt+Del] Del word | [Enter] Save | [ESC] Cancel ";
      } else if (this.variableEditMode === "edit") {
        statusText =
          " [Ctrl+K] Clear all | [Opt+Del] Del word | [Enter] Save | [ESC] Cancel ";
      } else if (this.variableEditMode === "delete") {
        statusText = " [Y] Confirm Delete | [N] Cancel ";
      } else {
        statusText = " Variable Editor ";
      }
    } else if (this.headerMode) {
      if (this.headerEditMode === "list") {
        statusText =
          " [↑↓] Navigate | [A] Add | [E/Enter] Edit | [D] Delete | [ESC] Exit ";
      } else if (this.headerEditMode === "add") {
        statusText =
          " [Tab] Switch field | [Ctrl+K] Clear | [Opt+Del] Del word | [Enter] Save | [ESC] Cancel ";
      } else if (this.headerEditMode === "edit") {
        statusText =
          " [Ctrl+K] Clear all | [Opt+Del] Del word | [Enter] Save | [ESC] Cancel ";
      } else if (this.headerEditMode === "delete") {
        statusText = " [Y] Confirm Delete | [N] Cancel ";
      } else {
        statusText = " Header Editor ";
      }
    } else if (this.documentationMode) {
      statusText = " [↑↓/PgUp/PgDn] Navigate | [Space] Expand/Collapse | [m/ESC] Close ";
    } else if (this.historyMode) {
      const count = this.historyEntries.length;
      if (count === 0) {
        statusText = " [ESC] Exit History ";
      } else {
        statusText =
          ` [↑↓] Navigate ${count} entries | [Enter] View response | [ESC] Exit `;
      }
    } else if (this.gotoMode) {
      statusText =
        ` Go to (hex): :${this.gotoQuery}_ | [Enter] Jump | [ESC] Cancel `;
    } else if (this.searchMode) {
      const matchCount = this.searchResults.length;
      const currentMatch = matchCount > 0 ? this.searchResultIndex + 1 : 0;
      statusText =
        ` Search: ${this.searchQuery}_ | ${currentMatch}/${matchCount} matches | [Ctrl+R] Next | [ESC] Cancel | [Enter] Select `;
    } else {
      const fullscreenHint = this.fullscreenMode ? " [FULLSCREEN] " : "";
      const help =
        `${fullscreenHint}[↑↓] Nav | [Enter] Execute | [i] Inspect | [f] Fullscreen | [v] Vars | [h] Headers | [Ctrl+H] History | [d] Dup | [p] Profile | [q] Quit `;
      statusText = this.statusMessage || help;
    }

    const padding = " ".repeat(Math.max(0, width - statusText.length));
    this.write(`\x1b[7m${statusText}${padding}\x1b[0m`);
  }

  async handleInput(): Promise<void> {
    // Larger buffer to handle paste operations (e.g., JWT tokens)
    const buf = new Uint8Array(4096);

    while (this.running) {
      const n = await Deno.stdin.read(buf);
      if (!n) continue;

      const input = buf.subarray(0, n);

      // Handle help mode
      if (this.helpMode) {
        // ESC or ? to close help
        if (input.length === 1 && (input[0] === 27 || input[0] === 63)) {
          this.helpMode = false;
          this.helpScrollOffset = 0; // Reset scroll
          this.draw();
          continue;
        }

        // Arrow keys for scrolling
        if (input.length === 3 && input[0] === 27 && input[1] === 91) {
          if (input[2] === 65) {
            // Up
            this.helpScrollOffset = Math.max(0, this.helpScrollOffset - 1);
            this.draw();
          } else if (input[2] === 66) {
            // Down
            this.helpScrollOffset = Math.min(
              this.maxHelpScrollOffset,
              this.helpScrollOffset + 1,
            );
            this.draw();
          }
        }
        continue;
      }

      // Handle documentation mode
      if (this.documentationMode) {
        // ESC or m to close documentation
        if (input.length === 1 && (input[0] === 27 || input[0] === 109)) {
          this.documentationMode = false;
          this.documentationScrollOffset = 0;
          this.documentationCursorIndex = 0;
          this.documentationCollapsedFields.clear();
          this.draw();
          continue;
        }

        // Space to toggle collapse/expand
        if (input.length === 1 && input[0] === 32) {
          // Get current field under cursor
          const navItems = this.getCurrentDocNavItems();
          const fieldItems = navItems.filter((item: any) => item.type === 'field');
          const cursorField = fieldItems[this.documentationCursorIndex];

          if (cursorField && cursorField.isCollapsible) {
            if (this.documentationCollapsedFields.has(cursorField.fieldPath)) {
              this.documentationCollapsedFields.delete(cursorField.fieldPath);
            } else {
              this.documentationCollapsedFields.add(cursorField.fieldPath);
            }
            this.draw();
          }
          continue;
        }

        // Arrow keys for navigation
        if (input.length === 3 && input[0] === 27 && input[1] === 91) {
          if (input[2] === 65) {
            // Up - move cursor up
            this.documentationCursorIndex = Math.max(0, this.documentationCursorIndex - 1);
            this.draw();
          } else if (input[2] === 66) {
            // Down - move cursor down
            this.documentationCursorIndex = Math.min(
              this.documentationMaxCursorIndex,
              this.documentationCursorIndex + 1,
            );
            this.draw();
          }
        }

        // Page Up/Down without the tilde (some terminals)
        if (input.length === 4 && input[0] === 27 && input[1] === 91) {
          if (input[2] === 53 && input[3] === 126) {
            // Page Up
            this.documentationCursorIndex = Math.max(0, this.documentationCursorIndex - 10);
            this.draw();
          } else if (input[2] === 54 && input[3] === 126) {
            // Page Down
            this.documentationCursorIndex = Math.min(
              this.documentationMaxCursorIndex,
              this.documentationCursorIndex + 10,
            );
            this.draw();
          }
        }
        continue;
      }

      // Handle variable mode
      if (this.variableMode) {
        await this.handleVariableInput(input);
        continue;
      }

      // Handle header mode
      if (this.headerMode) {
        await this.handleHeaderInput(input);
        continue;
      }

      // Handle history mode
      if (this.historyMode) {
        await this.handleHistoryInput(input);
        continue;
      }

      // Handle goto mode
      if (this.gotoMode) {
        await this.handleGotoInput(input);
        continue;
      }

      // Handle search mode
      if (this.searchMode) {
        await this.handleSearchInput(input);
        continue;
      }

      // Check for : (start goto)
      if (input.length === 1 && input[0] === 58) {
        this.enterGotoMode();
        continue;
      }

      // Check for Ctrl+R (start search)
      if (input.length === 1 && input[0] === 18) {
        this.enterSearchMode();
        continue;
      }

      // Check for Ctrl+H (open history)
      if (input.length === 1 && input[0] === 8) {
        await this.enterHistoryMode();
        continue;
      }

      // Check for ESC key (clear status message)
      if (input.length === 1 && input[0] === 27) {
        this.statusMessage = "";
        this.draw();
        continue;
      }

      // Check for special keys
      if (input.length === 3 && input[0] === 27 && input[1] === 91) {
        // Arrow keys (circular)
        if (input[2] === 65) {
          // Up
          if (this.selectedIndex === 0) {
            this.selectedIndex = this.files.length - 1; // Wrap to bottom
          } else {
            this.selectedIndex--;
          }
          this.responseScrollOffset = 0; // Reset scroll when changing files
          this.draw();
        } else if (input[2] === 66) {
          // Down
          if (this.selectedIndex === this.files.length - 1) {
            this.selectedIndex = 0; // Wrap to top
          } else {
            this.selectedIndex++;
          }
          this.responseScrollOffset = 0; // Reset scroll when changing files
          this.draw();
        }
      } else if (input.length === 4 && input[0] === 27 && input[1] === 91) {
        // Page Up/Down (ESC[5~ and ESC[6~)
        if (input[2] === 53 && input[3] === 126) {
          // Page Up
          this.selectedIndex = Math.max(0, this.selectedIndex - this.pageSize);
          this.draw();
        } else if (input[2] === 54 && input[3] === 126) {
          // Page Down
          this.selectedIndex = Math.min(
            this.files.length - 1,
            this.selectedIndex + this.pageSize,
          );
          this.draw();
        }
      } else if (input.length === 1) {
        const char = String.fromCharCode(input[0]);

        if (char === "q") {
          this.running = false;
        } else if (char === "\r") {
          // Enter
          await this.executeSelected();
        } else if (char === "d") {
          await this.duplicateSelected();
        } else if (char === "p") {
          await this.selectProfile();
        } else if (char === "s") {
          await this.saveResponse();
        } else if (char === "c") {
          await this.copyResponse();
        } else if (char === "r") {
          await this.refreshFiles();
        } else if (char === "i") {
          await this.inspectRequest();
        } else if (char === "v") {
          this.enterVariableMode();
        } else if (char === "h") {
          this.enterHeaderMode();
        } else if (char === "f") {
          this.toggleFullscreen();
        } else if (char === "j") {
          // Scroll response down
          if (this.response && this.response.body) {
            this.responseScrollOffset = Math.min(
              this.maxResponseScrollOffset,
              this.responseScrollOffset + 1,
            );
            this.draw();
          }
        } else if (char === "k") {
          // Scroll response up
          if (this.response && this.response.body) {
            this.responseScrollOffset = Math.max(
              0,
              this.responseScrollOffset - 1,
            );
            this.draw();
          }
        } else if (char === "H") {
          // Toggle response headers visibility
          this.showResponseHeaders = !this.showResponseHeaders;
          this.draw();
        } else if (char === "B") {
          // Toggle response body visibility
          this.showResponseBody = !this.showResponseBody;
          this.draw();
        } else if (char === "?") {
          // Toggle help modal
          this.helpMode = !this.helpMode;
          if (this.helpMode) {
            this.helpScrollOffset = 0; // Reset scroll when opening
          }
          this.draw();
        } else if (char === "m") {
          // Toggle documentation panel
          this.documentationMode = !this.documentationMode;
          if (this.documentationMode) {
            this.documentationScrollOffset = 0;
            this.documentationCursorIndex = 0;
            // Initialize collapsed fields - collapse all by default
            this.initializeCollapsedFields();
          }
          this.draw();
        }
      }
    }
  }

  async handleSearchInput(input: Uint8Array): Promise<void> {
    // Ctrl+R - cycle through results
    if (input.length === 1 && input[0] === 18) {
      if (this.searchResults.length > 0) {
        this.searchResultIndex = (this.searchResultIndex + 1) %
          this.searchResults.length;
        this.selectedIndex = this.searchResults[this.searchResultIndex];
        this.draw();
      }
      return;
    }

    // ESC - exit search
    if (input.length === 1 && input[0] === 27) {
      this.exitSearchMode();
      return;
    }

    // Enter - select and exit search
    if (input.length === 1 && input[0] === 13) {
      this.exitSearchMode();
      return;
    }

    // Backspace - remove character
    if (input.length === 1 && input[0] === 127) {
      if (this.searchQuery.length > 0) {
        this.searchQuery = this.searchQuery.slice(0, -1);
        this.updateSearchResults();
        this.draw();
      }
      return;
    }

    // Printable characters - add to query
    if (input.length === 1 && input[0] >= 32 && input[0] <= 126) {
      const char = String.fromCharCode(input[0]);
      this.searchQuery += char;
      this.updateSearchResults();
      this.draw();
    }
  }

  enterSearchMode(): void {
    this.searchMode = true;
    this.searchQuery = "";
    this.searchResults = [];
    this.searchResultIndex = 0;
    this.draw();
  }

  exitSearchMode(): void {
    this.searchMode = false;
    this.searchQuery = "";
    this.searchResults = [];
    this.searchResultIndex = 0;
    this.draw();
  }

  updateSearchResults(): void {
    const query = this.searchQuery.toLowerCase();
    this.searchResults = [];

    if (query === "") {
      return;
    }

    for (let i = 0; i < this.files.length; i++) {
      const fileName = this.files[i].name.toLowerCase();
      if (fileName.includes(query)) {
        this.searchResults.push(i);
      }
    }

    // Select first result if any
    if (this.searchResults.length > 0) {
      this.searchResultIndex = 0;
      this.selectedIndex = this.searchResults[0];
    }
  }

  async handleGotoInput(input: Uint8Array): Promise<void> {
    // ESC - exit goto
    if (input.length === 1 && input[0] === 27) {
      this.exitGotoMode();
      return;
    }

    // Enter - jump to line (parse as hex)
    if (input.length === 1 && input[0] === 13) {
      const lineNum = parseInt(this.gotoQuery, 16);
      if (!isNaN(lineNum) && lineNum >= 1 && lineNum <= this.files.length) {
        this.selectedIndex = lineNum - 1;
      }
      this.exitGotoMode();
      return;
    }

    // Backspace - remove character
    if (input.length === 1 && input[0] === 127) {
      if (this.gotoQuery.length > 0) {
        this.gotoQuery = this.gotoQuery.slice(0, -1);
        this.draw();
      }
      return;
    }

    // Hex characters: 0-9, A-F, a-f
    if (input.length === 1) {
      const char = String.fromCharCode(input[0]);
      if (
        (input[0] >= 48 && input[0] <= 57) || // 0-9
        (input[0] >= 65 && input[0] <= 70) || // A-F
        (input[0] >= 97 && input[0] <= 102)
      ) { // a-f
        this.gotoQuery += char.toUpperCase();
        this.draw();
      }
    }
  }

  enterGotoMode(): void {
    this.gotoMode = true;
    this.gotoQuery = "";
    this.draw();
  }

  exitGotoMode(): void {
    this.gotoMode = false;
    this.gotoQuery = "";
    this.draw();
  }

  enterVariableMode(): void {
    this.variableMode = true;
    this.variableEditMode = "list";
    this.variableIndex = 0;
    this.variableEditKey = "";
    this.variableEditValue = "";
    this.variableEditField = "key";
    this.draw();
  }

  exitVariableMode(): void {
    this.variableMode = false;
    this.variableEditMode = "list";
    this.variableIndex = 0;
    this.variableEditKey = "";
    this.variableEditValue = "";
    this.draw();
  }

  async handleVariableInput(input: Uint8Array): Promise<void> {
    // ESC - exit variable mode
    if (input.length === 1 && input[0] === 27) {
      this.exitVariableMode();
      return;
    }

    // In list mode
    if (this.variableEditMode === "list") {
      const variables = this.sessionManager.getProfileVariables();
      const varEntries = Object.entries(variables);

      // Arrow keys
      if (input.length === 3 && input[0] === 27 && input[1] === 91) {
        if (input[2] === 65) {
          // Up
          this.variableIndex = Math.max(0, this.variableIndex - 1);
          this.draw();
        } else if (input[2] === 66) {
          // Down
          this.variableIndex = Math.min(
            varEntries.length,
            this.variableIndex + 1,
          );
          this.draw();
        }
        return;
      }

      if (input.length === 1) {
        const char = String.fromCharCode(input[0]);

        if (char === "a" || char === "A") {
          // Add new variable
          this.variableEditMode = "add";
          this.variableEditKey = "";
          this.variableEditValue = "";
          this.variableEditField = "key";
          this.draw();
        } else if (char === "e" || char === "E" || char === "\r") {
          // Edit selected variable
          if (this.variableIndex < varEntries.length) {
            this.variableEditMode = "edit";
            const [key, value] = varEntries[this.variableIndex];
            this.variableEditKey = key;
            this.variableEditValue = value;
            this.variableEditField = "value";
            this.draw();
          }
        } else if (char === "d" || char === "D") {
          // Delete selected variable
          if (this.variableIndex < varEntries.length) {
            this.variableEditMode = "delete";
            const [key] = varEntries[this.variableIndex];
            this.variableEditKey = key;
            this.draw();
          }
        }
      }
    } // In add/edit mode
    else if (
      this.variableEditMode === "add" || this.variableEditMode === "edit"
    ) {
      // Ctrl+K - clear entire value
      if (input.length === 1 && input[0] === 11) {
        if (this.variableEditMode === "edit") {
          this.variableEditValue = "";
          this.draw();
        } else if (this.variableEditField === "key") {
          this.variableEditKey = "";
          this.draw();
        } else if (this.variableEditField === "value") {
          this.variableEditValue = "";
          this.draw();
        }
        return;
      }

      // Option+Delete (macOS) or Alt+Backspace - delete previous word
      if (input.length === 2 && input[0] === 27 && input[1] === 127) {
        if (this.variableEditMode === "edit") {
          this.variableEditValue = this.deleteLastWord(this.variableEditValue);
          this.draw();
        } else if (this.variableEditField === "key") {
          this.variableEditKey = this.deleteLastWord(this.variableEditKey);
          this.draw();
        } else if (this.variableEditField === "value") {
          this.variableEditValue = this.deleteLastWord(this.variableEditValue);
          this.draw();
        }
        return;
      }

      // Tab - switch between key and value fields (only in add mode)
      if (input.length === 1 && input[0] === 9) {
        if (this.variableEditMode === "add") {
          this.variableEditField = this.variableEditField === "key"
            ? "value"
            : "key";
          this.draw();
        }
        return;
      }

      // Enter - save
      if (input.length === 1 && input[0] === 13) {
        if (this.variableEditKey.trim()) {
          this.sessionManager.setProfileVariable(
            this.variableEditKey.trim(),
            this.variableEditValue,
          );
          await this.sessionManager.saveProfiles();
          this.statusMessage =
            ` Variable '${this.variableEditKey}' saved to profile `;
          this.variableEditMode = "list";
          this.draw();
        }
        return;
      }

      // Backspace
      if (input.length === 1 && input[0] === 127) {
        // In edit mode, only allow editing value
        if (this.variableEditMode === "edit") {
          if (this.variableEditValue.length > 0) {
            this.variableEditValue = this.variableEditValue.slice(0, -1);
            this.draw();
          }
        } else {
          // In add mode, respect the current field
          if (
            this.variableEditField === "key" && this.variableEditKey.length > 0
          ) {
            this.variableEditKey = this.variableEditKey.slice(0, -1);
            this.draw();
          } else if (
            this.variableEditField === "value" &&
            this.variableEditValue.length > 0
          ) {
            this.variableEditValue = this.variableEditValue.slice(0, -1);
            this.draw();
          }
        }
        return;
      }

      // Printable characters (handles paste - multiple chars at once)
      let hasValidChars = false;
      for (let i = 0; i < input.length; i++) {
        if (input[i] >= 32 && input[i] <= 126) {
          const char = String.fromCharCode(input[i]);
          hasValidChars = true;

          // In edit mode, only allow editing value
          if (this.variableEditMode === "edit") {
            this.variableEditValue += char;
          } else {
            // In add mode, respect the current field
            if (this.variableEditField === "key") {
              this.variableEditKey += char;
            } else {
              this.variableEditValue += char;
            }
          }
        }
      }
      if (hasValidChars) {
        this.draw();
      }
    } // In delete confirmation mode
    else if (this.variableEditMode === "delete") {
      if (input.length === 1) {
        const char = String.fromCharCode(input[0]).toLowerCase();

        if (char === "y") {
          // Confirm delete
          this.sessionManager.deleteProfileVariable(this.variableEditKey);
          await this.sessionManager.saveProfiles();
          this.statusMessage =
            ` Variable '${this.variableEditKey}' deleted from profile `;
          this.variableEditMode = "list";
          this.variableIndex = Math.max(0, this.variableIndex - 1);
          this.draw();
        } else if (char === "n" || char === "\r") {
          // Cancel delete
          this.variableEditMode = "list";
          this.draw();
        }
      }
    }
  }

  enterHeaderMode(): void {
    this.headerMode = true;
    this.headerEditMode = "list";
    this.headerIndex = 0;
    this.headerEditKey = "";
    this.headerEditValue = "";
    this.headerEditField = "key";
    this.draw();
  }

  exitHeaderMode(): void {
    this.headerMode = false;
    this.headerEditMode = "list";
    this.headerIndex = 0;
    this.headerEditKey = "";
    this.headerEditValue = "";
    this.draw();
  }

  async enterHistoryMode(): Promise<void> {
    if (this.selectedIndex >= this.files.length) {
      this.statusMessage = " No file selected ";
      this.draw();
      return;
    }

    const file = this.files[this.selectedIndex];
    this.statusMessage = ` Loading history for ${file.name}... `;
    this.draw();

    try {
      this.historyEntries = await this.historyManager.getHistory(file.path);
      this.historyMode = true;
      this.historyIndex = 0;
      this.statusMessage = "";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error loading history: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  exitHistoryMode(): void {
    this.historyMode = false;
    this.historyEntries = [];
    this.historyIndex = 0;
    this.draw();
  }

  toggleFullscreen(): void {
    this.fullscreenMode = !this.fullscreenMode;
    this.statusMessage = this.fullscreenMode
      ? " Fullscreen mode ON - Press [f] to exit "
      : " Fullscreen mode OFF ";
    this.draw();
  }

  async handleHistoryInput(input: Uint8Array): Promise<void> {
    // ESC - exit history mode
    if (input.length === 1 && input[0] === 27) {
      this.exitHistoryMode();
      return;
    }

    // Arrow keys
    if (input.length === 3 && input[0] === 27 && input[1] === 91) {
      if (input[2] === 65) {
        // Up
        this.historyIndex = Math.max(0, this.historyIndex - 1);
        this.draw();
      } else if (input[2] === 66) {
        // Down
        this.historyIndex = Math.min(
          this.historyEntries.length - 1,
          this.historyIndex + 1,
        );
        this.draw();
      }
      return;
    }

    // Enter - view selected history entry
    if (input.length === 1 && input[0] === 13) {
      if (this.historyIndex < this.historyEntries.length) {
        const entry = this.historyEntries[this.historyIndex];
        // Convert history entry to RequestResult format to display it
        this.response = {
          status: entry.responseStatus,
          statusText: entry.responseStatusText,
          headers: entry.responseHeaders,
          body: this.beautifyJson(entry.responseBody),
          duration: entry.duration,
          requestSize: entry.requestSize || 0,
          responseSize: entry.responseSize || 0,
          error: entry.error,
        };
        this.responseScrollOffset = 0; // Reset scroll for new response
        this.exitHistoryMode();
      }
      return;
    }
  }

  async handleHeaderInput(input: Uint8Array): Promise<void> {
    // ESC - exit header mode
    if (input.length === 1 && input[0] === 27) {
      this.exitHeaderMode();
      return;
    }

    // In list mode
    if (this.headerEditMode === "list") {
      const headers = this.sessionManager.getProfileHeaders();
      const headerEntries = Object.entries(headers);

      // Arrow keys
      if (input.length === 3 && input[0] === 27 && input[1] === 91) {
        if (input[2] === 65) {
          // Up
          this.headerIndex = Math.max(0, this.headerIndex - 1);
          this.draw();
        } else if (input[2] === 66) {
          // Down
          this.headerIndex = Math.min(
            headerEntries.length,
            this.headerIndex + 1,
          );
          this.draw();
        }
        return;
      }

      if (input.length === 1) {
        const char = String.fromCharCode(input[0]);

        if (char === "a" || char === "A") {
          // Add new header
          this.headerEditMode = "add";
          this.headerEditKey = "";
          this.headerEditValue = "";
          this.headerEditField = "key";
          this.draw();
        } else if (char === "e" || char === "E" || char === "\r") {
          // Edit selected header
          if (this.headerIndex < headerEntries.length) {
            this.headerEditMode = "edit";
            const [key, value] = headerEntries[this.headerIndex];
            this.headerEditKey = key;
            this.headerEditValue = value;
            this.headerEditField = "value";
            this.draw();
          }
        } else if (char === "d" || char === "D") {
          // Delete selected header
          if (this.headerIndex < headerEntries.length) {
            this.headerEditMode = "delete";
            const [key] = headerEntries[this.headerIndex];
            this.headerEditKey = key;
            this.draw();
          }
        }
      }
    } // In add/edit mode
    else if (this.headerEditMode === "add" || this.headerEditMode === "edit") {
      // Ctrl+K - clear entire value
      if (input.length === 1 && input[0] === 11) {
        if (this.headerEditMode === "edit") {
          this.headerEditValue = "";
          this.draw();
        } else if (this.headerEditField === "key") {
          this.headerEditKey = "";
          this.draw();
        } else if (this.headerEditField === "value") {
          this.headerEditValue = "";
          this.draw();
        }
        return;
      }

      // Option+Delete (macOS) or Alt+Backspace - delete previous word
      if (input.length === 2 && input[0] === 27 && input[1] === 127) {
        if (this.headerEditMode === "edit") {
          this.headerEditValue = this.deleteLastWord(this.headerEditValue);
          this.draw();
        } else if (this.headerEditField === "key") {
          this.headerEditKey = this.deleteLastWord(this.headerEditKey);
          this.draw();
        } else if (this.headerEditField === "value") {
          this.headerEditValue = this.deleteLastWord(this.headerEditValue);
          this.draw();
        }
        return;
      }

      // Tab - switch between key and value fields (only in add mode)
      if (input.length === 1 && input[0] === 9) {
        if (this.headerEditMode === "add") {
          this.headerEditField = this.headerEditField === "key"
            ? "value"
            : "key";
          this.draw();
        }
        return;
      }

      // Enter - save
      if (input.length === 1 && input[0] === 13) {
        if (this.headerEditKey.trim()) {
          this.sessionManager.setProfileHeader(
            this.headerEditKey.trim(),
            this.headerEditValue,
          );
          await this.sessionManager.saveProfiles();
          this.statusMessage =
            ` Header '${this.headerEditKey}' saved to profile `;
          this.headerEditMode = "list";
          this.draw();
        }
        return;
      }

      // Backspace
      if (input.length === 1 && input[0] === 127) {
        // In edit mode, only allow editing value
        if (this.headerEditMode === "edit") {
          if (this.headerEditValue.length > 0) {
            this.headerEditValue = this.headerEditValue.slice(0, -1);
            this.draw();
          }
        } else {
          // In add mode, respect the current field
          if (this.headerEditField === "key" && this.headerEditKey.length > 0) {
            this.headerEditKey = this.headerEditKey.slice(0, -1);
            this.draw();
          } else if (
            this.headerEditField === "value" && this.headerEditValue.length > 0
          ) {
            this.headerEditValue = this.headerEditValue.slice(0, -1);
            this.draw();
          }
        }
        return;
      }

      // Printable characters (handles paste - multiple chars at once)
      let hasValidChars = false;
      for (let i = 0; i < input.length; i++) {
        if (input[i] >= 32 && input[i] <= 126) {
          const char = String.fromCharCode(input[i]);
          hasValidChars = true;

          // In edit mode, only allow editing value
          if (this.headerEditMode === "edit") {
            this.headerEditValue += char;
          } else {
            // In add mode, respect the current field
            if (this.headerEditField === "key") {
              this.headerEditKey += char;
            } else {
              this.headerEditValue += char;
            }
          }
        }
      }
      if (hasValidChars) {
        this.draw();
      }
    } // In delete confirmation mode
    else if (this.headerEditMode === "delete") {
      if (input.length === 1) {
        const char = String.fromCharCode(input[0]).toLowerCase();

        if (char === "y") {
          // Confirm delete
          this.sessionManager.deleteProfileHeader(this.headerEditKey);
          await this.sessionManager.saveProfiles();
          this.statusMessage =
            ` Header '${this.headerEditKey}' deleted from profile `;
          this.headerEditMode = "list";
          this.headerIndex = Math.max(0, this.headerIndex - 1);
          this.draw();
        } else if (char === "n" || char === "\r") {
          // Cancel delete
          this.headerEditMode = "list";
          this.draw();
        }
      }
    }
  }

  async executeSelected(): Promise<void> {
    if (this.selectedIndex >= this.files.length) return;

    const file = this.files[this.selectedIndex];
    this.statusMessage = ` Executing ${file.name}... `;
    this.draw();

    try {
      const content = await Deno.readTextFile(file.path);
      const parsed = parseHttpFile(content);

      if (parsed.requests.length === 0) {
        this.statusMessage = " No requests found in file ";
        this.draw();
        return;
      }

      // Execute first request
      const request = parsed.requests[0];
      const variables = this.sessionManager.getVariables();
      const profileHeaders = this.sessionManager.getActiveHeaders();

      this.response = await this.executor.execute(
        request,
        variables,
        profileHeaders,
      );

      // Beautify JSON response body
      if (this.response && this.response.body) {
        this.response.body = this.beautifyJson(this.response.body);
      }

      // Reset scroll offset for new response
      this.responseScrollOffset = 0;

      // Save to history if enabled
      if (this.sessionManager.isHistoryEnabled()) {
        try {
          const substituted = applyVariables(request, variables);
          const mergedHeaders = { ...profileHeaders, ...substituted.headers };

          await this.historyManager.save({
            timestamp: new Date().toISOString(),
            requestFile: file.path,
            requestName: request.name,
            method: substituted.method,
            url: substituted.url,
            headers: mergedHeaders,
            body: substituted.body,
            responseStatus: this.response.status,
            responseStatusText: this.response.statusText,
            responseHeaders: this.response.headers,
            responseBody: this.response.body,
            duration: this.response.duration,
            requestSize: this.response.requestSize,
            responseSize: this.response.responseSize,
            error: this.response.error,
          });
        } catch (historyError) {
          // Don't fail the request if history save fails
          console.error("Failed to save history:", historyError);
        }
      }

      // Save response body variables if it's JSON
      try {
        const json = JSON.parse(this.response.body);
        if (json.token) this.sessionManager.setVariable("token", json.token);
        if (json.accessToken) {
          this.sessionManager.setVariable("token", json.accessToken);
        }
        await this.sessionManager.save();
      } catch {
        // Not JSON or no token
      }

      this.statusMessage = "";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async duplicateSelected(): Promise<void> {
    if (this.selectedIndex >= this.files.length) return;

    const file = this.files[this.selectedIndex];
    const ext = path.extname(file.path);
    const base = file.path.slice(0, -ext.length);

    let counter = 1;
    let newPath = `${base}-copy${ext}`;

    while (await Deno.stat(newPath).catch(() => null)) {
      counter++;
      newPath = `${base}-copy${counter}${ext}`;
    }

    try {
      const content = await Deno.readTextFile(file.path);
      await Deno.writeTextFile(newPath, content);
      await this.loadFiles();
      this.selectedIndex = this.files.findIndex((f) => f.path === newPath);
      this.statusMessage = ` Duplicated to ${path.basename(newPath)} `;
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error duplicating: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async selectProfile(): Promise<void> {
    const profiles = this.sessionManager.getProfiles();

    if (profiles.length === 0) {
      this.statusMessage =
        " No profiles available. Add them to .profiles.json ";
      this.draw();
      return;
    }

    // Simple profile selection (cycle through)
    const current = this.sessionManager.getActiveProfile();
    const currentIdx = current
      ? profiles.findIndex((p) => p.name === current.name)
      : -1;
    const nextIdx = (currentIdx + 1) % profiles.length;

    this.sessionManager.setActiveProfile(profiles[nextIdx].name);

    // Clear session variables (temporary state) when switching profiles
    this.sessionManager.clearSessionVariables();
    await this.sessionManager.save();

    // Update requestsDir based on new profile's workdir
    this.requestsDir = this.sessionManager.getWorkdir();

    // Reload files from new workdir
    await this.loadFiles();
    this.selectedIndex = 0; // Reset selection

    this.statusMessage = ` Switched to profile: ${
      profiles[nextIdx].name
    } (session cleared) `;
    this.draw();
  }

  async saveResponse(): Promise<void> {
    if (!this.response) {
      this.statusMessage = " No response to save ";
      this.draw();
      return;
    }

    try {
      const timestamp =
        new Date().toISOString().replace(/:/g, "-").split(".")[0];
      const inspectionMode = (this.response as any).inspectionMode;
      const filename = inspectionMode
        ? `inspection-${timestamp}.txt`
        : `response-${timestamp}.txt`;

      let content = "";

      if (inspectionMode) {
        const inspectionData = (this.response as any).inspectionData;
        content += "REQUEST INSPECTION\n";
        content += "==================\n\n";
        content += `${inspectionData.method} ${inspectionData.url}\n\n`;

        if (Object.keys(this.response.headers).length > 0) {
          content += "Headers (merged with profile):\n";
          for (const [key, value] of Object.entries(this.response.headers)) {
            content += `  ${key}: ${value}\n`;
          }
          content += "\n";
        }

        if (this.response.body) {
          content += "Body:\n";
          content += this.response.body;
        }
      } else {
        content +=
          `Status: ${this.response.status} ${this.response.statusText}\n`;
        content += `Duration: ${Math.round(this.response.duration)}ms\n\n`;

        if (this.response.error) {
          content += "Error:\n";
          content += this.response.error + "\n";
        } else {
          if (Object.keys(this.response.headers).length > 0) {
            content += "Headers:\n";
            for (const [key, value] of Object.entries(this.response.headers)) {
              content += `  ${key}: ${value}\n`;
            }
            content += "\n";
          }

          content += "Body:\n";
          content += this.response.body;
        }
      }

      await Deno.writeTextFile(filename, content);
      this.statusMessage = ` Saved to ${filename} `;
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error saving: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async copyResponse(): Promise<void> {
    if (!this.response) {
      this.statusMessage = " No response to copy ";
      this.draw();
      return;
    }

    try {
      // Detect OS and use appropriate clipboard command
      const os = Deno.build.os;
      let cmd: string[];

      if (os === "darwin") {
        cmd = ["pbcopy"];
      } else if (os === "linux") {
        cmd = ["xclip", "-selection", "clipboard"];
      } else if (os === "windows") {
        cmd = ["clip"];
      } else {
        this.statusMessage = " Clipboard not supported on this OS ";
        this.draw();
        return;
      }

      // Copy error message if there's an error, otherwise copy body
      const textToCopy = this.response.error || this.response.body;

      const process = new Deno.Command(cmd[0], {
        args: cmd.slice(1),
        stdin: "piped",
      });

      const child = process.spawn();
      const writer = child.stdin.getWriter();
      await writer.write(new TextEncoder().encode(textToCopy));
      await writer.close();
      await child.status;

      this.statusMessage = this.response.error
        ? " Error message copied to clipboard "
        : " Response body copied to clipboard ";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error copying: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async inspectRequest(): Promise<void> {
    if (this.selectedIndex >= this.files.length) return;

    const file = this.files[this.selectedIndex];
    this.statusMessage = ` Inspecting ${file.name}... `;
    this.draw();

    try {
      const content = await Deno.readTextFile(file.path);
      const parsed = parseHttpFile(content);

      if (parsed.requests.length === 0) {
        this.statusMessage = " No requests found in file ";
        this.draw();
        return;
      }

      // Get first request
      const request = parsed.requests[0];
      const variables = this.sessionManager.getVariables();
      const profileHeaders = this.sessionManager.getActiveHeaders();

      // Apply variable substitution
      const substituted = applyVariables(request, variables);

      // Merge headers (profile + request)
      const mergedHeaders = { ...profileHeaders, ...substituted.headers };

      // Create inspection result that looks like RequestResult
      const inspection = {
        status: 0,
        statusText: "INSPECTION",
        headers: mergedHeaders,
        body: substituted.body || "",
        duration: 0,
        inspectionMode: true,
        inspectionData: {
          name: request.name || "Unnamed Request",
          method: substituted.method,
          url: substituted.url,
        },
      };

      // Store as response to display it
      this.response = inspection as any;
      this.statusMessage =
        " [Inspection Mode] Press Enter to execute, ESC to clear ";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error inspecting: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async refreshFiles(): Promise<void> {
    this.statusMessage = " Refreshing file list... ";
    this.draw();

    try {
      await this.loadFiles();
      this.selectedIndex = Math.min(
        this.selectedIndex,
        Math.max(0, this.files.length - 1),
      );
      this.statusMessage = ` Refreshed - ${this.files.length} file(s) found `;
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error refreshing: ${
        error instanceof Error ? error.message : String(error)
      } `;
      this.draw();
    }
  }

  async start(): Promise<void> {
    await this.init();

    // Enable alternate screen buffer (prevents scrolling)
    console.log("\x1b[?1049h");
    // Hide cursor
    console.log("\x1b[?25l");
    // Enable raw mode
    Deno.stdin.setRaw(true);

    this.running = true;
    this.draw();

    await this.handleInput();

    // Restore terminal
    Deno.stdin.setRaw(false);
    // Show cursor
    console.log("\x1b[?25h");
    // Disable alternate screen buffer
    console.log("\x1b[?1049l");
    console.log("Goodbye!");
  }
}

/**
 * CLI runner for executing HTTP requests without TUI
 * Usage: restcli <path-to-http-file> [--profile <profile-name>]
 */
async function runCLI() {
  const args = Deno.args;

  // Parse command line arguments
  let filePath = "";
  let profileOverride: string | null = null;
  let fullOutput = false;
  let yamlOutput = false;

  // Conditional logger - only logs when --full is set
  // Default: noop (suppressed for clean piping to jq)
  // With --full: stdout
  const clog = (...args: any[]) => {
    if (fullOutput) console.log(...args);
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--help" || arg === "-h") {
      console.log(`
restcli - Terminal HTTP Request Tool

USAGE:
  restcli                              # Open interactive TUI
  restcli <file.http> [OPTIONS]        # Execute request in CLI mode

OPTIONS:
  -h, --help                Show this help message
  -f, --full                Show full output (status, headers, body)
                            Default: body only (perfect for piping)
  -y, --yaml                Convert JSON response to YAML format
  -p, --profile <name>      Use a specific profile for the request

EXAMPLES:
  # TUI mode
  restcli

  # CLI mode - body only (pipe to jq)
  restcli requests/users.http | jq '.users[]'

  # Full output
  restcli --full requests/users.http

  # YAML output
  restcli --yaml requests/users.http

  # With profile
  restcli --profile Admin requests/users.http

  # Combine flags
  restcli --full --yaml --profile Admin requests/users.http

For more information, see: https://github.com/your-repo/http-tui
`);
      Deno.exit(0);
    } else if (arg === "--profile" || arg === "-p") {
      if (i + 1 < args.length) {
        profileOverride = args[i + 1];
        i++; // Skip next arg
      } else {
        console.error("Error: --profile requires a profile name");
        Deno.exit(1);
      }
    } else if (arg === "--full" || arg === "-f") {
      fullOutput = true;
    } else if (arg === "--yaml" || arg === "-y") {
      yamlOutput = true;
    } else if (!filePath) {
      filePath = arg;
    }
  }

  if (!filePath) {
    console.error("Error: No file path specified");
    Deno.exit(1);
  }

  try {
    // Check if config directory exists, use it if available
    const configManager = new ConfigManager();
    const isInitialized = await configManager.isInitialized();

    let baseDir = ".";
    if (isInitialized) {
      baseDir = configManager.getConfigDir();
    }

    // Load session
    const sessionManager = new SessionManager(baseDir);
    await sessionManager.load();

    // Override profile if specified
    if (profileOverride) {
      const profiles = sessionManager.getProfiles();
      const profile = profiles.find((p) => p.name === profileOverride);
      if (!profile) {
        console.error(`Error: Profile "${profileOverride}" not found`);
        console.error("Available profiles:");
        profiles.forEach((p) => console.error(`  - ${p.name}`));
        Deno.exit(1);
      }
      sessionManager.setActiveProfile(profileOverride);
      clog(`Using profile: ${profileOverride}\n`);
    }

    // Read and parse file
    const content = await Deno.readTextFile(filePath);
    const parsed = parseHttpFile(content);

    if (parsed.requests.length === 0) {
      console.error("No requests found in file");
      Deno.exit(1);
    }

    clog(`Found ${parsed.requests.length} request(s) in ${filePath}\n`);

    // Execute first request
    const request = parsed.requests[0];
    clog(`Executing: ${request.name || "Unnamed Request"}`);
    clog(`${request.method} ${request.url}\n`);

    const executor = new RequestExecutor();
    const variables = sessionManager.getVariables();
    const profileHeaders = sessionManager.getActiveHeaders();

    const result = await executor.execute(request, variables, profileHeaders);

    // Save to history if enabled
    if (sessionManager.isHistoryEnabled()) {
      const historyManager = new HistoryManager(baseDir);
      const substituted = applyVariables(request, variables);
      const mergedHeaders = { ...profileHeaders, ...substituted.headers };

      const historyPath = await historyManager.save({
        timestamp: new Date().toISOString(),
        requestFile: filePath,
        requestName: request.name,
        method: substituted.method,
        url: substituted.url,
        headers: mergedHeaders,
        body: substituted.body,
        responseStatus: result.status,
        responseStatusText: result.statusText,
        responseHeaders: result.headers,
        responseBody: result.body,
        duration: result.duration,
        requestSize: result.requestSize,
        responseSize: result.responseSize,
        error: result.error,
      });
      clog(`📝 History saved to: ${historyPath}\n`);
    }

    // Display result
    if (result.error) {
      console.error(`❌ Error: ${result.error}`);
      Deno.exit(1);
    }

    const statusColor = result.status >= 200 && result.status < 300
      ? "\x1b[32m"
      : result.status >= 400
      ? "\x1b[31m"
      : "\x1b[33m";

    // Helper function to format bytes
    function formatBytes(bytes: number): string {
      if (bytes === 0) return "0 B";
      const k = 1024;
      const sizes = ["B", "KB", "MB"];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return Math.round((bytes / Math.pow(k, i)) * 10) / 10 + " " + sizes[i];
    }

    // Status and headers (only with --full)
    clog(
      `${statusColor}${result.status} ${result.statusText}\x1b[0m | ${
        Math.round(result.duration)
      }ms | Req: ${formatBytes(result.requestSize)} | Res: ${
        formatBytes(result.responseSize)
      }\n`,
    );

    clog("Headers:");
    for (const [key, value] of Object.entries(result.headers)) {
      clog(`  ${key}: ${value}`);
    }

    clog("\nBody:");

    // Body (always shown on stdout)
    try {
      const json = JSON.parse(result.body);
      if (yamlOutput) {
        console.log(yamlStringify(json));
      } else {
        console.log(JSON.stringify(json, null, 2));
      }
    } catch {
      console.log(result.body);
    }

    // Try to extract token
    try {
      const json = JSON.parse(result.body);
      if (json.token) {
        sessionManager.setVariable("token", json.token);
        await sessionManager.save();
        console.log("\n✓ Saved token to session");
      }
      if (json.accessToken) {
        sessionManager.setVariable("token", json.accessToken);
        await sessionManager.save();
        console.log("\n✓ Saved accessToken to session");
      }
    } catch {
      // Not JSON or no token
    }
  } catch (error) {
    console.error(
      `Error: ${error instanceof Error ? error.message : String(error)}`,
    );
    Deno.exit(1);
  }
}

if (import.meta.main) {
  // Check if we should run in CLI mode or TUI mode
  const args = Deno.args;

  // If args exist and first arg is not a flag, it's a file path → CLI mode
  if (args.length > 0 && !args[0].startsWith("-")) {
    await runCLI();
  } else if (args.length > 0 && (args[0] === "--help" || args[0] === "-h" || args[0] === "--profile" || args[0] === "-p" || args[0] === "--full" || args[0] === "-f" || args[0] === "--yaml" || args[0] === "-y")) {
    // Flags with file (or help flag) → CLI mode
    await runCLI();
  } else {
    // No args → TUI mode
    const tui = new TUI();
    await tui.start();
  }
}
