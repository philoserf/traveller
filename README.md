# traveller

The site at [traveller.philoserf.com](https://traveller.philoserf.com/): _A Referee’s Notebook_, campaigns and commentary for the 1977 Classic Traveller rules.

A Hugo site styled after the 1977 Little Black Books: the home page is the Book 1 cover, and any other page is an interior page. Scaffolded from [philoserf/sandbox](https://github.com/philoserf/sandbox).

## Usage

`task setup` installs the toolchain from `Brewfile`. Then `task dev` serves locally, `task check` runs the strict build in memory, and `task deploy` builds and deploys by hand.

`task new:post -- <slug>` starts a post in `content/posts/`. Posts start as drafts; remove `draft: true` to publish. Drafts double as the list of articles in progress: `hugo list drafts` shows them, and `task dev` renders them locally.

## Deployment

A Cloudflare Worker serving `public/` as static assets, with Hugo's `404.html` as the not-found page, at the custom domain `traveller.philoserf.com`. `wrangler.jsonc` holds the Worker, its build, and its domain. Its `build.command` is the whole build, the same on a Mac (`task deploy`) and in Workers Builds.

A cron trigger in `wrangler.jsonc` runs `worker.js` every Monday at 09:00 UTC (05:00 Eastern in summer, 04:00 in winter). It POSTs the Worker's own deploy hook, so Workers Builds rebuilds the site and a future-dated post goes live within a week of its date. That script is the Worker's only code: requests are served from `public/`, and one matching no file is handed straight back to asset serving for the 404 page.

### Cloudflare settings

These live only in Cloudflare, so restore them by hand if the Worker is ever recreated:

- **Builds connection** (Workers & Pages → `traveller` → Settings → Builds): GitHub `philoserf/traveller`, root directory `/`, build command empty, since `build.command` is the build. The Cloudflare GitHub app needs access to this repo.
- **Production branch** `main`, deploy command `bunx wrangler deploy`.
- **Non-production branches** on, deploy command `bunx wrangler versions upload`.
- **Build variable** `HUGO_VERSION=latest` on both.
- **Deploy hook** "Monday rebuild" on `main`. Its URL is the Worker secret `DEPLOY_HOOK`, which `worker.js` POSTs on the cron. To replace it, create a hook under Settings → Builds → Deploy Hooks, pipe its URL to `bunx wrangler secret put DEPLOY_HOOK`, and delete the old hook.

On a fresh account, run `task deploy` once to create the Worker and attach the domain, then connect Builds and set `DEPLOY_HOOK`. Until the secret exists, the cron fails and nothing else does. The domain attaches only when no other DNS record holds the hostname.

## Disclaimer

An unofficial fan site for GDW's 1977 Classic Traveller. Traveller is a registered trademark of Mongoose Publishing; this site is not affiliated with or endorsed by them. See [DISCLAIMER.md](./DISCLAIMER.md) for the full notice.

## License

- **Code / config**: [MIT](./LICENSE).
- **Content**: [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/). See [CONTENT-LICENSE.md](./CONTENT-LICENSE.md).
