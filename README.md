# traveller

An API-first implementation of the Traveller5 (T5) tabletop RPG rules.
Types first, then tools: domain types for worlds, characters, and
starships live in the `world`, `character`, and `starship` packages,
built on a shared `ehex` (extended-hex) primitive.

## Legal / trademark notice

**Traveller** is a registered trademark of Far Future Enterprises.
Portions of this work are derived from the Traveller game in general
and the Traveller5 Core Rules in particular, used under Far Future
Enterprises' Fair Use Policy. This is an unofficial, non-commercial fan
project and is not affiliated with, endorsed by, or sponsored by Far
Future Enterprises. Copyright of Traveller game content remains with
Far Future Enterprises. See <https://farfuture.net> for the current,
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
3. Run `task text` (or just `task`) to extract matching `reference/*.txt`
   files from those PDFs. The task is a no-op / skips when the PDFs
   haven't changed, so it's safe to re-run.

`reference/` stays untracked — only your local copies exist, and they
aren't committed.
