package shell

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

const zshAlias = `function ws() {
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
`

const bashAlias = zshAlias // identical for bash

const fishAlias = `function ws
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
`

func NewCmdShellAlias(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "alias [bash|zsh|fish]",
		Short: "emit ws and wt shell function definitions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := resolveShell(args)
			if err != nil {
				return err
			}
			switch shell {
			case "zsh":
				fmt.Fprint(f.IO.Out, zshAlias)
			case "bash":
				fmt.Fprint(f.IO.Out, bashAlias)
			case "fish":
				fmt.Fprint(f.IO.Out, fishAlias)
			default:
				fmt.Fprintf(f.IO.ErrOut, "unsupported shell: %s\n", shell)
				return errors.New("unsupported shell")
			}
			return nil
		},
	}
}
