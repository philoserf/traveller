# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
task                 # = task check: fmt:check, vet, lint, test, build (what CI runs)
task fmt             # auto-format; mutates files, not part of `task`/CI
task text            # regenerate reference/*.txt from reference/*.pdf
task brew            # install go, golangci-lint, go-task, poppler (once)
```

`Taskfile.yml` is the single source of truth for "passing" — `.github/workflows/ci.yml`
runs the same `task check`, nothing CI-only. Run `task check` before opening a PR.

Single test / package:

```sh
go test ./character/ -run TestQREBSPointScaleMatchesTheBook -v
go test ./character/                                   # one package
```

`golangci-lint fmt ./...` is the formatter (gofumpt/goimports/golines). Plain `gofmt`
is not enough — `fmt:check` will still fail on long lines and import grouping.

## Rules source: read it before implementing

This project transcribes the Traveller5 rules. **The rulebook text is the authority
for every mechanic — not the GitHub issue describing it.** Issue text has been wrong
outright more than once (misquoted Fame multipliers, misnamed table dimensions), so
read the cited page and quote it in both the doc comment and the PR body.

`reference/` holds `pdftotext` extracts of the three T5 core rulebooks. Book 1
(`Traveller5 Core Rules Book 1 Characters and Combat.txt`, ~19k lines) covers all of
character generation. **`reference/` is git-ignored** — it is reproduced locally from
user-supplied PDFs via `task text` (see README "Contributing: rulebook reference
material"). If it is missing, say so and ask; do not treat a rule as unverifiable
while the directory exists unread.

Two properties of those extracts matter when transcribing a table:

- **Adjacent page columns are interleaved into table rows.** Neighbouring body text
  and even a table's own header get spliced into data rows.
- **Read tables by character offset, not visually.** Extracting each token's
  start/end column recovers the real grid; visual reading has produced wrong
  transcriptions.
- **The book contradicts itself.** When two passages disagree, look for a printed
  value table or worked example that only one reading reproduces, and record which
  won and why in the doc comment.

## Architecture

Std-lib only. `go.mod` has no `require` block and `depguard` enforces it — a new
third-party dependency is a deliberate decision, not a drive-by.

Packages layer strictly downward (no cycles):

```
ehex, dice          primitives: extended-hex digits (0-9,A-Z less I/O); the dice Roller
  └─ world          UWP, trade codes, world generation
       └─ system    stars, orbits, satellites
            └─ sector   hex grid, subsectors
  └─ character      chargen: UPP, careers, skills, mustering out  (largest by far, ~26k lines)
  └─ starship       types and constants; generation not yet built
render              text output for character/sector/system/world
api                 HTTP handlers (see README for the endpoint table)
cmd/*               chargen, secgen, shipgen, sysgen, worldgen, server, client
```

`character` is where nearly all the complexity lives — see `character/CLAUDE.md` for
its file-naming convention.

### Generation is seeded and reproducible

Everything generative takes a `*dice.Roller` built from an explicit seed. The API
echoes the resolved seed back so a result can be reproduced; CLIs take `-seed`.

**The reproducibility contract is the dice stream, not just the output.** This has
concrete consequences:

- Implement a book procedure _literally_, not as a distributional equivalent. Book 1's
  "reroll if >3" is a reroll loop (`rollRestrictedD6`), not `Uniform(3)` — same
  distribution, different draw count. A roll the book makes conditionally ("...if
  required") must consume no die when not required.
- Every career exists twice: standalone `Generate<X>Character` and a segment inside
  `GenerateCareerChainCharacter`. `TestCareerChainSingleEntryMatchesLegacyGenerator`
  asserts a seed yields an identical character either way. **A new roll added to one
  path must sit at the same position relative to other draws in the other**, or that
  test fails with a whole-struct diff that is painful to read.
- Dice cost per skill draw is not fixed — cells like "One Trade" resolve themselves
  with further rolls. Assert the draw _shape_, not a total.
- Counting draws by wrapping `rand.Source` does not work: `rand.IntN`
  rejection-samples, so a source returning a constant deadlocks. Compare two same-seed
  rollers after the operation instead.

Any rules change shifts generated output. Measure the before/after distribution and
state it in the PR rather than letting review discover it. A fix that changes nothing
measurable is usually a signal the reading is wrong, not that the rule is minor.

## Conventions

- **Rule text lives beside the code.** Doc comments quote the governing passage with
  its page number, and record why an ambiguity resolved the way it did. This is the
  primary documentation — it is why completed work is not restated in `PLAN.md`.
- **Table values follow the book; Go identifiers don't have to.** For a string a
  table emits or compares against the reference text — skill names, rank titles —
  normalize the book's own column-width abbreviation to its own canonical spelled-out
  form when that full form is corroborated elsewhere in the text ("JOT" →
  "Jack of all Trades", confirmed against Scout's own table), but preserve the book's
  literal term verbatim, however odd or inconsistent with a sibling table, when it's a
  confirmed standing usage rather than a typo (Marine's "Coronel" stays split from
  Soldier's "Colonel" — the book prints it three times, consistently; a one-off OCR
  slip like "Pilor" does get corrected to "Pilot"). This test has never applied to Go
  identifiers themselves — no struct field, function, or constant needs to match a
  printed table header, only the values it produces.
- **Tests are white-box on purpose** (`testpackage` is disabled) so they can reach
  unexported tables and generation logic. Prefer assertions derived from generated
  data over pinned magic numbers; seed-pinned fixtures have needed re-deriving after
  nearly every rules change, and some have passed only by luck.
- **Verify a new test fails against the pre-fix code** before trusting it. Mutating
  the implementation and counting failing assertions is the established way to show a
  test has teeth.
- **Generalize on the second instance**, not the first — several shared helpers
  (`rollRestrictedD6`, `resolveCareerLoop`, `musterOutRow`) were extracted only once a
  second career genuinely needed them.
- **Prefer derived accessors to stored fields** when a value is a function of others
  (`Character.NobleTitle()`, `LandGrantIncome()`), so they cannot drift.

`PLAN.md` tracks remaining character-generation work, ordered with rationale, plus
standing per-PR checks and lessons. Starship generation is deliberately out of its
scope.

## Linting

`.golangci.yml` enables every linter, then disables only those that fight this repo's
deliberate design (transcribed tables are magic numbers by nature; career data are
package-level globals; white-box tests). Rationale is inline per linter.

Ones that bite most often in practice: `godot` (comments end in a period — put it
_outside_ a closing quote), `lll` (120 cols), `wsl_v5` (blank line before a block
after several statements), `funlen`/`gocognit`, `prealloc`, and `unparam` (an unused
parameter fails the build, so do not stub one "for symmetry").
