// Package agent provides unified agent settings management.
package agent

import (
	"github.com/cursorworkshop/cursor-gastown/internal/config"
	"github.com/cursorworkshop/cursor-gastown/internal/cursor"
)

// EnsureSettingsForRole ensures agent settings exist for the given agent preset and role.
// This is a unified function that delegates to the appropriate agent-specific implementation.
//
// For Cursor: Creates .cursor/rules/gastown.mdc with rules and .cursor/hooks.json
// For other agents: Currently no-op (may be extended in future)
func EnsureSettingsForRole(workDir, role string, agentName string) error {
	// If no agent specified, default to cursor
	if agentName == "" {
		agentName = "cursor"
	}

	preset := config.GetAgentPresetByName(agentName)
	if preset == nil {
		// Unknown agent, use cursor as fallback
		return cursor.EnsureSettingsForRole(workDir, role)
	}

	switch preset.Name {
	case config.AgentCursor:
		return cursor.EnsureSettingsForRole(workDir, role)
	case config.AgentGemini, config.AgentCodex, config.AgentAuggie, config.AgentAmp:
		// These agents don't have a similar settings/rules mechanism yet
		// They may read AGENTS.md or have their own config
		return nil
	default:
		// Unknown preset, use cursor as fallback
		return cursor.EnsureSettingsForRole(workDir, role)
	}
}

// EnsureSettingsForAllAgents ensures settings exist for all supported agents.
// This is useful during installation to prepare the workspace for any agent.
func EnsureSettingsForAllAgents(workDir, role string) error {
	// Ensure Cursor rules and hooks
	return cursor.EnsureSettingsForRole(workDir, role)
}
