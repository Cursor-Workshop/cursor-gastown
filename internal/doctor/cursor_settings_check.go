package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cursorworkshop/cursor-gastown/internal/cursor"
	"github.com/cursorworkshop/cursor-gastown/internal/session"
	"github.com/cursorworkshop/cursor-gastown/internal/templates"
	"github.com/cursorworkshop/cursor-gastown/internal/tmux"
	"github.com/cursorworkshop/cursor-gastown/internal/workspace"
)

// gitFileStatus represents the git status of a file.
type gitFileStatus string

const (
	gitStatusUntracked       gitFileStatus = "untracked"        // File not tracked by git
	gitStatusTrackedClean    gitFileStatus = "tracked-clean"    // Tracked, no local modifications
	gitStatusTrackedModified gitFileStatus = "tracked-modified" // Tracked with local modifications
	gitStatusUnknown         gitFileStatus = "unknown"          // Not in a git repo or error
)

// CursorSettingsCheck verifies that Cursor settings files match the expected templates.
// Detects stale settings files that are missing required hooks or configuration.
type CursorSettingsCheck struct {
	FixableCheck
	staleSettings []staleSettingsInfo
}

type staleSettingsInfo struct {
	path          string        // Full path to settings.json
	agentType     string        // e.g., "witness", "refinery", "deacon", "mayor"
	rigName       string        // Rig name (empty for town-level agents)
	sessionName   string        // tmux session name for cycling
	missing       []string      // What's missing from the settings
	wrongLocation bool          // True if file is in wrong location (should be deleted)
	gitStatus     gitFileStatus // Git status for wrong-location files (for safe deletion)
}

// NewCursorSettingsCheck creates a new Cursor settings validation check.
func NewCursorSettingsCheck() *CursorSettingsCheck {
	return &CursorSettingsCheck{
		FixableCheck: FixableCheck{
			BaseCheck: BaseCheck{
				CheckName:        "cursor-settings",
				CheckDescription: "Verify Cursor settings files match expected templates",
			},
		},
	}
}

// Run checks all Cursor settings files for staleness.
func (c *CursorSettingsCheck) Run(ctx *CheckContext) *CheckResult {
	c.staleSettings = nil

	var details []string
	var hasModifiedFiles bool

	// Find all settings.json files
	settingsFiles := c.findSettingsFiles(ctx.TownRoot)

	for _, sf := range settingsFiles {
		// Files in wrong locations are always stale (should be deleted)
		if sf.wrongLocation {
			// Check git status to determine safe deletion strategy
			sf.gitStatus = c.getGitFileStatus(sf.path)
			c.staleSettings = append(c.staleSettings, sf)

			// Provide detailed message based on git status
			var statusMsg string
			switch sf.gitStatus {
			case gitStatusUntracked:
				statusMsg = "wrong location, untracked (safe to delete)"
			case gitStatusTrackedClean:
				statusMsg = "wrong location, tracked but unmodified (safe to delete)"
			case gitStatusTrackedModified:
				statusMsg = "wrong location, tracked with local modifications (manual review needed)"
				hasModifiedFiles = true
			default:
				statusMsg = "wrong location (inside source repo)"
			}
			details = append(details, fmt.Sprintf("%s: %s", sf.path, statusMsg))
			continue
		}

		// Check content of files in correct locations
		missing := c.checkSettings(sf.path, sf.agentType)
		if len(missing) > 0 {
			sf.missing = missing
			c.staleSettings = append(c.staleSettings, sf)
			details = append(details, fmt.Sprintf("%s: missing %s", sf.path, strings.Join(missing, ", ")))
		}
	}

	if len(c.staleSettings) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: "All Cursor settings files are up to date",
		}
	}

	fixHint := "Run 'gt doctor --fix' to update settings and restart affected agents"
	if hasModifiedFiles {
		fixHint = "Run 'gt doctor --fix' to fix safe issues. Files with local modifications require manual review."
	}

	return &CheckResult{
		Name:    c.Name(),
		Status:  StatusError,
		Message: fmt.Sprintf("Found %d stale Cursor config file(s) in wrong location", len(c.staleSettings)),
		Details: details,
		FixHint: fixHint,
	}
}

// findSettingsFiles locates all .cursor/ settings files and identifies their agent type.
func (c *CursorSettingsCheck) findSettingsFiles(townRoot string) []staleSettingsInfo {
	var files []staleSettingsInfo

	// Check for STALE settings at town root (~/gt/.cursor/)
	// This is WRONG - settings here pollute ALL child workspaces via directory traversal.
	// Mayor settings should be at ~/gt/mayor/.cursor/ instead.
	staleTownRootSettings := filepath.Join(townRoot, ".cursor", "hooks.json")
	if fileExists(staleTownRootSettings) {
		files = append(files, staleSettingsInfo{
			path:          staleTownRootSettings,
			agentType:     "mayor",
			sessionName:   "hq-mayor",
			wrongLocation: true,
			gitStatus:     c.getGitFileStatus(staleTownRootSettings),
			missing:       []string{"should be at mayor/.cursor/, not town root"},
		})
	}

	// Check for STALE CLAUDE.md at town root (~/gt/CLAUDE.md)
	// This is WRONG - CLAUDE.md here is inherited by ALL agents via directory traversal,
	// causing crew/polecat/etc to receive Mayor-specific instructions.
	// Mayor's CLAUDE.md should be at ~/gt/mayor/CLAUDE.md instead.
	staleTownRootCLAUDEmd := filepath.Join(townRoot, "CLAUDE.md")
	if fileExists(staleTownRootCLAUDEmd) {
		files = append(files, staleSettingsInfo{
			path:          staleTownRootCLAUDEmd,
			agentType:     "mayor",
			sessionName:   "hq-mayor",
			wrongLocation: true,
			gitStatus:     c.getGitFileStatus(staleTownRootCLAUDEmd),
			missing:       []string{"should be at mayor/CLAUDE.md, not town root"},
		})
	}

	// Town-level: mayor (~/gt/mayor/.cursor/hooks.json) - CORRECT location
	mayorSettings := filepath.Join(townRoot, "mayor", ".cursor", "hooks.json")
	if fileExists(mayorSettings) {
		files = append(files, staleSettingsInfo{
			path:        mayorSettings,
			agentType:   "mayor",
			sessionName: "hq-mayor",
		})
	}

	// Town-level: deacon (~/gt/deacon/.cursor/hooks.json)
	deaconSettings := filepath.Join(townRoot, "deacon", ".cursor", "hooks.json")
	if fileExists(deaconSettings) {
		files = append(files, staleSettingsInfo{
			path:        deaconSettings,
			agentType:   "deacon",
			sessionName: "hq-deacon",
		})
	}

	// Find rig directories
	entries, err := os.ReadDir(townRoot)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		rigName := entry.Name()
		rigPath := filepath.Join(townRoot, rigName)

		// Skip known non-rig directories
		if rigName == "mayor" || rigName == "deacon" || rigName == "daemon" ||
			rigName == ".git" || rigName == "docs" || rigName[0] == '.' {
			continue
		}

		// Check for witness settings - witness/.cursor/ is correct (outside git repo)
		// Settings in witness/rig/.cursor/ are wrong (inside source repo)
		witnessSettings := filepath.Join(rigPath, "witness", ".cursor", "hooks.json")
		if fileExists(witnessSettings) {
			files = append(files, staleSettingsInfo{
				path:        witnessSettings,
				agentType:   "witness",
				rigName:     rigName,
				sessionName: fmt.Sprintf("gt-%s-witness", rigName),
			})
		}
		witnessWrongSettings := filepath.Join(rigPath, "witness", "rig", ".cursor", "hooks.json")
		if fileExists(witnessWrongSettings) {
			files = append(files, staleSettingsInfo{
				path:          witnessWrongSettings,
				agentType:     "witness",
				rigName:       rigName,
				sessionName:   fmt.Sprintf("gt-%s-witness", rigName),
				wrongLocation: true,
			})
		}

		// Check for refinery settings - refinery/.cursor/ is correct (outside git repo)
		// Settings in refinery/rig/.cursor/ are wrong (inside source repo)
		refinerySettings := filepath.Join(rigPath, "refinery", ".cursor", "hooks.json")
		if fileExists(refinerySettings) {
			files = append(files, staleSettingsInfo{
				path:        refinerySettings,
				agentType:   "refinery",
				rigName:     rigName,
				sessionName: fmt.Sprintf("gt-%s-refinery", rigName),
			})
		}
		refineryWrongSettings := filepath.Join(rigPath, "refinery", "rig", ".cursor", "hooks.json")
		if fileExists(refineryWrongSettings) {
			files = append(files, staleSettingsInfo{
				path:          refineryWrongSettings,
				agentType:     "refinery",
				rigName:       rigName,
				sessionName:   fmt.Sprintf("gt-%s-refinery", rigName),
				wrongLocation: true,
			})
		}

		// Check for crew settings - crew/.cursor/ is correct (shared by all crew, outside git repos)
		// Settings in crew/<name>/.cursor/ are wrong (inside git repos)
		crewDir := filepath.Join(rigPath, "crew")
		crewSettings := filepath.Join(crewDir, ".cursor", "hooks.json")
		if fileExists(crewSettings) {
			files = append(files, staleSettingsInfo{
				path:        crewSettings,
				agentType:   "crew",
				rigName:     rigName,
				sessionName: "", // Shared settings, no single session
			})
		}
		if dirExists(crewDir) {
			crewEntries, _ := os.ReadDir(crewDir)
			for _, crewEntry := range crewEntries {
				if !crewEntry.IsDir() || crewEntry.Name() == ".cursor" {
					continue
				}
				crewWrongSettings := filepath.Join(crewDir, crewEntry.Name(), ".cursor", "hooks.json")
				if fileExists(crewWrongSettings) {
					files = append(files, staleSettingsInfo{
						path:          crewWrongSettings,
						agentType:     "crew",
						rigName:       rigName,
						sessionName:   fmt.Sprintf("gt-%s-crew-%s", rigName, crewEntry.Name()),
						wrongLocation: true,
					})
				}
			}
		}

		// Check for polecat settings - polecats/.cursor/ is correct (shared by all polecats, outside git repos)
		// Settings in polecats/<name>/.cursor/ are wrong (inside git repos)
		polecatsDir := filepath.Join(rigPath, "polecats")
		polecatsSettings := filepath.Join(polecatsDir, ".cursor", "hooks.json")
		if fileExists(polecatsSettings) {
			files = append(files, staleSettingsInfo{
				path:        polecatsSettings,
				agentType:   "polecat",
				rigName:     rigName,
				sessionName: "", // Shared settings, no single session
			})
		}
		if dirExists(polecatsDir) {
			polecatEntries, _ := os.ReadDir(polecatsDir)
			for _, pcEntry := range polecatEntries {
				if !pcEntry.IsDir() || pcEntry.Name() == ".cursor" {
					continue
				}
				pcWrongSettings := filepath.Join(polecatsDir, pcEntry.Name(), ".cursor", "hooks.json")
				if fileExists(pcWrongSettings) {
					files = append(files, staleSettingsInfo{
						path:          pcWrongSettings,
						agentType:     "polecat",
						rigName:       rigName,
						sessionName:   fmt.Sprintf("gt-%s-%s", rigName, pcEntry.Name()),
						wrongLocation: true,
					})
				}
			}
		}
	}

	return files
}

// checkSettings compares a settings file against the expected template.
// Returns a list of what's missing.
// agentType is reserved for future role-specific validation.
func (c *CursorSettingsCheck) checkSettings(path, _ string) []string {
	var missing []string

	// Read the actual settings
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{"unreadable"}
	}

	var actual map[string]any
	if err := json.Unmarshal(data, &actual); err != nil {
		return []string{"invalid JSON"}
	}

	// Check for required elements based on Cursor hooks.json template
	// All templates should have:
	// 1. version field
	// 2. hooks object with beforeSubmitPrompt and stop hooks

	// Check version
	if _, ok := actual["version"]; !ok {
		missing = append(missing, "version")
	}

	// Check hooks
	hooks, ok := actual["hooks"].(map[string]any)
	if !ok {
		return append(missing, "hooks")
	}

	// Check beforeSubmitPrompt hook exists (for mail check)
	if !c.hookHasCommand(hooks, "beforeSubmitPrompt") {
		missing = append(missing, "beforeSubmitPrompt hook")
	}

	// Check stop hook exists (for costs recording)
	if !c.hookHasCommand(hooks, "stop") {
		missing = append(missing, "stop hook")
	}

	return missing
}

// getGitFileStatus determines the git status of a file.
// Returns untracked, tracked-clean, tracked-modified, or unknown.
func (c *CursorSettingsCheck) getGitFileStatus(filePath string) gitFileStatus {
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	// Check if we're in a git repo
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return gitStatusUnknown
	}

	// Check if file is tracked
	cmd = exec.Command("git", "-C", dir, "ls-files", fileName)
	output, err := cmd.Output()
	if err != nil {
		return gitStatusUnknown
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		// File is not tracked
		return gitStatusUntracked
	}

	// File is tracked - check if modified
	cmd = exec.Command("git", "-C", dir, "diff", "--quiet", fileName)
	if err := cmd.Run(); err != nil {
		// Non-zero exit means file has changes
		return gitStatusTrackedModified
	}

	// Also check for staged changes
	cmd = exec.Command("git", "-C", dir, "diff", "--cached", "--quiet", fileName)
	if err := cmd.Run(); err != nil {
		return gitStatusTrackedModified
	}

	return gitStatusTrackedClean
}

// hookHasCommand checks if a hook type exists and has at least one command.
func (c *CursorSettingsCheck) hookHasCommand(hooks map[string]any, hookName string) bool {
	hookList, ok := hooks[hookName].([]any)
	if !ok || len(hookList) == 0 {
		return false
	}

	// Check that at least one hook has a command
	for _, hook := range hookList {
		hookMap, ok := hook.(map[string]any)
		if !ok {
			continue
		}
		if _, hasCommand := hookMap["command"]; hasCommand {
			return true
		}
	}
	return false
}

// Fix deletes stale settings files and restarts affected agents.
// Files with local modifications are skipped to avoid losing user changes.
func (c *CursorSettingsCheck) Fix(ctx *CheckContext) error {
	var errors []string
	var skipped []string
	t := tmux.NewTmux()

	for _, sf := range c.staleSettings {
		// Skip files with local modifications - require manual review
		if sf.wrongLocation && sf.gitStatus == gitStatusTrackedModified {
			skipped = append(skipped, fmt.Sprintf("%s: has local modifications, skipping", sf.path))
			continue
		}

		// Delete the stale settings file
		if err := os.Remove(sf.path); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", sf.path, err))
			continue
		}

		// Also delete parent .cursor directory if empty
		cursorDir := filepath.Dir(sf.path)
		_ = os.Remove(cursorDir) // Best-effort, will fail if not empty

		// For files in wrong locations, delete and create at correct location
		if sf.wrongLocation {
			mayorDir := filepath.Join(ctx.TownRoot, "mayor")

			// For mayor settings at town root, create at mayor/.cursor/
			if sf.agentType == "mayor" && strings.HasSuffix(cursorDir, ".cursor") && !strings.Contains(sf.path, "/mayor/") {
				if err := os.MkdirAll(mayorDir, 0755); err == nil {
					_ = cursor.EnsureSettingsForRole(mayorDir, "mayor")
				}
			}

			// For mayor CLAUDE.md at town root, create at mayor/
			if sf.agentType == "mayor" && strings.HasSuffix(sf.path, "CLAUDE.md") && !strings.Contains(sf.path, "/mayor/") {
				townName, _ := workspace.GetTownName(ctx.TownRoot)
				if err := templates.CreateMayorCLAUDEmd(
					mayorDir,
					ctx.TownRoot,
					townName,
					session.MayorSessionName(),
					session.DeaconSessionName(),
				); err != nil {
					errors = append(errors, fmt.Sprintf("failed to create mayor/CLAUDE.md: %v", err))
				}
			}

			// Town-root files were inherited by ALL agents via directory traversal.
			// Cycle all Gas Town sessions so they pick up the corrected file locations.
			// This includes gt-* (rig agents) and hq-* (mayor, deacon).
			sessions, _ := t.ListSessions()
			for _, sess := range sessions {
				if strings.HasPrefix(sess, session.Prefix) || strings.HasPrefix(sess, session.HQPrefix) {
					_ = t.KillSession(sess)
				}
			}
			continue
		}

		// Recreate settings using EnsureSettingsForRole
		workDir := filepath.Dir(cursorDir) // agent work directory
		if err := cursor.EnsureSettingsForRole(workDir, sf.agentType); err != nil {
			errors = append(errors, fmt.Sprintf("failed to recreate settings for %s: %v", sf.path, err))
			continue
		}

		// Only cycle patrol roles if --restart-sessions was explicitly passed.
		// This prevents unexpected session restarts during routine --fix operations.
		// Crew and polecats are spawned on-demand and won't auto-restart anyway.
		if ctx.RestartSessions {
			if sf.agentType == "witness" || sf.agentType == "refinery" ||
				sf.agentType == "deacon" || sf.agentType == "mayor" {
				running, _ := t.HasSession(sf.sessionName)
				if running {
					// Cycle the agent by killing and letting gt up restart it
					_ = t.KillSession(sf.sessionName)
				}
			}
		}
	}

	// Report skipped files as warnings, not errors
	if len(skipped) > 0 {
		for _, s := range skipped {
			fmt.Printf("  Warning: %s\n", s)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
