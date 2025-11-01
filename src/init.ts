import { ConfigManager } from "./config.ts";

/**
 * Initialize restcli configuration
 * Usage: deno task init or restcli init
 */
async function main() {
  const args = Deno.args;

  console.log("🚀 Initializing restcli...\n");

  const configManager = new ConfigManager();

  // Check if already initialized
  const isInit = await configManager.isInitialized();

  if (isInit && !args.includes("--force")) {
    console.log(`ℹ️  Config directory already exists: ${configManager.getConfigDir()}`);
    console.log("\nUse 'restcli init --force' to reinitialize or '--migrate' to migrate from current directory\n");

    if (args.includes("--migrate")) {
      console.log("📦 Migrating existing configuration...\n");
      await configManager.migrate(".");
    }

    return;
  }

  // Initialize
  await configManager.init();

  if (args.includes("--migrate")) {
    console.log("\n📦 Migrating existing configuration...\n");
    await configManager.migrate(".");
  } else {
    console.log("\n📝 Creating example configuration files...\n");
    await configManager.createExamples();
  }

  console.log(`\n✅ Configuration initialized at: ${configManager.getConfigDir()}`);
  console.log("\n📚 Next steps:");
  console.log("  1. Edit ~/.restcli/.profiles.json to configure your profiles");
  console.log("  2. Edit ~/.restcli/.session.json to set variables");
  console.log("  3. Add your .http request files to ~/.restcli/requests/");
  console.log("  4. Run 'restcli' to start the TUI\n");
}

if (import.meta.main) {
  await main();
}
