# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

The site at `traveller.philoserf.com`: one Hugo home page listing the philoserf Traveller tools. It was scaffolded from `philoserf/sandbox` and does not track it; diverge freely.

## Commands

All commands use Task (`Taskfile.yml`). `task check` is the gate: the two Hugo passes of `wrangler.jsonc`'s `build.command`, in memory, with `--panicOnWarning`. There is no other test.

## Architecture

- **Config**: `hugo.yaml`. `disableKinds` turns off RSS, sections, and taxonomies; the home page outputs HTML, `llms.txt`, and `site.webmanifest`.
- **Design**: the 1977 Little Black Books. The home page is the Book 1 cover (black, white italic Optima, a full-width red rule, red TRAVELLER); every other page is an interior page (white paper, Univers-style sans, justified text, bold underlined caps for `h2`, a centred folio). `baseof.html` gives the home page its own body and wraps every other page in the interior frame. An interior page's folio is its place among regular pages by weight; a template can set its own with a `folio` block, as `404.html` does.
- **Home page**: `layouts/index.html` renders `params.home.tools`, linking each `repo` to `github.com/philoserf/<repo>`. `layouts/index.llms.txt` renders the same list. Add a tool by adding an entry there.
- **Pages**: a markdown file under `content/` renders with `layouts/_default/single.html`. Add a `menus.main` entry in `hugo.yaml` to put it in the nav.
- **Theme colors**: `params.themeColor` must match `--paper` in `assets/css/style.css`; the home page's theme-color meta in `baseof.html` must match `--cover-black`.

## Deployment

A static-assets-only Cloudflare Worker at the custom domain `traveller.philoserf.com`; `README.md` lists the settings that live only in Cloudflare. `task deploy` builds and deploys by hand. Workers Builds deploys every push to `main` once the Builds connection exists.
