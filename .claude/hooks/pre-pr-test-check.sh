#!/usr/bin/env bash
# .claude/hooks/pre-pr-test-check.sh
#
# PreToolUse hook, matched on the Bash tool. Fires before every Bash command
# the agent runs. Only acts when the command looks like a PR creation —
# everything else passes through untouched.
#
# Exit code 0 = allow the tool call to proceed.
# Exit code 2 = block it; stderr is fed back to the agent as the reason.

set -euo pipefail

input=$(cat)
command=$(echo "$input" | jq -r '.tool_input.command // empty')

# Only act on commands that look like PR creation. Everything else is a no-op.
if [[ "$command" != *"gh pr create"* && "$command" != *"git push"* ]]; then
  exit 0
fi

echo "PR creation detected — running go test ./... before allowing it." >&2

if ! go test ./... 2>&1 >&2; then
  echo "" >&2
  echo "Tests are failing. Fix them before creating the PR — do not skip or weaken a test to get past this check." >&2
  exit 2
fi

echo "All tests passed — allowing PR creation." >&2
exit 0
