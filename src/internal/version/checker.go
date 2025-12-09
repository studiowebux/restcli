package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/studiowebux/restcli/releases/latest"
	checkTimeout = 5 * time.Second
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(currentVersion string) (available bool, latestVersion string, url string, err error) {
	client := &http.Client{
		Timeout: checkTimeout,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "restcli/"+currentVersion)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	latestVersion = strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	if latestVersion != "" && isNewerVersion(latestVersion, currentVersion) {
		return true, latestVersion, release.HTMLURL, nil
	}

	return false, latestVersion, release.HTMLURL, nil
}

// isNewerVersion compares two semantic versions and returns true if latest > current
// Supports versions like "0.0.28", "1.2.3", "0.0.29-dev", etc.
func isNewerVersion(latest, current string) bool {
	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)

	// Pad shorter version with zeros
	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for len(latestParts) < maxLen {
		latestParts = append(latestParts, 0)
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, 0)
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return false
}

// parseVersion parses a version string into integer parts
// Handles pre-release versions by stripping everything after "-" or "+"
func parseVersion(version string) []int {
	// Strip pre-release and build metadata (everything after - or +)
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			// If we can't parse a number, skip it
			continue
		}
		result = append(result, num)
	}

	return result
}
