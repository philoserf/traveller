#!/usr/bin/env bash
# The whole build, the same on a Mac and in Workers Builds: wrangler.jsonc's
# build.command runs it before every deploy, `task build` runs it, and
# `task check` runs it with `check`.
#
#   build.sh         validate, then build the site into public/
#   build.sh check   validate, then build in memory; writes nothing
set -o errexit -o nounset -o pipefail

mode=${1:-build}
case $mode in
build | check) ;;
*)
  echo "usage: build.sh [check]" >&2
  exit 1
  ;;
esac

# Workers Builds clones shallow, which gives every page the head commit's date
# under enableGitInfo, so fetch full history first (and fail if that fails).
if [ "$(git rev-parse --is-shallow-repository)" != false ]; then
  git fetch --unshallow --quiet
fi
# Hugo keys git dates by the paths git log prints, which git quotes when they
# hold non-ASCII characters, so turn quoting off in the clone's config.
git config core.quotepath false

# The validate pass renders drafts and future-dated posts in memory, so one
# written ahead of time fails here rather than on the day it publishes.
hugo --buildFuture --buildDrafts --renderToMemory --gc --panicOnWarning

if [ "$mode" = check ]; then
  hugo --gc --minify --renderToMemory --panicOnWarning
else
  # rm -rf because Hugo never prunes public/.
  rm -rf public
  hugo --gc --minify --panicOnWarning
fi
