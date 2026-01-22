#!/bin/bash
set -euo pipefail

LOG_PATH="/tmp/cursor-hooks.log"
HOOKS_JSON="/Users/vasilistsolis/Documents/Github/cursor-gastown/.cursor/hooks.json"

rm -f "$LOG_PATH"

echo "Running hook command simulations from ${HOOKS_JSON}..."

run_hook_event() {
  local event="$1"
  local payload="$2"
  local hooks_dir
  hooks_dir="$(dirname "$HOOKS_JSON")"

  python3 - "$HOOKS_JSON" "$event" <<'PY'
import json
import sys

path, event = sys.argv[1], sys.argv[2]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)

hooks = data.get("hooks", {})
commands = hooks.get(event, [])
for entry in commands:
    command = entry.get("command")
    if command:
        print(command)
PY
}

run_and_echo() {
  local event="$1"
  local payload="$2"
  local hooks_dir
  hooks_dir="$(dirname "$HOOKS_JSON")"

  while IFS= read -r cmd; do
    if [[ -n "$cmd" ]]; then
      echo "-> ${event}: ${cmd}"
      (cd "$hooks_dir" && printf '%s' "$payload" | bash -c "$cmd")
    fi
  done < <(run_hook_event "$event" "$payload")
}

run_and_echo "sessionStart" '{"hook_event_name":"sessionStart"}'
run_and_echo "beforeSubmitPrompt" '{"hook_event_name":"beforeSubmitPrompt"}'
run_and_echo "stop" '{"hook_event_name":"stop","status":"completed","loop_count":0}'
run_and_echo "sessionEnd" '{"hook_event_name":"sessionEnd","reason":"completed"}'
run_and_echo "preCompact" '{"hook_event_name":"preCompact","trigger":"manual"}'

echo ""
echo "=== Hook log ==="
if [[ -f "$LOG_PATH" ]]; then
  cat "$LOG_PATH"
else
  echo "(no hook log file created)"
fi
