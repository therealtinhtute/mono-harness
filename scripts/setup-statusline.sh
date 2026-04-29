#!/usr/bin/env bash
# setup-statusline.sh — Install the @slim statusline for Claude Code

set -euo pipefail

DEST="${HOME}/.claude/statusline.sh"
BACKUP="${DEST}.backup.$(date +%s)"

echo "Installing slim statusline..."

# Backup existing
if [ -f "$DEST" ]; then
  cp "$DEST" "$BACKUP"
  echo "  Backed up existing statusline to ${BACKUP}"
fi

# Write the statusline script
cat > "$DEST" << 'STATUSLINE_EOF'
#!/bin/sh
# slim.sh — model-first layout: ✦ model  ▰▱ N%  ϟ tpm  ⌥ branch  +N -N

TPM_STATE_PREFIX="claude-code-statusline-tpm"
TPM_WINDOW_MS=300000
MODEL_STATE_PREFIX="claude-code-statusline-model"
SUBAGENT_STATE_PREFIX="claude-code-statusline-subagent"

input=$(cat)

eval "$(echo "$input" | jq -r '
  "cwd=\(.cwd // "" | @sh)",
  "session_id=\(.session_id // "" | @sh)",
  "transcript_path=\(.transcript_path // "" | @sh)",
  "used=\(.context_window.used_percentage // 0 | floor | @sh)",
  "model=\(.model.display_name // "unknown" | @sh)",
  "total_in=\(.context_window.total_input_tokens // 0 | floor | @sh)",
  "total_out=\(.context_window.total_output_tokens // 0 | floor | @sh)",
  "duration_ms=\(.cost.total_duration_ms // 0 | floor | @sh)"
')"

used=${used:-0}; model=${model:-unknown}
total_in=${total_in:-0}; total_out=${total_out:-0}; duration_ms=${duration_ms:-0}

safe_id=$(printf '%s' "$session_id" | tr -dc 'a-zA-Z0-9_-')
total_tokens=$((total_in + total_out))

# TPM calculation
tpm=0
if [ "$duration_ms" -gt 0 ]; then
  tpm=$(( (total_tokens * 60000) / duration_ms ))
fi

# Session-scoped model cache
if [ -n "$safe_id" ]; then
  model_file="/tmp/${MODEL_STATE_PREFIX}-${safe_id}"
  if [ -f "$model_file" ]; then
    cached_model=$(head -1 "$model_file" 2>/dev/null)
    cached_duration=$(tail -1 "$model_file" 2>/dev/null | grep -E '^[0-9]+$' || echo 0)
    if [ "$duration_ms" -lt "$cached_duration" ] 2>/dev/null; then
      [ "$model" != "unknown" ] && printf '%s\n%s\n' "$model" "$duration_ms" > "$model_file"
    elif [ "$duration_ms" -gt "$cached_duration" ] 2>/dev/null && [ "$model" != "unknown" ]; then
      printf '%s\n%s\n' "$model" "$duration_ms" > "$model_file"
    else
      [ -n "$cached_model" ] && model="$cached_model"
    fi
  elif [ "$model" != "unknown" ]; then
    printf '%s\n%s\n' "$model" "$duration_ms" > "$model_file"
  fi
fi

# Progress bar
filled=$((used / 20))
[ "$used" -ge 95 ] && filled=5
[ "$filled" -gt 5 ] && filled=5
case $filled in
  0) bar="▱▱▱▱▱" ;; 1) bar="▰▱▱▱▱" ;; 2) bar="▰▰▱▱▱" ;; 3) bar="▰▰▰▱▱" ;; 4) bar="▰▰▰▰▱" ;; 5) bar="▰▰▰▰▰" ;;
esac

# Context color
if [ "$used" -ge 75 ]; then
  ctx_color="\033[38;5;208m"
elif [ "$used" -ge 50 ]; then
  ctx_color="\033[93m"
elif [ "$used" -ge 35 ]; then
  ctx_color="\033[38;5;148m"
else
  ctx_color="\033[38;5;247m"
fi

dim="\033[38;5;247m"
reset="\033[0m"
sep="  "

# Git branch
branch=""
if [ -n "$cwd" ]; then
  branch=$(git --no-optional-locks -C "$cwd" rev-parse --abbrev-ref HEAD 2>/dev/null)
  [ "$branch" = "HEAD" ] && branch=$(git --no-optional-locks -C "$cwd" rev-parse --short HEAD 2>/dev/null)
  [ -z "$branch" ] && branch=$(git --no-optional-locks -C "$cwd" symbolic-ref --short HEAD 2>/dev/null)
fi

# Output
printf "\033[38;5;252m✦ %s${reset}" "$model"
printf "${sep}${ctx_color}%s %s%%${reset}" "$bar" "$used"
if [ "$tpm" -gt 0 ]; then
  if [ "$tpm" -ge 10000 ]; then
    tpm_display="$((tpm / 1000))k"
  elif [ "$tpm" -ge 1000 ]; then
    tpm_display="$((tpm / 1000)).$((tpm % 1000 / 100))k"
  else
    tpm_display="$tpm"
  fi
  printf "${sep}${dim}ϟ %s tpm${reset}" "$tpm_display"
fi
[ -n "$branch" ] && printf "${sep}\033[36m⌥ %s${reset}" "$branch"
printf "\n"
STATUSLINE_EOF

chmod +x "$DEST"

echo "  Installed to ${DEST}"
echo ""
echo "Add this to your ~/.claude/settings.json:"
echo '  "statusLine": { "type": "command", "command": "bash ~/.claude/statusline.sh" }'
echo ""
echo "Requires jq (brew install jq)."
