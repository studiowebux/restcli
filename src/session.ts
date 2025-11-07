import { exists } from "@std/fs";
import * as path from "@std/path";
import type { OAuthConfig } from "./oauth/oauth-config.ts";

/**
 * Multi-value variable configuration
 * Allows a variable to have multiple predefined options with one active
 */
export interface MultiValueVariable {
  options: string[];      // Available values
  active: number;         // Index of currently active option
  description?: string;   // Optional description
}

/**
 * Variable value can be either a simple string or multi-value config
 */
export type VariableValue = string | MultiValueVariable;

/**
 * Type guard to check if a variable is multi-value
 */
export function isMultiValueVariable(value: VariableValue): value is MultiValueVariable {
  return typeof value === 'object' &&
         value !== null &&
         'options' in value &&
         'active' in value &&
         Array.isArray(value.options);
}

export interface Session {
  variables: Record<string, string>;
  activeProfile?: string;
  historyEnabled?: boolean; // Default: true if not specified
}

export interface HeaderProfile {
  name: string;
  headers: Record<string, string>;
  variables?: Record<string, VariableValue>;
  workdir?: string; // Custom working directory for this profile
  oauth?: OAuthConfig; // OAuth configuration
  editor?: string; // External editor command (e.g., "code", "zed", "vim")
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
      JSON.stringify(this.session, null, 2),
    );
  }

  getVariables(): Record<string, string> {
    // Start with global variables
    const vars = { ...this.session.variables };

    // Merge profile-specific variables (they override global ones)
    const profile = this.getActiveProfile();
    if (profile && profile.variables) {
      for (const [key, value] of Object.entries(profile.variables)) {
        if (isMultiValueVariable(value)) {
          // Resolve multi-value variable to active option
          if (value.options.length > 0 && value.active >= 0 && value.active < value.options.length) {
            vars[key] = value.options[value.active];
          } else {
            // Invalid active index, use empty string
            vars[key] = '';
          }
        } else {
          // Simple string variable
          vars[key] = value;
        }
      }
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
   * Returns raw variable values (including multi-value objects)
   */
  getProfileVariables(): Record<string, VariableValue> {
    const profile = this.getActiveProfile();
    return profile?.variables ?? {};
  }

  /**
   * Set a variable in the active profile
   */
  setProfileVariable(key: string, value: VariableValue): void {
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
   * Check if a variable is multi-value
   */
  isVariableMultiValue(key: string): boolean {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return false;
    }
    return isMultiValueVariable(profile.variables[key]);
  }

  /**
   * Get options for a multi-value variable
   */
  getVariableOptions(key: string): string[] | undefined {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return undefined;
    }
    const value = profile.variables[key];
    return isMultiValueVariable(value) ? value.options : undefined;
  }

  /**
   * Get active option index for a multi-value variable
   */
  getVariableActiveOption(key: string): number | undefined {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return undefined;
    }
    const value = profile.variables[key];
    return isMultiValueVariable(value) ? value.active : undefined;
  }

  /**
   * Set the active option for a multi-value variable
   */
  setVariableActiveOption(key: string, activeIndex: number): boolean {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return false;
    }
    const value = profile.variables[key];
    if (!isMultiValueVariable(value)) {
      return false;
    }
    if (activeIndex < 0 || activeIndex >= value.options.length) {
      return false;
    }
    value.active = activeIndex;
    return true;
  }

  /**
   * Add an option to a multi-value variable
   */
  addVariableOption(key: string, option: string): boolean {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return false;
    }
    const value = profile.variables[key];
    if (!isMultiValueVariable(value)) {
      return false;
    }
    // Check for duplicate
    if (value.options.includes(option)) {
      return false;
    }
    value.options.push(option);
    return true;
  }

  /**
   * Remove an option from a multi-value variable
   */
  removeVariableOption(key: string, optionIndex: number): boolean {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return false;
    }
    const value = profile.variables[key];
    if (!isMultiValueVariable(value)) {
      return false;
    }
    if (optionIndex < 0 || optionIndex >= value.options.length) {
      return false;
    }
    // Don't allow removing the active option
    if (optionIndex === value.active) {
      return false;
    }
    value.options.splice(optionIndex, 1);
    // Adjust active index if needed
    if (value.active > optionIndex) {
      value.active--;
    }
    return true;
  }

  /**
   * Update an option value in a multi-value variable
   */
  updateVariableOption(key: string, optionIndex: number, newValue: string): boolean {
    const profile = this.getActiveProfile();
    if (!profile || !profile.variables || !(key in profile.variables)) {
      return false;
    }
    const value = profile.variables[key];
    if (!isMultiValueVariable(value)) {
      return false;
    }
    if (optionIndex < 0 || optionIndex >= value.options.length) {
      return false;
    }
    // Check for duplicate (excluding current index)
    const otherOptions = value.options.filter((_, i) => i !== optionIndex);
    if (otherOptions.includes(newValue)) {
      return false;
    }
    value.options[optionIndex] = newValue;
    return true;
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
    return this.profiles.find((p) => p.name === this.session.activeProfile);
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
      JSON.stringify(this.profiles, null, 2),
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
   * Get configured editor for the active profile
   * Returns undefined if no editor is configured
   */
  getEditor(): string | undefined {
    const profile = this.getActiveProfile();
    return profile?.editor;
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
