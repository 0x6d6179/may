package shell

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

const zshShim = `if [[ -z "$_MAY_SHELL_INIT" ]]; then
  export _MAY_SHELL_INIT=1

  function ws() {
    local _may_out
    _may_out=$(may ws "$@")
    if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
      builtin cd "$_may_out"
    fi
  }

  function wt() {
    local _may_out
    _may_out=$(may wt "$@")
    if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
      builtin cd "$_may_out"
    fi
  }

  autoload -Uz add-zsh-hook
  function _may_id_hook() { may id status --apply --quiet }
  add-zsh-hook chpwd _may_id_hook

  eval "$(may shell completion zsh)"
fi
`

const bashShim = `if [[ -z "$_MAY_SHELL_INIT" ]]; then
  export _MAY_SHELL_INIT=1

  function ws() {
    local _may_out
    _may_out=$(may ws "$@")
    if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
      builtin cd "$_may_out"
    fi
  }

  function wt() {
    local _may_out
    _may_out=$(may wt "$@")
    if [[ -n "$_may_out" ]] && [[ -d "$_may_out" ]]; then
      builtin cd "$_may_out"
    fi
  }

  PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND;}may id status --apply --quiet"

  eval "$(may shell completion bash)"
fi
`

const fishShim = `if not set -q _MAY_SHELL_INIT
  set -gx _MAY_SHELL_INIT 1

  function ws
    set _may_out (may ws $argv)
    if test -n "$_may_out" -a -d "$_may_out"
      builtin cd "$_may_out"
    end
  end

  function wt
    set _may_out (may wt $argv)
    if test -n "$_may_out" -a -d "$_may_out"
      builtin cd "$_may_out"
    end
  end

  function --on-variable PWD _may_id_hook
    may id status --apply --quiet
  end

  may shell completion fish | source
end
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
				fmt.Fprint(f.IO.Out, zshShim)
			case "bash":
				fmt.Fprint(f.IO.Out, bashShim)
			case "fish":
				fmt.Fprint(f.IO.Out, fishShim)
			default:
				fmt.Fprintf(f.IO.ErrOut, "unsupported shell: %s\n", shell)
				return errors.New("unsupported shell")
			}
			return nil
		},
	}
}
