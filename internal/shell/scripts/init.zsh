# Recall Zsh Integration
recall_widget() {
  local cmd
  # Ensure we call the binary, not the function wrapper
  cmd=$(command recall < /dev/tty)
  if [[ -n "$cmd" ]]; then
    LBUFFER="$cmd"
    zle accept-line
  fi
  zle reset-prompt
}

zle -N recall_widget
# Standard Ctrl+Space bindings
bindkey '^@' recall_widget
bindkey '^ ' recall_widget

# Ensure it works in standard keymaps
bindkey -M emacs '^@' recall_widget
bindkey -M viins '^@' recall_widget
bindkey -M vicmd '^@' recall_widget

# Fallback: Ctrl+R (Replace standard history search)
bindkey -M emacs '^R' recall_widget
bindkey -M viins '^R' recall_widget
bindkey -M vicmd '^R' recall_widget
bindkey '^R' recall_widget

echo "Recall Shell Integration Loaded"

# Wrapper to execute command when running 'recall' manually
recall() {
  if [[ "$1" == "save" ]]; then
    # Zsh: fc -ln -1 gets the last command from history
    local last_cmd=$(fc -ln -1)
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
    print -s -- "$cmd"
    eval -- "$cmd"
  fi
}
