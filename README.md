# tt

A terminal calendar for your [ttrack.top](https://ttrack.top) time entries. It shows the current month with the hours and the first lines of the description for each day, and opens the full description of a day on demand.

## Requirements

- [Go](https://go.dev/dl/) 1.26 or newer
- A ttrack.top API token

## Install

```sh
go install github.com/ricardoruwer/tt-cli/cmd/tt@latest
```

The binary is placed in `~/go/bin` (or `$GOBIN`, if set), and that directory **must be in your `PATH`** for the `tt` command to be found. Go does not add it for you. For zsh:

```sh
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
```

Open a new terminal (or run `source ~/.zshrc`) and check with `tt`. `go install` prints nothing when it succeeds, apart from download messages.

If `tt` is not found, find where Go put it with `go env GOBIN GOPATH`: the binary goes in `GOBIN` when it is set, otherwise in `GOPATH/bin`. Make sure that directory is in your `PATH`.

**asdf users:** the asdf Go plugin sets `GOBIN` to a directory inside the installed Go version, which is not in your `PATH`. Run `asdf reshim golang` after installing so asdf creates the `tt` shim, and run `go install` again after switching Go versions.

To update, run the same `go install` command again. To uninstall, delete `~/go/bin/tt`.

## Configure

Create `~/.config/tt/config.toml` (or `$XDG_CONFIG_HOME/tt/config.toml`) with your token and restrict it to your user:

```sh
mkdir -p ~/.config/tt
cat > ~/.config/tt/config.toml <<'EOF'
token = "<your token>"
EOF
chmod 600 ~/.config/tt/config.toml
```

The `TT_TOKEN` environment variable, when set, overrides the token from the file.

## Usage

Run `tt`.

| Key | Action |
| --- | --- |
| `h` `j` `k` `l` / arrows | Move the selection |
| `enter` | Open the details of the selected day |
| `t` | Jump to today |
| `r` | Reload from the server |
| `q` | Quit (or close the details) |

In the details view: `j`/`k` scroll, `space` pages down, `g`/`G` jump to top/bottom, `esc` goes back.

## Development

```sh
make test   # run the unit tests
make build  # build ./bin/tt
go install ./cmd/tt  # install the local checkout into ~/go/bin
```
