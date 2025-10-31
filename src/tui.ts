import { walk } from "@std/fs";
import * as path from "@std/path";
import { parseHttpFile, applyVariables } from "./parser.ts";
import { RequestExecutor, type RequestResult } from "./executor.ts";
import { SessionManager } from "./session.ts";

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
  private executor: RequestExecutor;
  private requestsDir = "./requests";
  private running = false;
  private statusMessage = "";
  private searchMode = false;
  private searchQuery = "";
  private searchResults: number[] = [];
  private searchResultIndex = 0;
  private gotoMode = false;
  private gotoQuery = "";
  private pageSize = 10; // Will be calculated dynamically

  constructor() {
    this.sessionManager = new SessionManager();
    this.executor = new RequestExecutor();
  }

  async init(): Promise<void> {
    await this.sessionManager.load();
    await this.loadFiles();
  }

  async loadFiles(): Promise<void> {
    this.files = [];
    try {
      for await (const entry of walk(this.requestsDir, { exts: [".http"] })) {
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
    console.log("\x1b[2J\x1b[H");
  }

  private moveCursor(row: number, col: number): void {
    Deno.stdout.writeSync(new TextEncoder().encode(`\x1b[${row};${col}H`));
  }

  private write(text: string): void {
    Deno.stdout.writeSync(new TextEncoder().encode(text));
  }

  private draw(): void {
    this.clear();

    const width = Deno.consoleSize().columns;
    const height = Deno.consoleSize().rows;

    const sidebarWidth = Math.min(40, Math.floor(width * 0.3));
    const separatorCol = sidebarWidth + 1;
    const mainStartCol = separatorCol + 2;
    const mainWidth = width - mainStartCol;

    // Calculate page size for page up/down
    this.pageSize = Math.max(1, height - 7); // Reserve space for header, title, scroll indicator, status

    // Header
    this.drawHeader(width);

    // Sidebar
    this.drawSidebar(sidebarWidth, height - 3);

    // Vertical separator
    this.drawSeparator(separatorCol, height - 2);

    // Main content
    this.drawMain(mainStartCol, mainWidth, height - 3);

    // Status bar
    this.drawStatusBar(width, height);

    // Position cursor at bottom
    this.moveCursor(height, 1);
  }

  private drawSeparator(col: number, height: number): void {
    for (let row = 2; row <= height; row++) {
      this.moveCursor(row, col);
      this.write("\x1b[2m│\x1b[0m");
    }
  }

  private drawHeader(width: number): void {
    const activeProfile = this.sessionManager.getActiveProfile();
    const profileName = activeProfile ? activeProfile.name : "No Profile";
    const header = ` HTTP TUI | Profile: ${profileName} `;
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
    const startIdx = Math.max(0, Math.min(this.selectedIndex - Math.floor(maxVisibleLines / 2), totalFiles - maxVisibleLines));
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
        const lineNum = (globalIdx + 1).toString(16).toUpperCase().padStart(numWidth, " ");
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
          this.write(`${lineNumDisplay}\x1b[7m${prefixVisible}${displayName}\x1b[0m${padding}`);
        } else {
          this.write(`${lineNumDisplay}${prefixVisible}${displayName}${padding}`);
        }
      } else {
        const clearLine = " ".repeat(width);
        this.write(clearLine);
      }
    }

    // Scroll indicator
    this.moveCursor(3 + maxVisibleLines, 1);
    if (totalFiles > maxVisibleLines) {
      const scrollPercent = Math.round((this.selectedIndex / (totalFiles - 1)) * 100);
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
    const statusColor = this.response.status >= 200 && this.response.status < 300
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
      const statusText = `${this.response.status} ${this.response.statusText} | ${Math.round(this.response.duration)}ms`.slice(0, width - 2);
      this.write(`${statusColor}${statusText}\x1b[0m\x1b[K`);
      line++;
    }

    this.moveCursor(line++, startCol);
    this.write("\x1b[K");

    // Headers
    if (Object.keys(this.response.headers).length > 0) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[2mHeaders:\x1b[0m\x1b[K");
      for (const [key, value] of Object.entries(this.response.headers).slice(0, 5)) {
        this.moveCursor(line++, startCol);
        const display = `${key}: ${value}`.slice(0, width - 2);
        this.write(`  ${display}\x1b[K`);
      }
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }

    // Body
    this.moveCursor(line++, startCol);
    this.write("\x1b[2mBody:\x1b[0m\x1b[K");

    const bodyLines = this.response.body.split("\n");
    const maxLines = height - line;
    for (let i = 0; i < Math.min(bodyLines.length, maxLines); i++) {
      this.moveCursor(line++, startCol);
      const displayLine = bodyLines[i].slice(0, width - 2);
      this.write(`${displayLine}\x1b[K`);
    }

    // Clear remaining lines
    while (line < height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawInspection(startCol: number, width: number, height: number): void {
    const inspectionData = (this.response as any).inspectionData;
    let line = 3;

    // Request line
    this.moveCursor(line++, startCol);
    const requestLine = `${inspectionData.method} ${inspectionData.url}`.slice(0, width - 2);
    this.write(`\x1b[1;36m${requestLine}\x1b[0m\x1b[K`);

    line++;
    this.moveCursor(line++, startCol);
    this.write("\x1b[K");

    // Headers
    if (this.response && Object.keys(this.response.headers).length > 0) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[2mHeaders (merged with profile):\x1b[0m\x1b[K");

      for (const [key, value] of Object.entries(this.response.headers)) {
        this.moveCursor(line++, startCol);
        const display = `  ${key}: ${value}`.slice(0, width - 2);
        this.write(`${display}\x1b[K`);

        if (line >= height - 5) break; // Leave room for body
      }

      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }

    // Body
    if (this.response && this.response.body) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[2mBody:\x1b[0m\x1b[K");

      const bodyLines = this.response.body.split("\n");
      const maxLines = height - line - 2;
      for (let i = 0; i < Math.min(bodyLines.length, maxLines); i++) {
        this.moveCursor(line++, startCol);
        const displayLine = bodyLines[i].slice(0, width - 2);
        this.write(`${displayLine}\x1b[K`);
      }
    }

    // Clear remaining lines
    while (line < height) {
      this.moveCursor(line++, startCol);
      this.write("\x1b[K");
    }
  }

  private drawStatusBar(width: number, row: number): void {
    this.moveCursor(row, 1);

    let statusText: string;

    if (this.gotoMode) {
      statusText = ` Go to (hex): :${this.gotoQuery}_ | [Enter] Jump | [ESC] Cancel `;
    } else if (this.searchMode) {
      const matchCount = this.searchResults.length;
      const currentMatch = matchCount > 0 ? this.searchResultIndex + 1 : 0;
      statusText = ` Search: ${this.searchQuery}_ | ${currentMatch}/${matchCount} matches | [Ctrl+R] Next | [ESC] Cancel | [Enter] Select `;
    } else {
      const help = " [↑↓] Nav | [Enter] Execute | [i] Inspect | [:] Goto | [Ctrl+R] Search | [d] Dup | [s] Save | [c] Copy | [p] Profile | [q] Quit ";
      statusText = this.statusMessage || help;
    }

    const padding = " ".repeat(Math.max(0, width - statusText.length));
    this.write(`\x1b[7m${statusText}${padding}\x1b[0m`);
  }

  async handleInput(): Promise<void> {
    const buf = new Uint8Array(8);

    while (this.running) {
      const n = await Deno.stdin.read(buf);
      if (!n) continue;

      const input = buf.subarray(0, n);

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
          this.draw();
        } else if (input[2] === 66) {
          // Down
          if (this.selectedIndex === this.files.length - 1) {
            this.selectedIndex = 0; // Wrap to top
          } else {
            this.selectedIndex++;
          }
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
          this.selectedIndex = Math.min(this.files.length - 1, this.selectedIndex + this.pageSize);
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
        }
      }
    }
  }

  async handleSearchInput(input: Uint8Array): Promise<void> {
    // Ctrl+R - cycle through results
    if (input.length === 1 && input[0] === 18) {
      if (this.searchResults.length > 0) {
        this.searchResultIndex = (this.searchResultIndex + 1) % this.searchResults.length;
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
      if ((input[0] >= 48 && input[0] <= 57) || // 0-9
          (input[0] >= 65 && input[0] <= 70) || // A-F
          (input[0] >= 97 && input[0] <= 102)) { // a-f
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

      this.response = await this.executor.execute(request, variables, profileHeaders);

      // Save response body variables if it's JSON
      try {
        const json = JSON.parse(this.response.body);
        if (json.token) this.sessionManager.setVariable("token", json.token);
        if (json.accessToken) this.sessionManager.setVariable("token", json.accessToken);
        await this.sessionManager.save();
      } catch {
        // Not JSON or no token
      }

      this.statusMessage = "";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error: ${error instanceof Error ? error.message : String(error)} `;
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
      this.selectedIndex = this.files.findIndex(f => f.path === newPath);
      this.statusMessage = ` Duplicated to ${path.basename(newPath)} `;
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error duplicating: ${error instanceof Error ? error.message : String(error)} `;
      this.draw();
    }
  }

  async selectProfile(): Promise<void> {
    const profiles = this.sessionManager.getProfiles();

    if (profiles.length === 0) {
      this.statusMessage = " No profiles available. Add them to .profiles.json ";
      this.draw();
      return;
    }

    // Simple profile selection (cycle through)
    const current = this.sessionManager.getActiveProfile();
    const currentIdx = current ? profiles.findIndex(p => p.name === current.name) : -1;
    const nextIdx = (currentIdx + 1) % profiles.length;

    this.sessionManager.setActiveProfile(profiles[nextIdx].name);
    await this.sessionManager.save();

    this.statusMessage = ` Switched to profile: ${profiles[nextIdx].name} `;
    this.draw();
  }

  async saveResponse(): Promise<void> {
    if (!this.response) {
      this.statusMessage = " No response to save ";
      this.draw();
      return;
    }

    try {
      const timestamp = new Date().toISOString().replace(/:/g, "-").split(".")[0];
      const inspectionMode = (this.response as any).inspectionMode;
      const filename = inspectionMode ? `inspection-${timestamp}.txt` : `response-${timestamp}.txt`;

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
        content += `Status: ${this.response.status} ${this.response.statusText}\n`;
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
      this.statusMessage = ` Error saving: ${error instanceof Error ? error.message : String(error)} `;
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
      this.statusMessage = ` Error copying: ${error instanceof Error ? error.message : String(error)} `;
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
        }
      };

      // Store as response to display it
      this.response = inspection as any;
      this.statusMessage = " [Inspection Mode] Press Enter to execute, ESC to clear ";
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error inspecting: ${error instanceof Error ? error.message : String(error)} `;
      this.draw();
    }
  }

  async refreshFiles(): Promise<void> {
    this.statusMessage = " Refreshing file list... ";
    this.draw();

    try {
      await this.loadFiles();
      this.selectedIndex = Math.min(this.selectedIndex, Math.max(0, this.files.length - 1));
      this.statusMessage = ` Refreshed - ${this.files.length} file(s) found `;
      this.draw();
    } catch (error) {
      this.statusMessage = ` Error refreshing: ${error instanceof Error ? error.message : String(error)} `;
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

if (import.meta.main) {
  const tui = new TUI();
  await tui.start();
}
