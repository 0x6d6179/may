package root

import (
	"github.com/0x6d6179/may/internal/cmd/ai"
	"github.com/0x6d6179/may/internal/cmd/alias"
	"github.com/0x6d6179/may/internal/cmd/b64"
	"github.com/0x6d6179/may/internal/cmd/branch"
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
	"github.com/0x6d6179/may/internal/cmd/recent"
	"github.com/0x6d6179/may/internal/cmd/run"
	"github.com/0x6d6179/may/internal/cmd/secret"
	"github.com/0x6d6179/may/internal/cmd/shell"
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

	cmd.Version = version.Version

	// workspace & navigation
	cmd.AddCommand(ws.NewCmdWs(f))
	cmd.AddCommand(wt.NewCmdWt(f))
	cmd.AddCommand(j.NewCmdJ(f))
	cmd.AddCommand(branch.NewCmdBranch(f))
	cmd.AddCommand(recent.NewCmdRecent(f))
	cmd.AddCommand(open.NewCmdOpen(f))

	// ai
	cmd.AddCommand(ai.NewCmdAi(f))

	// git utilities
	cmd.AddCommand(stash.NewCmdStash(f))
	cmd.AddCommand(todo.NewCmdTodo(f))
	cmd.AddCommand(env.NewCmdEnv(f))

	// project tools
	cmd.AddCommand(run.NewCmdRun(f))
	cmd.AddCommand(port.NewCmdPort(f))
	cmd.AddCommand(db.NewCmdDb(f))

	// system & path
	cmd.AddCommand(pathcmd.NewCmdPath(f))
	cmd.AddCommand(ip.NewCmdIp(f))
	cmd.AddCommand(dotfiles.NewCmdDotfiles(f))
	cmd.AddCommand(weather.NewCmdWeather(f))

	// encode / decode
	cmd.AddCommand(b64.NewCmdB64(f))
	cmd.AddCommand(uuid.NewCmdUuid(f))
	cmd.AddCommand(hash.NewCmdHash(f))
	cmd.AddCommand(jwt.NewCmdJwt(f))
	cmd.AddCommand(secret.NewCmdSecret(f))

	// identity & meta
	cmd.AddCommand(id.NewCmdId(f))
	cmd.AddCommand(alias.NewCmdAlias(f))
	cmd.AddCommand(shell.NewCmdShell(f, cmd))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(initcmd.NewCmdInit(f))

	completionCmd := completion.NewCmdCompletion(f, cmd)
	completionCmd.Hidden = true
	cmd.AddCommand(completionCmd)

	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Short = "help about any command"
		}
	}

	return cmd
}
