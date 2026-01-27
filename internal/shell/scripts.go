package shell

import (
	"fmt"
)

const ZshScript = `
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
`

const BashScript = `
# Bash support
# Bash support
bind -x '"\C-@": "READLINE_LINE=$(recall < /dev/tty); READLINE_POINT=${#READLINE_LINE}; history -s \"$READLINE_LINE\"; eval \"$READLINE_LINE\""'

recall() {
  if [[ "$1" == "save" ]]; then
    # Bash: history 2 | head -n 1 gets previous command. 
    # Check if 'fc' is available or use history expansion
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
`

func PrintInitScript(shellType string) {
	switch shellType {
	case "zsh":
		fmt.Println(ZshScript)
	case "bash":
		fmt.Println(BashScript)
	default:
		fmt.Printf("# Unsupported shell: %s. Defaulting to Zsh.\n%s", shellType, ZshScript)
	}
}
