# may

personal productivity toolkit for developers.

```sh
brew tap 0x6d6179/may && brew install may
```

---

## what it does

may is a single binary that replaces a collection of shell scripts and aliases with a consistent, discoverable interface.

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
  todo        find todo, fixme, and hack comments
  env         manage .env files

project tools
  run         run project scripts (package.json, Makefile, etc.)
  port        show or kill processes on a port
  db          connect to a database from .env urls

system & path
  dotfiles    manage dotfile symlinks with git-backed migration
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
  id          git identity management — switch profiles per directory
  sshm        ssh connection manager
  alias       manage shell function aliases
  commands    list and toggle available commands
  shell       configure shell integration
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

may uses a shell wrapper to enable directory-changing commands (`ws`, `wt`, `j`). run:

```sh
may shell configure
```

this writes a block to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) and lets you pick which integrations to enable.

---

## docs

- [setup guide](docs/setup.md)
- [dev guide](docs/dev.md)
