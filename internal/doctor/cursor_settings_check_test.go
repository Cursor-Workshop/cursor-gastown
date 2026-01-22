package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCursorSettingsCheck(t *testing.T) {
	check := NewCursorSettingsCheck()

	if check.Name() != "cursor-settings" {
		t.Errorf("expected name 'cursor-settings', got %q", check.Name())
	}

	if !check.CanFix() {
		t.Error("expected CanFix to return true")
	}
}

func TestCursorSettingsCheck_NoSettingsFiles(t *testing.T) {
	tmpDir := t.TempDir()

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK when no settings files, got %v", result.Status)
	}
}

// createValidSettings creates a valid hooks.json with all required elements.
func createValidSettings(t *testing.T, path string) {
	t.Helper()

	settings := map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"beforeSubmitPrompt": []any{
				map[string]any{
					"command": ".cursor/hooks/gastown-prompt.sh",
				},
			},
			"stop": []any{
				map[string]any{
					"command": ".cursor/hooks/gastown-stop.sh",
				},
			},
		},
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

// createStaleSettings creates a hooks.json missing required elements.
func createStaleSettings(t *testing.T, path string, missingElements ...string) {
	t.Helper()

	settings := map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"beforeSubmitPrompt": []any{
				map[string]any{
					"command": ".cursor/hooks/gastown-prompt.sh",
				},
			},
			"stop": []any{
				map[string]any{
					"command": ".cursor/hooks/gastown-stop.sh",
				},
			},
		},
	}

	for _, missing := range missingElements {
		switch missing {
		case "version":
			delete(settings, "version")
		case "hooks":
			delete(settings, "hooks")
		case "beforeSubmitPrompt":
			hooks := settings["hooks"].(map[string]any)
			delete(hooks, "beforeSubmitPrompt")
		case "stop":
			hooks := settings["hooks"].(map[string]any)
			delete(hooks, "stop")
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCursorSettingsCheck_ValidMayorSettings(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid mayor settings at correct location (mayor/.cursor/hooks.json)
	// NOT at town root (.cursor/hooks.json) which is wrong location
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createValidSettings(t, mayorSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_ValidDeaconSettings(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid deacon settings
	deaconSettings := filepath.Join(tmpDir, "deacon", ".cursor", "hooks.json")
	createValidSettings(t, deaconSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid deacon settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_ValidWitnessSettings(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create valid witness settings in correct location (witness/.cursor/, outside git repo)
	witnessSettings := filepath.Join(tmpDir, rigName, "witness", ".cursor", "hooks.json")
	createValidSettings(t, witnessSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid witness settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_ValidRefinerySettings(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create valid refinery settings in correct location (refinery/.cursor/, outside git repo)
	refinerySettings := filepath.Join(tmpDir, rigName, "refinery", ".cursor", "hooks.json")
	createValidSettings(t, refinerySettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid refinery settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_ValidCrewSettings(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create valid crew settings in correct location (crew/.cursor/, shared by all crew)
	crewSettings := filepath.Join(tmpDir, rigName, "crew", ".cursor", "hooks.json")
	createValidSettings(t, crewSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid crew settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_ValidPolecatSettings(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create valid polecat settings in correct location (polecats/.cursor/, shared by all polecats)
	pcSettings := filepath.Join(tmpDir, rigName, "polecats", ".cursor", "hooks.json")
	createValidSettings(t, pcSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusOK {
		t.Errorf("expected StatusOK for valid polecat settings, got %v: %s", result.Status, result.Message)
	}
}

func TestCursorSettingsCheck_MissingVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stale mayor settings missing version (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createStaleSettings(t, mayorSettings, "version")

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for missing version, got %v", result.Status)
	}
	if !strings.Contains(result.Message, "1 stale") {
		t.Errorf("expected message about stale settings, got %q", result.Message)
	}
}

func TestCursorSettingsCheck_MissingHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stale settings missing hooks entirely (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createStaleSettings(t, mayorSettings, "hooks")

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for missing hooks, got %v", result.Status)
	}
}

func TestCursorSettingsCheck_MissingBeforeSubmitPrompt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stale settings missing beforeSubmitPrompt hook (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createStaleSettings(t, mayorSettings, "beforeSubmitPrompt")

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for missing beforeSubmitPrompt, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "beforeSubmitPrompt") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention beforeSubmitPrompt hook, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_MissingStopHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stale settings missing stop hook (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createStaleSettings(t, mayorSettings, "stop")

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for missing stop hook, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "stop hook") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention stop hook, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_WrongLocationWitness(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create settings in wrong location (witness/rig/.cursor/ instead of witness/.cursor/)
	// Settings inside git repos should be flagged as wrong location
	wrongSettings := filepath.Join(tmpDir, rigName, "witness", "rig", ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "wrong location") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention wrong location, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_WrongLocationRefinery(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create settings in wrong location (refinery/rig/.cursor/ instead of refinery/.cursor/)
	// Settings inside git repos should be flagged as wrong location
	wrongSettings := filepath.Join(tmpDir, rigName, "refinery", "rig", ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "wrong location") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention wrong location, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_MultipleStaleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create multiple stale settings files (all at correct locations)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createStaleSettings(t, mayorSettings, "beforeSubmitPrompt")

	deaconSettings := filepath.Join(tmpDir, "deacon", ".cursor", "hooks.json")
	createStaleSettings(t, deaconSettings, "stop")

	// Settings inside git repo (witness/rig/.cursor/) are wrong location
	witnessWrong := filepath.Join(tmpDir, rigName, "witness", "rig", ".cursor", "hooks.json")
	createValidSettings(t, witnessWrong) // Valid content but wrong location

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for multiple stale files, got %v", result.Status)
	}
	if !strings.Contains(result.Message, "3 stale") {
		t.Errorf("expected message about 3 stale files, got %q", result.Message)
	}
}

func TestCursorSettingsCheck_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid JSON file (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	if err := os.MkdirAll(filepath.Dir(mayorSettings), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mayorSettings, []byte("not valid json {"), 0644); err != nil {
		t.Fatal(err)
	}

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for invalid JSON, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "invalid JSON") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention invalid JSON, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_FixDeletesStaleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create stale settings in wrong location (inside git repo - easy to test - just delete, no recreate)
	rigName := "testrig"
	wrongSettings := filepath.Join(tmpDir, rigName, "witness", "rig", ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	// Run to detect
	result := check.Run(ctx)
	if result.Status != StatusError {
		t.Fatalf("expected StatusError before fix, got %v", result.Status)
	}

	// Apply fix
	if err := check.Fix(ctx); err != nil {
		t.Fatalf("Fix failed: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(wrongSettings); !os.IsNotExist(err) {
		t.Error("expected wrong location settings to be deleted")
	}

	// Verify check passes (no settings files means OK)
	result = check.Run(ctx)
	if result.Status != StatusOK {
		t.Errorf("expected StatusOK after fix, got %v", result.Status)
	}
}

func TestCursorSettingsCheck_SkipsNonRigDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories that should be skipped
	for _, skipDir := range []string{"mayor", "deacon", "daemon", ".git", "docs", ".hidden"} {
		dir := filepath.Join(tmpDir, skipDir, "witness", "rig", ".cursor")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		// These should NOT be detected as rig witness settings
		settingsPath := filepath.Join(dir, "hooks.json")
		createStaleSettings(t, settingsPath, "beforeSubmitPrompt")
	}

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	_ = check.Run(ctx)

	// Should only find mayor and deacon settings in their specific locations
	// The witness settings in these dirs should be ignored
	// Since we didn't create valid mayor/deacon settings, those will be stale
	// But the ones in "mayor/witness/rig/.cursor" should be ignored

	// Count how many stale files were found - should be 0 since none of the
	// skipped directories have their settings detected
	if len(check.staleSettings) != 0 {
		t.Errorf("expected 0 stale files (skipped dirs), got %d", len(check.staleSettings))
	}
}

func TestCursorSettingsCheck_MixedValidAndStale(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create valid mayor settings (at correct location)
	mayorSettings := filepath.Join(tmpDir, "mayor", ".cursor", "hooks.json")
	createValidSettings(t, mayorSettings)

	// Create stale witness settings in correct location (missing beforeSubmitPrompt)
	witnessSettings := filepath.Join(tmpDir, rigName, "witness", ".cursor", "hooks.json")
	createStaleSettings(t, witnessSettings, "beforeSubmitPrompt")

	// Create valid refinery settings in correct location
	refinerySettings := filepath.Join(tmpDir, rigName, "refinery", ".cursor", "hooks.json")
	createValidSettings(t, refinerySettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for mixed valid/stale, got %v", result.Status)
	}
	if !strings.Contains(result.Message, "1 stale") {
		t.Errorf("expected message about 1 stale file, got %q", result.Message)
	}
	// Should only report the witness settings as stale
	if len(result.Details) != 1 {
		t.Errorf("expected 1 detail, got %d: %v", len(result.Details), result.Details)
	}
}

func TestCursorSettingsCheck_WrongLocationCrew(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create settings in wrong location (crew/<name>/.cursor/ instead of crew/.cursor/)
	// Settings inside git repos should be flagged as wrong location
	wrongSettings := filepath.Join(tmpDir, rigName, "crew", "agent1", ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "wrong location") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention wrong location, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_WrongLocationPolecat(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create settings in wrong location (polecats/<name>/.cursor/ instead of polecats/.cursor/)
	// Settings inside git repos should be flagged as wrong location
	wrongSettings := filepath.Join(tmpDir, rigName, "polecats", "pc1", ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "wrong location") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention wrong location, got %v", result.Details)
	}
}

// initTestGitRepo initializes a git repo in the given directory for settings tests.
func initTestGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
}

// gitAddAndCommit adds and commits a file.
func gitAddAndCommit(t *testing.T, repoDir, filePath string) {
	t.Helper()
	// Get relative path from repo root
	relPath, err := filepath.Rel(repoDir, filePath)
	if err != nil {
		t.Fatal(err)
	}

	cmds := [][]string{
		{"git", "add", relPath},
		{"git", "commit", "-m", "Add file"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
}

func TestCursorSettingsCheck_GitStatusUntracked(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create an untracked settings file (not git added)
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	// Should mention "untracked"
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "untracked") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention untracked, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_GitStatusTrackedClean(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create settings and commit it (tracked, clean)
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)
	gitAddAndCommit(t, rigDir, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	// Should mention "tracked but unmodified"
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "tracked but unmodified") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention tracked but unmodified, got %v", result.Details)
	}
}

func TestCursorSettingsCheck_GitStatusTrackedModified(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create settings and commit it
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)
	gitAddAndCommit(t, rigDir, wrongSettings)

	// Modify the file after commit
	if err := os.WriteFile(wrongSettings, []byte(`{"modified": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	result := check.Run(ctx)

	if result.Status != StatusError {
		t.Errorf("expected StatusError for wrong location, got %v", result.Status)
	}
	// Should mention "local modifications"
	found := false
	for _, d := range result.Details {
		if strings.Contains(d, "local modifications") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected details to mention local modifications, got %v", result.Details)
	}
	// Should also mention manual review
	if !strings.Contains(result.FixHint, "manual review") {
		t.Errorf("expected fix hint to mention manual review, got %q", result.FixHint)
	}
}

func TestCursorSettingsCheck_FixSkipsModifiedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create settings and commit it
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)
	gitAddAndCommit(t, rigDir, wrongSettings)

	// Modify the file after commit
	if err := os.WriteFile(wrongSettings, []byte(`{"modified": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	// Run to detect
	result := check.Run(ctx)
	if result.Status != StatusError {
		t.Fatalf("expected StatusError before fix, got %v", result.Status)
	}

	// Apply fix - should NOT delete the modified file
	if err := check.Fix(ctx); err != nil {
		t.Fatalf("Fix failed: %v", err)
	}

	// Verify file still exists (was skipped)
	if _, err := os.Stat(wrongSettings); os.IsNotExist(err) {
		t.Error("expected modified file to be preserved, but it was deleted")
	}
}

func TestCursorSettingsCheck_FixDeletesUntrackedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create an untracked settings file (not git added)
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	// Run to detect
	result := check.Run(ctx)
	if result.Status != StatusError {
		t.Fatalf("expected StatusError before fix, got %v", result.Status)
	}

	// Apply fix - should delete the untracked file
	if err := check.Fix(ctx); err != nil {
		t.Fatalf("Fix failed: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(wrongSettings); !os.IsNotExist(err) {
		t.Error("expected untracked file to be deleted")
	}
}

func TestCursorSettingsCheck_FixDeletesTrackedCleanFiles(t *testing.T) {
	tmpDir := t.TempDir()
	rigName := "testrig"

	// Create a git repo to simulate a source repo
	rigDir := filepath.Join(tmpDir, rigName, "witness", "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, rigDir)

	// Create settings and commit it (tracked, clean)
	wrongSettings := filepath.Join(rigDir, ".cursor", "hooks.json")
	createValidSettings(t, wrongSettings)
	gitAddAndCommit(t, rigDir, wrongSettings)

	check := NewCursorSettingsCheck()
	ctx := &CheckContext{TownRoot: tmpDir}

	// Run to detect
	result := check.Run(ctx)
	if result.Status != StatusError {
		t.Fatalf("expected StatusError before fix, got %v", result.Status)
	}

	// Apply fix - should delete the tracked clean file
	if err := check.Fix(ctx); err != nil {
		t.Fatalf("Fix failed: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(wrongSettings); !os.IsNotExist(err) {
		t.Error("expected tracked clean file to be deleted")
	}
}
