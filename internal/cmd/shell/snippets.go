package shell

import (
	"fmt"
	"strings"
)

const (
	featureCore       = "core"
	featureWS         = "ws"
	featureWT         = "wt"
	featureAI         = "ai"
	featureAIFix      = "aifix"
	featureJ          = "j"
	featureSSHM       = "sshm"
	featureCompletion = "completion"
	featureDev        = "dev"
)

type AliasEntry struct {
	Name    string
	Command string
}

func BuildSnippet(shell string, features []string, devPath string, aliases ...AliasEntry) string {
	set := make(map[string]bool, len(features))
	for _, f := range features {
		set[f] = true
	}

	var b strings.Builder

	if set[featureDev] && devPath != "" {
		switch shell {
		case "fish":
			fmt.Fprintf(&b, "set -gx PATH %s $PATH\n", devPath)
		default:
			fmt.Fprintf(&b, "export PATH=%q:$PATH\n", devPath)
		}
		b.WriteString("# warning: dev mode — remove when not developing may\n")
	}

	if set[featureCore] {
		b.WriteString(mayFunctionSnippet(shell))
		b.WriteString("\n\n")
	}
	if set[featureWS] {
		b.WriteString(wsSnippet(shell))
		b.WriteString("\n")
	}
	if set[featureWT] {
		b.WriteString(wtSnippet(shell))
		b.WriteString("\n")
	}
	if set[featureAI] {
		b.WriteString(aiSnippet(shell))
		b.WriteString("\n\n")
	}
	if set[featureJ] {
		b.WriteString(jSnippet(shell))
		b.WriteString("\n")
		b.WriteString(kSnippet(shell))
		b.WriteString("\n")
	}
	if set[featureSSHM] {
		b.WriteString(sshmSnippet(shell))
		b.WriteString("\n")
	}
	if set[featureAIFix] {
		b.WriteString(aiFixSnippet(shell))
		b.WriteString("\n\n")
	}
	if set[featureCore] {
		b.WriteString(idHookSnippet(shell))
		b.WriteString("\n\n")
	}
	for _, a := range aliases {
		b.WriteString(userAliasSnippet(shell, a.Name, a.Command))
		b.WriteString("\n")
	}
	if set[featureCompletion] {
		b.WriteString(completionSnippet(shell))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func mayFunctionSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function may\n" +
			"  set _may_out (\\command may $argv)\n" +
			"  if test -n \"$_may_out\" -a -d \"$_may_out\"\n" +
			"    builtin cd \"$_may_out\"\n" +
			"  else if test -n \"$_may_out\"\n" +
			"    printf '%s\\n' \"$_may_out\"\n" +
			"  end\n" +
			"end"
	default:
		return "function may() {\n" +
			"  local _may_out\n" +
			"  _may_out=$(\\command may \"$@\")\n" +
			"  if [[ -n \"$_may_out\" ]] && [[ -d \"$_may_out\" ]]; then\n" +
			"    \\builtin cd -- \"$_may_out\"\n" +
			"  elif [[ -n \"$_may_out\" ]]; then\n" +
			"    printf '%s\\n' \"$_may_out\"\n" +
			"  fi\n" +
			"}"
	}
}

func wsSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function ws; may ws $argv; end"
	default:
		return "function ws() { may ws \"$@\"; }"
	}
}

func wtSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function wt; may wt $argv; end"
	default:
		return "function wt() { may wt \"$@\"; }"
	}
}

func aiSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function ai; may ai $argv; end"
	default:
		return "function ai() { may ai \"$@\"; }"
	}
}

func jSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function j; may j $argv; end"
	default:
		return "function j() { may j \"$@\"; }"
	}
}

func kSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function k; prevd; end"
	default:
		return "function k() { \\builtin cd - ; }"
	}
}

func userAliasSnippet(shell, name, command string) string {
	switch shell {
	case "fish":
		return fmt.Sprintf("function %s; may %s $argv; end", name, command)
	default:
		return fmt.Sprintf("function %s() { may %s \"$@\"; }", name, command)
	}
}

func sshmSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function sshm; may sshm $argv; end"
	default:
		return "function sshm() { may sshm \"$@\"; }"
	}
}

func idHookSnippet(shell string) string {
	switch shell {
	case "fish":
		return "functions --erase _may_id_hook 2>/dev/null\n" +
			"function --on-variable PWD _may_id_hook\n" +
			"  \\command may id status --apply --quiet\n" +
			"end"
	case "zsh":
		return "autoload -Uz add-zsh-hook\n" +
			"function _may_id_hook() { \\command may id status --apply --quiet }\n" +
			"chpwd_functions=(\"${(@)chpwd_functions:#_may_id_hook}\")\n" +
			"chpwd_functions+=(_may_id_hook)"
	default:
		return "if [[ ${PROMPT_COMMAND:-} != *'_may_id_hook'* ]]; then\n" +
			"  function _may_id_hook() { \\command may id status --apply --quiet; }\n" +
			"  PROMPT_COMMAND=\"${PROMPT_COMMAND:+${PROMPT_COMMAND}; }_may_id_hook\"\n" +
			"fi"
	}
}

func aiFixSnippet(shell string) string {
	switch shell {
	case "fish":
		return "function _may_ai_fix_postexec --on-event fish_postexec\n" +
			"  set -l last_status $status\n" +
			"  if test $last_status -ne 0\n" +
			"    \\command may ai fix --exit-code $last_status -- $argv[1]\n" +
			"  end\n" +
			"end"
	case "zsh":
		return "_may_ai_fix_last_cmd=\"\"\n" +
			"function _may_ai_fix_preexec() { _may_ai_fix_last_cmd=\"$1\"; }\n" +
			"function _may_ai_fix_precmd() {\n" +
			"  local ec=$?\n" +
			"  if [[ $ec -ne 0 ]] && [[ -n \"$_may_ai_fix_last_cmd\" ]]; then\n" +
			"    \\command may ai fix --exit-code \"$ec\" -- \"$_may_ai_fix_last_cmd\"\n" +
			"  fi\n" +
			"  _may_ai_fix_last_cmd=\"\"\n" +
			"}\n" +
			"autoload -Uz add-zsh-hook\n" +
			"add-zsh-hook preexec _may_ai_fix_preexec\n" +
			"add-zsh-hook precmd _may_ai_fix_precmd"
	default:
		return "_may_ai_fix_last_cmd=\"\"\n" +
			"function _may_ai_fix_debug() { _may_ai_fix_last_cmd=\"$BASH_COMMAND\"; }\n" +
			"function _may_ai_fix_prompt() {\n" +
			"  local ec=$?\n" +
			"  if [[ $ec -ne 0 ]] && [[ -n \"$_may_ai_fix_last_cmd\" ]]; then\n" +
			"    \\command may ai fix --exit-code \"$ec\" -- \"$_may_ai_fix_last_cmd\"\n" +
			"  fi\n" +
			"  _may_ai_fix_last_cmd=\"\"\n" +
			"}\n" +
			"trap '_may_ai_fix_debug' DEBUG\n" +
			"if [[ ${PROMPT_COMMAND:-} != *'_may_ai_fix_prompt'* ]]; then\n" +
			"  PROMPT_COMMAND=\"${PROMPT_COMMAND:+${PROMPT_COMMAND}; }_may_ai_fix_prompt\"\n" +
			"fi"
	}
}

func completionSnippet(shell string) string {
	switch shell {
	case "fish":
		return "\\command may shell completion fish | source"
	default:
		return fmt.Sprintf("eval \"$(\\command may shell completion %s)\"", shell)
	}
}
