---
title: Colophon
description: "How this notebook is designed, set and built."
lastmod: 2026-11-14
---

## Design

The site copies the Little Black Books of 1977. The home page is the cover of Book 1: black, with white italic lines above a red rule and TRAVELLER in red. Every other page is an interior page: white paper, a sans-serif text face, justified paragraphs, bold underlined capitals for headings and a centred folio at the foot.

## Typefaces

The cover asks for Optima and the interior for Univers, the faces of the original books. Where a device lacks them, the cover falls back to Candara, URW Classico, Linux Biolinum or Segoe UI, and the interior to Helvetica Neue, Helvetica or Arial. No webfonts are served: what you see is set in fonts already installed on your device, so there are no font licences to list.

## Tools

The pages are built with [Hugo](https://gohugo.io/), with [Task](https://taskfile.dev/) running the build and checks and [Bun](https://bun.com/)’s formatter tidying the source. A [Cloudflare Worker](https://workers.cloudflare.com/) serves the built site. The source is on [GitHub](https://github.com/philoserf/traveller).

## Credits

All the art on the site, the cover image and the icons, is original to it.

## Notices

The [disclaimer]({{< relref "disclaimer" >}}) carries Mongoose Publishing’s notice for Traveller and the Creative Commons terms (CC BY-NC-SA 4.0) for the site’s own writing.
