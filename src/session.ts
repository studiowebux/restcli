import { exists } from "@std/fs";
import * as path from "@std/path";
import type { OAuthConfig } from "./oauth-config.ts";

export interface Session {
  variables: Record<string, string>;
  activeProfile?: string;
  historyEnabled?: boolean; // Default: true if not specified
}

export interface HeaderProfile {
  name: string;
  headers: Record<string, string>;
  variables?: Record<string, string>;
  workdir?: string; // Custom working directory for this profile
  oauth?: OAuthConfig; // OAuth configuration
}

export class SessionManager {
  private session: Session;
  private sessionFile: string;
  private profilesFile: string;
  private profiles: HeaderProfile[] = [];
  private baseDir: string;

  constructor(baseDir?: string) {
    // If baseDir not provided, use current directory for backward compatibility
    this.baseDir = baseDir ?? ".";
    this.sessionFile = path.join(this.baseDir, ".session.json");
    this.profilesFile = path.join(this.baseDir, ".profiles.json");
    this.session = { variables: {} };
  }

  async load(): Promise<void> {
    // Load session
    if (await exists(this.sessionFile)) {
      const content = await Deno.readTextFile(this.sessionFile);
      this.session = JSON.parse(content);
    }

    // Load profiles
    if (await exists(this.profilesFile)) {
      const content = await Deno.readTextFile(this.profilesFile);
      this.profiles = JSON.parse(content);
    }
  }

  async save(): Promise<void> {
    await Deno.writeTextFile(
      this.sessionFile,
      JSON.stringify(this.session, null, 2)
    );
  }

  getVariables(): Record<string, string> {
    // Start with global variables
    const vars = { ...this.session.variables };

    // Merge profile-specific variables (they override global ones)
    const profile = this.getActiveProfile();
    if (profile && profile.variables) {
      Object.assign(vars, profile.variables);
    }

    return vars;
  }

  setVariable(key: string, value: string): void {
    this.session.variables[key] = value;
  }

  /**
   * Clear session variables (temporary state)
   */
  clearSessionVariables(): void {
    this.session.variables = {};
  }

  /**
   * Get profile-specific variables (not merged with session)
   */
  getProfileVariables(): Record<string, string> {
    const profile = this.getActiveProfile();
    return profile?.variables ?? {};
  }

  /**
   * Set a variable in the active profile
   */
  setProfileVariable(key: string, value: string): void {
    const profile = this.getActiveProfile();
    if (profile) {
      if (!profile.variables) {
        profile.variables = {};
      }
      profile.variables[key] = value;
    }
  }

  /**
   * Delete a variable from the active profile
   */
  deleteProfileVariable(key: string): void {
    const profile = this.getActiveProfile();
    if (profile && profile.variables) {
      delete profile.variables[key];
    }
  }

  /**
   * Get profile-specific headers (not merged, not substituted)
   */
  getProfileHeaders(): Record<string, string> {
    const profile = this.getActiveProfile();
    return profile?.headers ?? {};
  }

  /**
   * Set a header in the active profile
   */
  setProfileHeader(key: string, value: string): void {
    const profile = this.getActiveProfile();
    if (profile) {
      if (!profile.headers) {
        profile.headers = {};
      }
      profile.headers[key] = value;
    }
  }

  /**
   * Delete a header from the active profile
   */
  deleteProfileHeader(key: string): void {
    const profile = this.getActiveProfile();
    if (profile && profile.headers) {
      delete profile.headers[key];
    }
  }

  getProfiles(): HeaderProfile[] {
    return this.profiles;
  }

  getActiveProfile(): HeaderProfile | undefined {
    if (!this.session.activeProfile) return undefined;
    return this.profiles.find(p => p.name === this.session.activeProfile);
  }

  setActiveProfile(name: string): void {
    this.session.activeProfile = name;
  }

  addProfile(profile: HeaderProfile): void {
    this.profiles.push(profile);
  }

  /**
   * Save profiles to disk
   */
  async saveProfiles(): Promise<void> {
    await Deno.writeTextFile(
      this.profilesFile,
      JSON.stringify(this.profiles, null, 2)
    );
  }

  /**
   * Get OAuth config for active profile
   */
  getOAuthConfig(): OAuthConfig | undefined {
    const profile = this.getActiveProfile();
    return profile?.oauth;
  }

  /**
   * Check if history saving is enabled
   * Defaults to true if not explicitly set
   */
  isHistoryEnabled(): boolean {
    return this.session.historyEnabled !== false;
  }

  /**
   * Set history enabled/disabled
   */
  setHistoryEnabled(enabled: boolean): void {
    this.session.historyEnabled = enabled;
  }

  /**
   * Get working directory for the active profile
   * Returns the profile's workdir or defaults to "requests"
   * Always returns absolute path relative to baseDir
   */
  getWorkdir(): string {
    const profile = this.getActiveProfile();
    const relativeWorkdir = profile?.workdir ?? "requests";

    // If it's already absolute, return as-is
    if (path.isAbsolute(relativeWorkdir)) {
      return relativeWorkdir;
    }

    // Make it absolute relative to baseDir
    return path.join(this.baseDir, relativeWorkdir);
  }

  /**
   * Get headers from active profile with variable substitution
   */
  getActiveHeaders(): Record<string, string> {
    const profile = this.getActiveProfile();
    if (!profile) return {};

    // Get merged variables (global + profile-specific)
    const vars = this.getVariables();

    const headers: Record<string, string> = {};
    for (const [key, value] of Object.entries(profile.headers)) {
      headers[key] = this.substituteVars(value, vars);
    }
    return headers;
  }

  private substituteVars(text: string, vars?: Record<string, string>): string {
    const variables = vars ?? this.session.variables;
    return text.replace(/\{\{(\w+)\}\}/g, (_, varName) => {
      return variables[varName] ?? `{{${varName}}}`;
    });
  }
}
