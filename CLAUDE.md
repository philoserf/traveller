# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

The site at `traveller.philoserf.com`: one Hugo home page listing the philoserf Traveller tools. Layouts and `assets/css/style.css` were copied from `philoserf/sandbox` (itself a staging copy of `philoserf/site`) and trimmed. This repo does not track either of them; diverge freely.

## Commands

All commands use Task (`Taskfile.yml`). `task check` is the gate: the two Hugo passes of `wrangler.jsonc`'s `build.command`, in memory, with `--panicOnWarning`. There is no other test.

## Architecture

- **Config**: `hugo.yaml`. `disableKinds` turns off RSS, sections, and taxonomies; the home page outputs HTML, `llms.txt`, and `site.webmanifest`.
- **Home page**: `layouts/index.html` renders `params.home.tools`, linking each `repo` to `github.com/philoserf/<repo>`. `layouts/index.llms.txt` renders the same list. Add a tool by adding an entry there.
- **Pages**: a markdown file under `content/` renders with `layouts/_default/single.html`. Add a `menus.main` entry in `hugo.yaml` to put it in the nav.
- **Theme colors**: `params.themeColor` must match `--background-color` in `assets/css/style.css`.
- `style.css` still holds rules for home-page sections this site dropped (featured, recent, motto, archive); they are unused, not load-bearing.

## Deployment

A static-assets-only Cloudflare Worker at the custom domain `traveller.philoserf.com`; `README.md` lists the settings that live only in Cloudflare. `task deploy` builds and deploys by hand. Workers Builds deploys every push to `main` once the Builds connection exists.
