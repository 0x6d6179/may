# may

personal productivity toolkit for developers.

```sh
brew tap 0x6d6179/may && brew install may
```

---

## what it does

may is a single binary that replaces a pile of shell scripts and ad-hoc aliases with a consistent, discoverable interface.

```
workspace & navigation
  ws          switch workspace and cd into it
  wt          git worktree manager
  j           smart directory jump with fuzzy matching
  branch      interactive branch switcher
  recent      show recently visited projects
  open        open repository in browser

ai
  ai          ai assistant (openrouter / cerebras / any openai-compat api)

git utilities
  stash       interactive stash manager
  todo        find todo, fixme, and hack comments in the codebase
  env         manage .env files

project tools
  run         run project scripts (package.json, Makefile, etc.)
  port        show or kill processes on a port
  db          connect to a database from .env urls

system & path
  dotfiles    git-backed dotfile migration, sync, and restore
  ip          show local and public ip addresses
  path        inspect and debug $PATH
  weather     weather forecast in your terminal

encode / decode
  b64         base64 encode/decode
  uuid        generate uuids
  hash        hash strings or files (sha256, md5, etc.)
  jwt         decode and inspect jwt tokens
  secret      encrypt or decrypt secrets for safe sharing
  qr          generate qr codes

identity & meta
  id          per-directory git identity — auto-switch name/email by path
  sshm        ssh connection manager
  alias       user-defined shell aliases (map any name to any may command)
  commands    enable or disable commands; show status  (also: cmd, cmds)
  shell       manage shell integration — configure aliases and hooks
```

---

## install

### homebrew

```sh
brew tap 0x6d6179/may && brew install may
```

### script

```sh
curl -fsSL https://raw.githubusercontent.com/0x6d6179/may/main/install.sh | sh
```

### go install

```sh
go install github.com/0x6d6179/may/cmd/may@latest
```

after installing, run:

```sh
may init
```

---

## shell integration

may needs a shell wrapper to intercept stdout — this is how `ws`, `wt`, and `j` change your directory instead of just printing a path.

run the interactive setup:

```sh
may shell configure
```

it writes a managed block to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) with two always-on pieces:

- `may()` — wraps the binary; if stdout is a directory path, calls `cd` instead of printing
- `_may_id_hook` — fires on `cd` to apply the right git identity for the directory

everything else is opt-in shell function aliases — `function ws() { may ws "$@"; }` — you pick which commands get them.

---

## docs

- [setup guide](docs/setup.md)
- [dev guide](docs/dev.md)
