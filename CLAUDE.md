# CLAUDE.md

`tt` is a terminal calendar (Go, Bubble Tea v2 + Lip Gloss v2) for time entries stored in [ttrack.top](https://ttrack.top). Module: `github.com/ricardoruwer/tt-cli`, Go 1.26.

## Commands

```sh
make test            # go test ./...
make build           # builds bin/tt (bin/ is gitignored)
go install ./cmd/tt  # install the local checkout
go test ./internal/ui -run TestPrintGrid -v   # prints a rendered grid
```

## Layout

- `cmd/tt/main.go` — loads config, builds the API client, runs the Bubble Tea program.
- `internal/config` — token from `~/.config/tt/config.toml` (honours `XDG_CONFIG_HOME`); `TT_TOKEN` overrides it. Warns when the file is readable by others. `BaseURL` is a hardcoded constant.
- `internal/api` — `Client.Month(ctx, year, month)`: `GET /api/time-tracking/:year/:month?separated=true` with a Bearer token; response is `{"data": [Entry]}`. 401/403 become a "token rejected" error; the body is capped at 4 MiB.
- `internal/model` — `Entry` and helpers (`Minutes`, `Lines`, `CodeReviewList`, `DeployList`, `FormatHM`, `IndexByDay`, `TotalMinutes`).
- `internal/calendar` — pure date math: Sunday-first week grid, `DaysInMonth`, `WorkingDays` (Mon–Fri, holidays ignored).
- `internal/ui` — `model.go` is state + `Update`/key handling; `view.go` is rendering.

## API payload

With `separated=true` each entry has `date` (`YYYY-MM-DD`), `time` (`HH:MM:SS`), `description` (newline-separated lines), `code_reviews` and `deploys` (comma-separated ticket keys, e.g. `"VIRT-3629, PRICE-5602"`). `code_reviews`/`deploys` can be `null`; they decode to `""`. Without the flag, code reviews and deploys are folded into `description`. The client always sends the flag, so the calendar never sees the folded form.

The API requires auth, so it cannot be explored anonymously.

## UI behaviour worth knowing

- Two modes: calendar grid and a per-day detail view (`m.detail`). The grid cell shows only the description lines; the detail view shows description, then "Code Reviews" and "Deploys" sections (a section is omitted when empty).
- Detail text is word-wrapped to the terminal width, so line count (and scroll clamping) depends on `m.width`; always go through `detailLines()`.
- Grid rendering must fill the terminal exactly (every line `m.width` wide, `m.height` lines total) — `TestGridFillsTerminalExactly` enforces it. Minimum terminal size is 70x(4 + 3 per week) and a "too small" message is shown below that.
- Fetches are keyed by year/month; `loadedMsg` for a month other than the current one is dropped.
- Entries are indexed by day of month via `IndexByDay`, which ignores entries outside the displayed month and malformed dates.

## Conventions

- Tests are plain `testing` with no extra deps; UI tests build a `Model` directly (`sampleModel` in `view_test.go`) and assert on `ansi.Strip(m.render())`.
- Keep the key table in `README.md` and `renderHelp()` in sync when changing keybindings.
- Never commit or print the user's token.
