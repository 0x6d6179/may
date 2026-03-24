package shell

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

const zshShim = `
function may() {
  local _may_out
  _may_out=$(\command may "$@")
  if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
    \builtin cd -- "$_may_out"
  elif [[ -n "$_may_out" ]]; then
    printf '%s\n' "$_may_out"
  fi
}

function ws() { may ws "$@"; }
function wt() { may wt "$@"; }

autoload -Uz add-zsh-hook
function _may_id_hook() { \command may id status --apply --quiet }
chpwd_functions=("${(@)chpwd_functions:#_may_id_hook}")
chpwd_functions+=(_may_id_hook)

eval "$(\command may shell completion zsh)"
`

const bashShim = `
function may() {
  local _may_out
  _may_out=$(\command may "$@")
  if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
    \builtin cd -- "$_may_out"
  elif [[ -n "$_may_out" ]]; then
    printf '%s\n' "$_may_out"
  fi
}

function ws() { may ws "$@"; }
function wt() { may wt "$@"; }

if [[ ${PROMPT_COMMAND:-} != *'_may_id_hook'* ]]; then
  function _may_id_hook() { \command may id status --apply --quiet; }
  PROMPT_COMMAND="${PROMPT_COMMAND:+${PROMPT_COMMAND}; }_may_id_hook"
fi

eval "$(\command may shell completion bash)"
`

const fishShim = `
function may
  set _may_out (\command may $argv)
  if test -n "$_may_out" -a -d "$_may_out"
    builtin cd "$_may_out"
  else if test -n "$_may_out"
    printf '%s\n' "$_may_out"
  end
end

function ws; may ws $argv; end
function wt; may wt $argv; end

functions --erase _may_id_hook 2>/dev/null
function --on-variable PWD _may_id_hook
  \command may id status --apply --quiet
end

\command may shell completion fish | source
`

func NewCmdShellInit(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "init [bash|zsh|fish]",
		Short: "emit a shell integration shim",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := resolveShell(args)
			if err != nil {
				return err
			}
			switch shell {
			case "zsh":
				_, err = io.WriteString(f.IO.Out, zshShim)
			case "bash":
				_, err = io.WriteString(f.IO.Out, bashShim)
			case "fish":
				_, err = io.WriteString(f.IO.Out, fishShim)
			default:
				fmt.Fprintf(f.IO.ErrOut, "unsupported shell: %s\n", shell)
				return errors.New("unsupported shell")
			}
			return err
		},
	}
}
