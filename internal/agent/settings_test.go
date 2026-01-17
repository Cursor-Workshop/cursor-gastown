package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSettingsForRole_Cursor(t *testing.T) {
	tmpDir := t.TempDir()

	err := EnsureSettingsForRole(tmpDir, "polecat", "cursor")
	if err != nil {
		t.Fatalf("EnsureSettingsForRole failed: %v", err)
	}

	// Check Cursor rules were created
	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); os.IsNotExist(err) {
		t.Error("Cursor rules not created")
	}

	// Check Cursor hooks were created
	cursorHooks := filepath.Join(tmpDir, ".cursor", "hooks.json")
	if _, err := os.Stat(cursorHooks); os.IsNotExist(err) {
		t.Error("Cursor hooks.json not created")
	}
}

func TestEnsureSettingsForRole_DefaultsToCursor(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty agent name should default to cursor
	err := EnsureSettingsForRole(tmpDir, "polecat", "")
	if err != nil {
		t.Fatalf("EnsureSettingsForRole failed: %v", err)
	}

	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); os.IsNotExist(err) {
		t.Error("Cursor rules not created for empty agent name")
	}
}

func TestEnsureSettingsForRole_UnknownAgent(t *testing.T) {
	tmpDir := t.TempDir()

	// Unknown agent should fall back to cursor
	err := EnsureSettingsForRole(tmpDir, "polecat", "unknown-agent")
	if err != nil {
		t.Fatalf("EnsureSettingsForRole failed: %v", err)
	}

	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); os.IsNotExist(err) {
		t.Error("Cursor rules not created for unknown agent")
	}
}

func TestEnsureSettingsForRole_Claude(t *testing.T) {
	tmpDir := t.TempDir()

	// Claude agent should be a no-op (we migrated to Cursor)
	err := EnsureSettingsForRole(tmpDir, "polecat", "claude")
	if err != nil {
		t.Fatalf("EnsureSettingsForRole failed: %v", err)
	}

	// Claude doesn't have settings anymore - this should be a no-op
	// (Claude is listed as an agent but doesn't create settings files)
	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); !os.IsNotExist(err) {
		t.Error("Cursor rules should not be created for Claude agent")
	}
}

func TestEnsureSettingsForRole_Gemini(t *testing.T) {
	tmpDir := t.TempDir()

	// Gemini doesn't have settings yet, should be a no-op
	err := EnsureSettingsForRole(tmpDir, "polecat", "gemini")
	if err != nil {
		t.Fatalf("EnsureSettingsForRole failed: %v", err)
	}

	// Neither settings should be created for Gemini
	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); !os.IsNotExist(err) {
		t.Error("Cursor rules should not be created for Gemini")
	}
}

func TestEnsureSettingsForAllAgents(t *testing.T) {
	tmpDir := t.TempDir()

	err := EnsureSettingsForAllAgents(tmpDir, "polecat")
	if err != nil {
		t.Fatalf("EnsureSettingsForAllAgents failed: %v", err)
	}

	// Only Cursor settings should be created (we migrated to Cursor-only)
	cursorRules := filepath.Join(tmpDir, ".cursor", "rules", "gastown.mdc")
	if _, err := os.Stat(cursorRules); os.IsNotExist(err) {
		t.Error("Cursor rules not created")
	}

	cursorHooks := filepath.Join(tmpDir, ".cursor", "hooks.json")
	if _, err := os.Stat(cursorHooks); os.IsNotExist(err) {
		t.Error("Cursor hooks.json not created")
	}
}
