import { exists } from "@std/fs";
import * as path from "@std/path";

export class ConfigManager {
  private configDir: string;

  constructor() {
    // Determine config directory
    const home = Deno.env.get("HOME") || Deno.env.get("USERPROFILE");
    if (!home) {
      throw new Error("Cannot determine home directory");
    }

    this.configDir = path.join(home, ".restcli");
  }

  /**
   * Get the config directory path
   */
  getConfigDir(): string {
    return this.configDir;
  }

  /**
   * Get path to a config file
   */
  getConfigPath(filename: string): string {
    return path.join(this.configDir, filename);
  }

  /**
   * Initialize config directory if it doesn't exist
   */
  async init(): Promise<void> {
    if (!await exists(this.configDir)) {
      await Deno.mkdir(this.configDir, { recursive: true });
      console.log(`✨ Initialized config directory: ${this.configDir}`);
    }

    // Create default subdirectories
    const subdirs = ["requests", "history"];
    for (const subdir of subdirs) {
      const dirPath = path.join(this.configDir, subdir);
      if (!await exists(dirPath)) {
        await Deno.mkdir(dirPath, { recursive: true });
      }
    }
  }

  /**
   * Check if config directory exists and is initialized
   */
  async isInitialized(): Promise<boolean> {
    return await exists(this.configDir);
  }

  /**
   * Migrate existing config from current directory to config directory
   */
  async migrate(fromDir: string = "."): Promise<void> {
    const filesToMigrate = [
      ".session.json",
      ".profiles.json",
    ];

    const dirsToMigrate = [
      "requests",
      "history",
    ];

    let migratedCount = 0;

    // Migrate files
    for (const file of filesToMigrate) {
      const sourcePath = path.join(fromDir, file);
      if (await exists(sourcePath)) {
        const destPath = this.getConfigPath(file);
        if (!await exists(destPath)) {
          await Deno.copyFile(sourcePath, destPath);
          console.log(`  Migrated: ${file}`);
          migratedCount++;
        }
      }
    }

    // Migrate directories
    for (const dir of dirsToMigrate) {
      const sourcePath = path.join(fromDir, dir);
      if (await exists(sourcePath)) {
        const destPath = this.getConfigPath(dir);

        // Recursively copy directory
        for await (const entry of Deno.readDir(sourcePath)) {
          if (entry.isFile) {
            const sourceFile = path.join(sourcePath, entry.name);
            const destFile = path.join(destPath, entry.name);
            if (!await exists(destFile)) {
              await Deno.copyFile(sourceFile, destFile);
              migratedCount++;
            }
          } else if (entry.isDirectory) {
            // Recursively handle subdirectories
            const sourceSubdir = path.join(sourcePath, entry.name);
            const destSubdir = path.join(destPath, entry.name);
            await Deno.mkdir(destSubdir, { recursive: true });

            // Copy files from subdirectory
            for await (const subEntry of Deno.readDir(sourceSubdir)) {
              if (subEntry.isFile) {
                const sourceFile = path.join(sourceSubdir, subEntry.name);
                const destFile = path.join(destSubdir, subEntry.name);
                if (!await exists(destFile)) {
                  await Deno.copyFile(sourceFile, destFile);
                  migratedCount++;
                }
              }
            }
          }
        }
        console.log(`  Migrated directory: ${dir}`);
      }
    }

    if (migratedCount > 0) {
      console.log(`\n✅ Migrated ${migratedCount} item(s) to ${this.configDir}`);
      console.log(`\n💡 You can now delete the old files from ${path.resolve(fromDir)}`);
    } else {
      console.log("No files to migrate.");
    }
  }

  /**
   * Create example config files
   */
  async createExamples(): Promise<void> {
    // Create example .session.json
    const sessionPath = this.getConfigPath(".session.json");
    if (!await exists(sessionPath)) {
      const exampleSession = {
        variables: {
          baseUrl: "http://localhost:3000",
          userId: "1",
        },
        activeProfile: "Dev - Default",
        historyEnabled: true,
      };
      await Deno.writeTextFile(sessionPath, JSON.stringify(exampleSession, null, 2));
      console.log("  Created: .session.json");
    }

    // Create example .profiles.json
    const profilesPath = this.getConfigPath(".profiles.json");
    if (!await exists(profilesPath)) {
      const exampleProfiles = [
        {
          name: "Dev - Default",
          workdir: "requests",
          variables: {
            baseUrl: "http://localhost:3000",
          },
          headers: {
            "Content-Type": "application/json",
          },
        },
      ];
      await Deno.writeTextFile(profilesPath, JSON.stringify(exampleProfiles, null, 2));
      console.log("  Created: .profiles.json");
    }

    // Create example request
    const requestsDir = this.getConfigPath("requests");
    const exampleRequestPath = path.join(requestsDir, "example.http");
    if (!await exists(exampleRequestPath)) {
      const exampleRequest = `### Example GET Request
GET {{baseUrl}}/api/endpoint

### Example POST Request
POST {{baseUrl}}/api/endpoint
Content-Type: application/json

{
  "key": "value"
}
`;
      await Deno.writeTextFile(exampleRequestPath, exampleRequest);
      console.log("  Created: requests/example.http");
    }
  }
}
