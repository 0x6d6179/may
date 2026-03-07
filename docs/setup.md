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
- **shell integration** — writes a block to your shell profile with the integrations you pick

---

## shell integration

run `may shell configure` at any time to reconfigure. it rewrites the managed block in your profile without touching anything else.

to apply immediately after first configure:

```sh
source ~/.zshrc   # or ~/.bashrc
```

### what the core wrapper does

the `may()` shell function intercepts stdout: if the output is a valid directory path, it calls `cd` instead of printing it. this is how `ws`, `wt`, and `j` change your working directory.

### available integrations

| integration | what it adds |
|---|---|
| `ws` | `ws <name>` — switch workspace |
| `wt` | `wt <name>` — switch worktree |
| `ai` | `ai <prompt>` — shorthand for `may ai` |
| `j` | `j <query>` — fuzzy jump; `k` — go back |
| `sshm` | `sshm` — shorthand for `may sshm` |
| `ai fix` | hooks into shell error handling to suggest fixes |
| `completion` | tab completion for all commands |

any command can also get a plain shell alias (`function branch() { may branch "$@"; }`) — pick them in `may shell configure`.

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
