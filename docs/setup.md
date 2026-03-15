# setup

## install

### homebrew (recommended)

```sh
brew tap 0x6d6179/may && brew install may
```

the tap uses the formula at `Formula/may.rb` in the main repo. builds from source — requires go.

### install script

```sh
curl -fsSL https://raw.githubusercontent.com/0x6d6179/may/main/install.sh | sh
```

tries to download a pre-built binary first. falls back to `go install` if no binary is available for your platform. runs `may init` on completion.

### go install

```sh
go install github.com/0x6d6179/may/cmd/may@latest
```

installs to `$(go env GOPATH)/bin`. make sure that directory is in your `$PATH`.

---

## first run

```sh
may init
```

the setup wizard walks you through:

- **workspace roots** — directories where your projects live (e.g. `~/Code`, `~/Projects`)
- **git identity** — name, email, github username. supports multiple profiles mapped per directory via `may id`
- **ai** — provider (openrouter, cerebras, or custom openai-compatible endpoint), api key, model
- **shell integration** — writes a managed block to your shell profile

---

## shell integration

run `may shell configure` at any time to reconfigure. it rewrites the managed block in your profile without touching anything else.

to apply immediately after first configure:

```sh
source ~/.zshrc   # or ~/.bashrc
```

### always-on: the core wrapper

two things are always written regardless of what you select:

**`may()` function** — wraps the binary. if stdout is a valid directory path, calls `cd` instead of printing it. this is what allows `ws`, `wt`, and `j` to change your working directory.

**`_may_id_hook`** — fires every time you `cd`. runs `may id status --apply --quiet` to auto-switch your git identity based on the directory.

### shell hooks

two optional integrations that modify shell behavior beyond simple aliases:

| hook | what it does |
|---|---|
| `ai fix` | hooks into `PROMPT_COMMAND` / `precmd` to detect non-zero exit codes and run `may ai fix` |
| `completion` | runs `eval "$(may shell completion zsh)"` (or bash/fish) for tab completion |

### command aliases

everything else — including `ws`, `wt`, `ai`, `j`, `sshm`, and any other command you pick — is a shell function alias:

```sh
function ws() { may ws "$@"; }
function branch() { may branch "$@"; }
```

they are all identical in nature. the only difference is which ones are checked by default. pick any combination in `may shell configure`.

---

## command management

```sh
may commands list        # show all commands with enabled/disabled status
may commands configure   # interactive enable/disable
```

`may commands configure` opens a multi-select. uncheck a command to disable it — it will be hidden and return an error if invoked. for commands that have a shell alias (`ws`, `wt`, `ai`, `j`, `sshm`), disabling a command will also remove its alias from your shell profile automatically.

`commands` is also available as `cmd` or `cmds`.

### alias vs commands

- `may alias` — manages **user-defined** shell aliases. you pick the name and which `may` command it maps to. these are written to your shell block.
- `may commands` — manages which **built-in** may commands are enabled, and reflects their status.

---

## ai setup

may sends requests to any openai-compatible api. set your key and model in `~/.config/may/config.yaml` or via `may init`:

```yaml
ai:
  provider: openrouter
  base_url: https://openrouter.ai/api/v1
  api_key: sk-or-...
  model: inception/mercury-2
```

or set `MAY_AI_API_KEY` as an environment variable to override the stored key.

---

## configuration

config lives at `~/.config/may/config.yaml`. dotfiles state lives at `~/.config/may/dotfiles.yaml`.

to edit directly:

```sh
$EDITOR ~/.config/may/config.yaml
```

---

## updating

```sh
may update
```

or via homebrew:

```sh
brew upgrade may
```
