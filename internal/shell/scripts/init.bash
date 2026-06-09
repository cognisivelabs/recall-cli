# Recall Bash Integration
bind -x '"\C-@": "READLINE_LINE=$(recall < /dev/tty); READLINE_POINT=${#READLINE_LINE}; history -s \"$READLINE_LINE\"; eval \"$READLINE_LINE\""'

recall() {
  if [[ "$1" == "save" ]]; then
    # Bash: history 2 | head -n 1 gets previous command
    local last_cmd=$(history 2 | head -n 1 | sed 's/^[ ]*[0-9]*[ ]*//')
    command recall save --last-cmd "$last_cmd"
    return
  fi

  if [[ $# -gt 0 ]]; then
    command recall "$@"
    return
  fi

  local cmd
  cmd=$(command recall < /dev/tty)
  if [[ -n "$cmd" ]]; then
    history -s "$cmd"
    eval "$cmd"
  fi
}
