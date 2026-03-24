package root

import (
	"fmt"

	"github.com/0x6d6179/may/internal/cmd/ai"
	"github.com/0x6d6179/may/internal/cmd/alias"
	"github.com/0x6d6179/may/internal/cmd/b64"
	"github.com/0x6d6179/may/internal/cmd/branch"
	"github.com/0x6d6179/may/internal/cmd/commands"
	"github.com/0x6d6179/may/internal/cmd/completion"
	"github.com/0x6d6179/may/internal/cmd/db"
	"github.com/0x6d6179/may/internal/cmd/dotfiles"
	"github.com/0x6d6179/may/internal/cmd/env"
	"github.com/0x6d6179/may/internal/cmd/hash"
	"github.com/0x6d6179/may/internal/cmd/id"
	initcmd "github.com/0x6d6179/may/internal/cmd/init"
	"github.com/0x6d6179/may/internal/cmd/ip"
	"github.com/0x6d6179/may/internal/cmd/j"
	"github.com/0x6d6179/may/internal/cmd/jwt"
	"github.com/0x6d6179/may/internal/cmd/open"
	pathcmd "github.com/0x6d6179/may/internal/cmd/path"
	"github.com/0x6d6179/may/internal/cmd/port"
	"github.com/0x6d6179/may/internal/cmd/qr"
	"github.com/0x6d6179/may/internal/cmd/recent"
	"github.com/0x6d6179/may/internal/cmd/run"
	"github.com/0x6d6179/may/internal/cmd/secret"
	"github.com/0x6d6179/may/internal/cmd/shell"
	"github.com/0x6d6179/may/internal/cmd/sshm"
	"github.com/0x6d6179/may/internal/cmd/stash"
	"github.com/0x6d6179/may/internal/cmd/todo"
	"github.com/0x6d6179/may/internal/cmd/update"
	"github.com/0x6d6179/may/internal/cmd/uuid"
	"github.com/0x6d6179/may/internal/cmd/weather"
	"github.com/0x6d6179/may/internal/cmd/ws"
	"github.com/0x6d6179/may/internal/cmd/wt"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/version"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "may",
		Short:         "personal productivity toolkit",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Version = version.Full()
	cmd.SetVersionTemplate("may {{.Version}}\n")

	cmd.AddGroup(
		&cobra.Group{ID: "nav", Title: "workspace & navigation"},
		&cobra.Group{ID: "ai", Title: "ai"},
		&cobra.Group{ID: "git", Title: "git utilities"},
		&cobra.Group{ID: "project", Title: "project tools"},
		&cobra.Group{ID: "system", Title: "system & path"},
		&cobra.Group{ID: "encode", Title: "encode / decode"},
		&cobra.Group{ID: "meta", Title: "identity & meta"},
	)

	// workspace & navigation
	addCmd(cmd, "nav", ws.NewCmdWs(f))
	addCmd(cmd, "nav", wt.NewCmdWt(f))
	addCmd(cmd, "nav", j.NewCmdJ(f))
	addCmd(cmd, "nav", branch.NewCmdBranch(f))
	addCmd(cmd, "nav", recent.NewCmdRecent(f))
	addCmd(cmd, "nav", open.NewCmdOpen(f))

	// ai
	addCmd(cmd, "ai", ai.NewCmdAi(f))

	// git utilities
	addCmd(cmd, "git", stash.NewCmdStash(f))
	addCmd(cmd, "git", todo.NewCmdTodo(f))
	addCmd(cmd, "git", env.NewCmdEnv(f))

	// project tools
	addCmd(cmd, "project", run.NewCmdRun(f))
	addCmd(cmd, "project", port.NewCmdPort(f))
	addCmd(cmd, "project", db.NewCmdDb(f))

	// system & path
	addCmd(cmd, "system", pathcmd.NewCmdPath(f))
	addCmd(cmd, "system", ip.NewCmdIp(f))
	addCmd(cmd, "system", dotfiles.NewCmdDotfiles(f))
	addCmd(cmd, "system", weather.NewCmdWeather(f))

	// encode / decode
	addCmd(cmd, "encode", b64.NewCmdB64(f))
	addCmd(cmd, "encode", uuid.NewCmdUuid(f))
	addCmd(cmd, "encode", hash.NewCmdHash(f))
	addCmd(cmd, "encode", jwt.NewCmdJwt(f))
	addCmd(cmd, "encode", secret.NewCmdSecret(f))
	addCmd(cmd, "encode", qr.NewCmdQr(f))

	// identity & meta
	addCmd(cmd, "meta", id.NewCmdId(f))
	addCmd(cmd, "meta", sshm.NewCmdSshm(f))
	addCmd(cmd, "meta", alias.NewCmdAlias(f))
	addCmd(cmd, "meta", commands.NewCmdCommands(f))
	addCmd(cmd, "meta", shell.NewCmdShell(f, cmd))
	addCmd(cmd, "meta", update.NewCmdUpdate(f))
	addCmd(cmd, "meta", initcmd.NewCmdInit(f))

	completionCmd := completion.NewCmdCompletion(f, cmd)
	completionCmd.Hidden = true
	cmd.AddCommand(completionCmd)

	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Short = "help about any command"
		}
	}

	applyDisabledCommands(cmd, f)

	return cmd
}

func addCmd(parent *cobra.Command, groupID string, child *cobra.Command) {
	child.GroupID = groupID
	parent.AddCommand(child)
}

var neverDisable = map[string]bool{
	"commands": true, "shell": true, "update": true, "init": true,
	"help": true, "completion": true,
}

func applyDisabledCommands(root *cobra.Command, f *factory.Factory) {
	cfg, err := f.Config()
	if err != nil || len(cfg.DisabledCommands) == 0 {
		return
	}
	disabled := make(map[string]bool, len(cfg.DisabledCommands))
	for _, n := range cfg.DisabledCommands {
		if !neverDisable[n] {
			disabled[n] = true
		}
	}
	for _, c := range root.Commands() {
		if !disabled[c.Name()] {
			continue
		}
		c.Hidden = true
		name := c.Name()
		c.RunE = func(_ *cobra.Command, _ []string) error {
			return fmt.Errorf("%s is disabled — run: may commands configure", name)
		}
	}
}
