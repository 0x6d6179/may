# dev guide

## prerequisites

- go 1.23+
- git

## clone and build

```sh
git clone https://github.com/0x6d6179/may
cd may
make build
```

produces a `may` binary in the current directory.

## install locally (dev mode)

```sh
make install
```

installs to `$(go env GOPATH)/bin`. if you want the shell wrapper to use your local build instead of an installed release, run:

```sh
may shell configure --dev
```

this prepends the current directory to `$PATH` so `may` resolves to your local build. a warning comment is written into the shell block — remove it when you're done.

## run tests

```sh
make test
```

## project layout

```
cmd/may/          entrypoint
internal/
  cmd/            one package per command
    ai/
    branch/
    dotfiles/
    ws/
    wt/
    ...
  config/         config load/save
  factory/        shared dependency container (io, config)
  iostreams/      stdin/stdout/stderr abstraction
  ui/             bubbletea components (select, multiselect, input, confirm, form)
  version/        version string (set via ldflags)
```

## adding a command

1. create `internal/cmd/<name>/<name>.go` — export `NewCmd<Name>(f *factory.Factory) *cobra.Command`
2. register it in `internal/cmd/root/root.go` with `addCmd(cmd, "<group>", <name>.NewCmd<Name>(f))`
3. add to `toggleableCommands` in `internal/cmd/init/init.go`
4. add to `allCommands` in `internal/cmd/shell/configure.go`
5. add to `reservedNames` in `internal/cmd/alias/alias.go`

## conventions

- all user-facing strings lowercase
- output goes to `f.IO.ErrOut` (stderr); only path strings for `cd` go to `f.IO.Out` (stdout)
- do not use `fmt.Println` — use `fmt.Fprintf(f.IO.ErrOut, ...)`
- do not use viper, go-git, or libgit2
- interactive ui uses bubbletea via the `internal/ui` package — `RunSelect`, `RunMultiSelect`, `RunInput`, `RunConfirm`, `RunForm`
- check `f.IO.IsErrTerminal()` before running interactive flows

## releases

tag with a semver tag — the makefile picks it up via `git describe`:

```sh
git tag v0.x.0
git push origin v0.x.0
```

update `Formula/may.rb` with the new version and sha256 of the release tarball:

```sh
curl -fsSL https://github.com/0x6d6179/may/archive/refs/tags/v0.x.0.tar.gz | shasum -a 256
```

then add `sha256 "..."` to the formula.
