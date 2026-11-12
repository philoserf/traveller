# traveller

The site at [traveller.philoserf.com](https://traveller.philoserf.com/): a home page for the philoserf Traveller tools ([t5chargen](https://github.com/philoserf/t5chargen), [ctchargen](https://github.com/philoserf/ctchargen), [ctworldgen](https://github.com/philoserf/ctworldgen), [cschargen](https://github.com/philoserf/cschargen)).

A Hugo site styled after the 1977 Little Black Books: the home page is the Book 1 cover, and any other page is an interior page. Scaffolded from [philoserf/sandbox](https://github.com/philoserf/sandbox).

## Usage

`task setup` installs the toolchain from `Brewfile`. Then `task dev` serves locally, `task check` runs the strict build in memory, and `task deploy` builds and deploys by hand.

The tool list is `params.home.tools` in `hugo.yaml`. Each entry's `repo` is a `philoserf` GitHub repo name.

## Deployment

A Cloudflare Worker serving `public/` as static assets, with no script, at the custom domain `traveller.philoserf.com`. `wrangler.jsonc` holds the Worker, its build, and its domain. Its `build.command` is the whole build, the same on a Mac (`task deploy`) and in Workers Builds.

### Cloudflare settings

These live only in Cloudflare, so restore them by hand if the Worker is ever recreated:

- **Builds connection** (Workers & Pages → `traveller` → Settings → Builds): GitHub `philoserf/traveller`, root directory `/`, build command empty, since `build.command` is the build. The Cloudflare GitHub app needs access to this repo.
- **Production branch** `main`, deploy command `bunx wrangler deploy`.
- **Non-production branches** on, deploy command `bunx wrangler versions upload`.
- **Build variable** `HUGO_VERSION=latest` on both.

On a fresh account, run `task deploy` once to create the Worker and attach the domain, then connect Builds. The domain attaches only when no other DNS record holds the hostname.

## License

- **Code / config**: [MIT](./LICENSE).
- **Content**: [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/). See [CONTENT-LICENSE.md](./CONTENT-LICENSE.md).
