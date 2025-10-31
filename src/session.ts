import { exists } from "@std/fs";
import * as path from "@std/path";

export interface Session {
  variables: Record<string, string>;
  activeProfile?: string;
}

export interface HeaderProfile {
  name: string;
  headers: Record<string, string>;
  variables?: Record<string, string>;
}

export class SessionManager {
  private session: Session;
  private sessionFile: string;
  private profilesFile: string;
  private profiles: HeaderProfile[] = [];

  constructor(baseDir: string = ".") {
    this.sessionFile = path.join(baseDir, ".session.json");
    this.profilesFile = path.join(baseDir, ".profiles.json");
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

  async saveProfiles(): Promise<void> {
    await Deno.writeTextFile(
      this.profilesFile,
      JSON.stringify(this.profiles, null, 2)
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
