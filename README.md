# traveller

[![CI](https://github.com/philoserf/traveller/actions/workflows/ci.yml/badge.svg)](https://github.com/philoserf/traveller/actions/workflows/ci.yml)

An API-first implementation of the Traveller5 (T5) tabletop RPG rules.
Types first, then tools: domain types for worlds, characters, and
starships live in the `world`, `character`, and `starship` packages,
built on a shared `ehex` (extended-hex) primitive.

## API

`go run ./cmd/server` starts the HTTP API on `:8080`. All endpoints are
read-only `GET`s returning JSON; an omitted `seed` resolves to a
time-derived one, which the response always echoes back so the result
can be reproduced later. Handler behavior is also documented via Go doc
comments in `api/*.go` (`go doc ./api`).

| Endpoint              | Query params                                                                                                               | Response          |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `GET /healthz`        | —                                                                                                                          | `healthzResponse` |
| `GET /worlds/random`  | `seed`                                                                                                                     | `WorldResponse`   |
| `GET /systems/random` | `seed`                                                                                                                     | `SystemResponse`  |
| `GET /sectors/random` | `seed`, `name` (default "Unnamed"), `density` (default "Standard" — see `sector.Density`), `subsector` (single letter A-P) | `SectorResponse`  |

A bad `seed`/`density`/`subsector` responds `400` with
`{"error": "..."}` (`errorResponse`). `cmd/client` is a CLI that talks to
this same API — see its `-h` output for each subcommand's flags.

## Development

```sh
task brew   # install go, golangci-lint, go-task, poppler (once)
task        # fmt-check, vet, lint, test, build — same checks CI runs
task fmt    # auto-format (mutates files; not run by `task`/CI)
```

`Taskfile.yml` is the single source of truth for what "passing" means —
`.github/workflows/ci.yml` runs the same `task check` a contributor runs
locally, nothing CI-only.

## License

This repository's original code is MIT licensed — see [LICENSE](LICENSE).
That covers the Go source only; it grants no rights to the Traveller
trademark or T5 game content, which remain Far Future Enterprises'
under the terms below.

## Legal / trademark notice

**Traveller** is a registered trademark of Far Future Enterprises.
Portions of this work are derived from the Traveller game in general
and the Traveller5 Core Rules in particular, used under Far Future
Enterprises' Fair Use Policy. This is an unofficial, non-commercial fan
project and is not affiliated with, endorsed by, or sponsored by Far
Future Enterprises. Copyright of Traveller game content remains with
Far Future Enterprises. See [https://farfuture.net](http://archive.today/KGpR8) for the current,
authoritative text of the Fair Use Policy.

This repository does not redistribute the T5 rulebook PDFs or any
extracted rules text — `reference/` is local reference material only
and is git-ignored.

## Contributing: rulebook reference material

Code and rules discussions assume you have the T5 core rulebooks on
hand, but this repo can't ship them for you. To reproduce the local
reference material:

1. Buy/obtain your own copy of the three T5 core rulebook PDFs from
   Far Future Enterprises and place them, with these exact filenames,
   in `reference/`:
   - `Traveller5 Core Rules Book 1 Characters and Combat.pdf`
   - `Traveller5 Core Rules Book 2 Starships.pdf`
   - `Traveller5 Core Rules Book 3 Worlds and Adventures.pdf`
2. Run `task brew` once to install `pdftotext` (via the `poppler`
   Homebrew formula listed in `Brewfile`).
3. Run `task text` to extract matching `reference/*.txt` files from
   those PDFs. The task is a no-op / skips when the PDFs haven't
   changed, so it's safe to re-run.

`reference/` stays untracked — only your local copies exist, and they
aren't committed.
