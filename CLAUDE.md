# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

The site at `traveller.philoserf.com`, "A Referee’s Notebook": dated posts on 1977 Classic Traveller, the Little Black Books before the Official Traveller Universe, with some commentary on what came after. Classic Traveller only: no other edition or setting (no Clement Sector), since the site carries only Mongoose Publishing’s notice. It was scaffolded from `philoserf/sandbox` and does not track it; diverge freely.

## Commands

All commands use Task (`Taskfile.yml`). `task check` is the gate: the two Hugo passes of `wrangler.jsonc`'s `build.command`, in memory, with `--panicOnWarning`. There is no other test.

## Architecture

- **Config**: `hugo.yaml`. Taxonomies are off. The home page outputs HTML, RSS (posts only, `layouts/index.rss.xml`), `llms.txt`, and `site.webmanifest`; sections output HTML only. `timeZone` makes a date-only post go live at local midnight; it appears at the next build after that, which the Monday cron guarantees within a week.
- **Design**: the 1977 Little Black Books. The home page is the Book 1 cover (black, white italic Optima, a full-width red rule, red TRAVELLER); every other page is an interior page (white paper, Univers-style sans, justified text, bold underlined caps for `h2`, a centred folio). `baseof.html` gives the home page its own body and wraps every other page in the interior frame. A post's folio is its place among the posts, oldest first, so it never changes; another page takes front matter `folio` or shows none. A template can override the `folio` block, as `404.html` and `list.html` do.
- **Home page**: `layouts/index.html` is the cover. Its two white lines are `params.home.kicker` and the site title, its tagline `params.home.tagline`, and the red line at its foot (the opening of the Free Trader Beowulf mayday, where Book 1 prints its publisher) `params.home.signal`; the ten newest posts sit below the fold. `static/og-image.png` is a 1200×630 rendering of the cover lines, so redraw it when they change.
- **Posts**: `task new:post -- <slug>` scaffolds `content/posts/<slug>.md` from `archetypes/posts.md` (`date` only, no `lastmod`; `enableGitInfo` backfills it). `layouts/_default/list.html` renders the archive like the LBB contents page, with dotted leaders.
- **Pages**: a markdown file under `content/` renders with `layouts/_default/single.html`. Add a `menus.main` entry in `hugo.yaml` to put it in the nav.
- **Disclaimer**: `DISCLAIMER.md` at the root is the one copy of the full Mongoose Publishing fan notice, worded as in ctchargen's README. `content/disclaimer.md` renders it through the `disclaimer` shortcode; the cover footer and `README.md` carry the short form.
- **Theme colors**: `params.themeColor` must match `--paper` in `assets/css/style.css`; the home page's theme-color meta in `baseof.html` must match `--cover-black`.

## Deployment

A Cloudflare Worker at the custom domain `traveller.philoserf.com`, serving `public/` as static assets; `README.md` lists the settings that live only in Cloudflare. `task deploy` builds and deploys by hand. Workers Builds deploys every push to `main` once the Builds connection exists. A Monday cron trigger runs `worker.js`, the Worker's only code, which POSTs the Worker's deploy hook (the secret `DEPLOY_HOOK`) so future-dated posts go live; its `fetch` handler only hands requests back to asset serving.
