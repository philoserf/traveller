# Character generation plan

Remaining work on **character generation only**. Starship generation
(#6) is out of scope: re-plan it on its own terms when chargen closes,
scoped into a sequence of PRs rather than one issue.

Every entry is checked against
`reference/Traveller5 Core Rules Book 1 Characters and Combat.txt`, with
the page text quoted where it decides scope.

**Working rule:** read the cited rule text first, implement against it,
and quote it in the PR. Do not implement from the issue text alone —
several issues paraphrase the rules inaccurately, and three have been
wrong outright.

Completed work is not restated here: git history, PR descriptions, and
the doc comments at each implementation site carry it, and the comments
quote the governing rule directly.

---

## Rulings

Rules questions still relevant to what's left below. Each is recorded
at its implementation site with the governing quote; repeated here only
for context.

**#93 Noble rank — ladder-tracked, with Soc-derivation as the fallback.**
p.65 elevates "to the next higher Noble rank and its associated increase
in Social Standing (if any)" — the "(if any)" only means anything if
rank leads and Soc follows, since Baronet/Baron, Viscount/Count and the
two Dukes each share a Soc.

**#93 Land Grants — a Mustering Out Soc increase awards one.** p.85:
"Each increase in Soc during CharGen awards a Land Grant."

## Remaining work

- **#165** — step C itself attending more than one institution in
  sequence (e.g. ED5 raising Edu to 5, then continuing straight into
  Trade School), distinct from Later Education's own mid-career
  mechanism.
- **#164** — wire Later Education into Entertainer/Citizen/Noble, which
  hand-roll their own career loops instead of sharing
  `resolveCareerLoop`.
- **#163** — the Tra path (Apprenticeship, Mentor, Training Course).
  Needs non-Human generation, which doesn't exist anywhere in this
  codebase; probably should not be built until it does.
- **#96** — Land Grant scope deferrals: Preferred World, geodesic hex
  maps, Moot proxies and voting, and grant improvement. Independent of
  each other and of the items above. Preferred World and hex placement
  both want a world-selection concept that does not exist yet.

## Standing checks for every PR

- Read and quote the governing rule text; the issue's paraphrase is not the source.
- Verify each new test fails against the pre-fix code before trusting it.
- When a fix changes generated output, measure the before/after distribution
  and state it in the PR rather than letting review discover it.
- Prefer assertions derived from generated data over pinned magic numbers —
  pinned fixtures have needed re-deriving after nearly every rules change.
- `task check` clean before opening.

## Lessons worth not relearning

- **Measure before believing a rules fix did anything.** A summed-per-career
  award can come out inert even when the code changed — check the rule's
  actual unit (e.g. "per instance" vs. "per career") before trusting a fix.
- **Measure compounding halves together, not stage by stage.** Two staged
  changes that interact (one raises a value, the other pushes past a
  threshold using that value) can each look minor measured alone and be
  dramatic measured together. A staged rollout still needs a combined
  measurement before calling it done.
- **A mechanic can be correct, tested, and still never run.** Unit tests on
  an unreachable code path prove nothing about generated output. A test
  that walks the whole path and asserts the thing happens at all is the
  only thing that catches this.
- **Check whether a rule forbids the thing before building it.** A
  test-coverage gap isn't automatically a bug — confirm the rules actually
  allow the case before writing a generator for it.
- **When a table matters, read the PDF's word coordinates, not the text
  extract.** `pdftotext -bbox-layout` gives every word an x/y box, which
  recovers the real grid outright instead of inferring it from a linear
  text extract that can splice adjacent columns into one row.
- **The book contradicts itself; printed tables settle it.** When two
  passages disagree, a printed value table or worked example usually
  reproduces only one reading — that one wins.
- **Check whether a mechanic can actually fire.** Implemented and
  unit-tested is not the same as reachable; sweep generated output to
  confirm a mechanic actually triggers before considering it done.
- **A fixture that passes can still be wrong.** Pinned magic-number
  fixtures have repeatedly passed by luck — two errors cancelling out, or
  a premise a later change quietly broke. Prefer assertions derived from
  generated data over pinned values.
