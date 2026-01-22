#!/bin/bash
# Gas Town beforeSubmitPrompt hook for Cursor
#
# Called right after user hits send but before backend request.
# This hook can block submission but cannot inject context.
# Use sessionStart for context injection.
#
# Input:  {"prompt": "...", "attachments": [...]}
# Output: {"continue": true|false, "user_message": "..."}

set -e

# Read JSON input from stdin (required by Cursor hooks protocol)
json_input=$(cat)

# Export PATH to ensure gt is available
export PATH="$HOME/go/bin:$HOME/bin:$HOME/.local/bin:$PATH"

# Only run if we're in a Gas Town context (GT_ROLE is set)
if [ -n "$GT_ROLE" ]; then
    # Check for mail and inject into context
    # Run in background to not block the prompt
    gt mail check --inject >/dev/null 2>&1 &
fi

# Always allow the prompt to continue
# Context injection happens at sessionStart, not here
echo '{"continue": true}'
