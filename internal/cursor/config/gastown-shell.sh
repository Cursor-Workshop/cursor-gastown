#!/bin/bash
# Gas Town shell execution hooks for Cursor
#
# Usage: gastown-shell.sh [before|after]
#
# beforeShellExecution: Called before shell commands run
#   Input:  {"command": "...", "cwd": "..."}
#   Output: {"permission": "allow"|"deny"|"ask", "user_message": "...", "agent_message": "..."}
#
# afterShellExecution: Called after shell commands complete
#   Input:  {"command": "...", "output": "...", "duration": N}
#   Output: (none expected, fire-and-forget)

HOOK_PHASE="${1:-after}"

# Read JSON input from stdin (required - must consume it)
input=$(cat)

# Export PATH to ensure gt is available
export PATH="$HOME/go/bin:$HOME/bin:$HOME/.local/bin:$PATH"

# Session state directory
STATE_DIR="/tmp/gastown-session-${GT_SESSION_ID:-$$}"

#--- BEFORE SHELL EXECUTION ---#
handle_before() {
    # Skip if not in Gas Town context
    if [ -z "$GT_ROLE" ]; then
        output_permission
        return
    fi

    # CLI PATHWAY: Mail injection on first command
    # (IDE uses beforeSubmitPrompt instead)
    if [ ! -f "$STATE_DIR/mail-checked" ]; then
        mkdir -p "$STATE_DIR"
        touch "$STATE_DIR/mail-checked"
        gt mail check --inject >/dev/null 2>&1 &
    fi

    output_permission
}

#--- AFTER SHELL EXECUTION ---#
handle_after() {
    # Skip if not in Gas Town context
    if [ -z "$GT_ROLE" ]; then
        exit 0
    fi

    # BOTH PATHWAYS: Audit logging (when GT_DEBUG set)
    if [ -n "$GT_DEBUG" ]; then
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        echo "[$timestamp] $input" >> /tmp/gastown-audit.log
    fi

    # CLI PATHWAY: Periodic cost recording
    # (IDE uses stop hook instead)
    mkdir -p "$STATE_DIR"
    count=$(cat "$STATE_DIR/cmd-count" 2>/dev/null || echo "0")
    count=$((count + 1))
    echo "$count" > "$STATE_DIR/cmd-count"
    
    # Record costs every 10 commands in CLI mode
    if [ $((count % 10)) -eq 0 ]; then
        gt costs record >/dev/null 2>&1 &
    fi

    exit 0
}

#--- OUTPUT HELPERS ---#
output_permission() {
    cat << 'EOF'
{
  "permission": "allow"
}
EOF
}

#--- MAIN ---#
case "$HOOK_PHASE" in
    before)
        # Log if debugging
        if [ -n "$GT_DEBUG" ]; then
            cmd=$(echo "$input" | grep -o '"command":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "?")
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] beforeShell: $cmd" >> /tmp/gastown-hooks.log
        fi
        
        handle_before
        ;;
    after)
        # Log if debugging
        if [ -n "$GT_DEBUG" ]; then
            cmd=$(echo "$input" | grep -o '"command":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "?")
            duration=$(echo "$input" | grep -o '"duration":[0-9]*' | cut -d':' -f2 2>/dev/null || echo "?")
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] afterShell: $cmd (${duration}ms)" >> /tmp/gastown-hooks.log
        fi
        
        handle_after
        ;;
    *)
        echo "Usage: $0 [before|after]" >&2
        exit 1
        ;;
esac
