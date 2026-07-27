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

Phases 1–3 are done (Mustering Out, per-career mechanics, cross-cutting
structures). Their detail is not repeated here: git history, the PR
descriptions, and the doc comments at each implementation site carry it,
and the comments quote the governing rule directly.

---

## Rulings

Rules questions this plan has settled. Each is recorded at its
implementation site with the governing quote; repeated here only so the
ordering below makes sense.

**#94 Scout Discovery Fame — x4 stands (p.91), against p.79's "+1".**
p.91 is not a summary: its "Mult" column carries every source's award
formula, including the non-multiplier ones ("Scholar =Rank", "Merchant
Ship Owner = 1D"). Settled; no behavior change.

**#93 Noble rank — ladder-tracked, with Soc-derivation as the fallback.**
p.65 elevates "to the next higher Noble rank and its associated increase
in Social Standing (if any)" — the "(if any)" only means anything if
rank leads and Soc follows, since Baronet/Baron, Viscount/Count and the
two Dukes each share a Soc.

**#93 Land Grants — a Mustering Out Soc increase awards one.** p.85:
"Each increase in Soc during CharGen awards a Land Grant."

## Shipped since this plan was written

**#94** (PR #97) recorded the Fame ruling; no behavior change.

**#93** (PR #98) put NobleTitle on the ladder and awarded fiefs for
Mustering Out Soc increases. Ten of thirteen careers could not hold a
Land Grant at all before it, despite Knighthood sitting on every one of
their Benefits tables.

**#36a** (PR #99) transcribed p.60's AVAILABLE SKILLS matrix from the
PDF's word bounding boxes and implemented Command College and ANM
School. #36 stays open for its pre-career half below.

**#103** was filed out of that round and closed without code: it asked
for standalone Craftsman and Functionary generators so the parity test
could cover them, but Book 1 forbids either as a first career, so a
character whose only career is Craftsman is not one the rules allow.
Both restrictions are printed — p.63's own checklist reads "Begin is
Automatic (not first career)" in the Craftsman column and "(not a first
career)" in the Functionary one, and p.87 says it twice more in prose.
The parity test guards against a career's two implementations drifting
apart; these two have one implementation each, so there was never a
risk to cover. Building the generators would have created the
duplication the test exists to catch.

Three things #36a settled are worth carrying forward, because the
next item depends on all of them:

- The p.60 matrix exists now (`character/education_skills.go`), so
  #36b inherits the skill lists rather than transcribing them.
- Its C-flagged Sciences block is the enumerable science list this
  codebase has never had, which makes "One Science" resolvable — 15
  cells beyond the 48 Major/Minor ones.
- ANM School moved level-6 skill counts barely at all (marine
  1031 → 1024). It grants distinct knowledges, not repeats, so it pushes
  nothing toward the level-6 threshold #95 needs. That still rests
  entirely on #36b's own `Major+1 per Pass`.

## 1. #36b — Education (CharGen step C)

The largest remaining chargen item, and the highest-value one.
Major/Minor cells are **8.8% of all skill-table cells** across the 13
careers (48 of 546), and every draw on one is discarded silently today.
Unresolvable cells total 11.9% (65 of 546).

Book 1 has the material despite the issue calling it unresearched.
Education is CharGen **step C** (p.72's own checklist, between B
Homeworld and D Select Career), not a characteristic tweak: p.59-61 give
18 institution rows, Apply / Pass-Fail / Waiver / Honors / Graduation
machinery, Major/Minor selection, and the skills matrix #36a
transcribes.

Note what the payoff actually is. Book 1's own footnote — "If the
character does not have a Major/Minor this benefit is lost" — means
today's silent discard is _correct_ for a character who never attended,
so the 8.8% is realized only for the educated. "One Science" becomes
resolvable at the same time (a further 15 cells): the matrix's C-flagged
Sciences block is the enumerable list this codebase has never had.

Inserting step C shifts the dice stream at the very start of every
character, so every seed-pinned fixture re-derives. It has to land in
one PR, at the identical position on both the chain and standalone
paths.

Two findings out of #36a bear on it directly. #100: p.60 marks
knowledge-only entries in bold and boldness survives no extraction
route tried so far, so every "Skill or Knowledge" grant on p.59 —
Apprenticeship's "Skill+4 or Knowledge+4", the Training Course, the
Major/Minor grants — currently cannot tell the two apart. #102: the
Educational Institution Chart costs a name and a rank die per school
attended, so deciding it after step C lands means moving the stream
twice.

## 2. #41 — Scholar Major/Minor selection and Waivers

Depends on #36. The payoff measured above is realized here.

## 3. #95 — Craftsman never reaches 40 Master Points

Zero Masterpieces across 6,000 generated chains, so QREBS and Vintage
never fire in practice.

Measure after #36b; do not implement first. The real question is
skill-level progression — every grant in this codebase is a flat +1, so
a level-6 skill needs six grants and five of them needs thirty, while
Book 1 casually assumes a Craftsman with 45 Master Points.

Education is the missing mechanism. p.59's institution table is where
Book 1 grants more than one level at a time — "Skill+4", "Major+2",
"Medic-4", "Pilot-3", "Major+1 per Pass" over four years, Honors
"Major+1", and Language at double rate. So this resolves as a
consequence of #36b rather than on its own, and the first move is to
re-run #95's own 6,000-chain measurement once step C lands.

## 4. #96 — Land Grant scope deferrals

Preferred World, geodesic hex maps, Moot proxies and voting, and grant
improvement. Independent of each other; none blocks anything above.
Preferred World and hex placement both want a world-selection concept
that does not exist yet.

## Standing checks for every PR

- Read and quote the governing rule text; the issue's paraphrase is not the source.
- Verify each new test fails against the pre-fix code before trusting it.
- When a fix changes generated output, measure the before/after distribution
  and state it in the PR rather than letting review discover it.
- Prefer assertions derived from generated data over pinned magic numbers —
  pinned fixtures have needed re-deriving after nearly every rules change.
- `task check` clean before opening.

## Lessons worth not relearning

- **Measure before believing a rules fix did anything.** #82's first
  implementation summed each career's Fame into one award, and the p.91
  cap came out **inert** — 4.2% of characters over the cap both before
  and after. That was the signal the granularity was wrong, not that the
  rule was minor. p.91 counts "Fame points _received_", so awards are
  per-instance.
- **Check whether a rule forbids the thing before building it.** #103
  asked for standalone Craftsman and Functionary generators to close a
  test-coverage gap, and the gap was real — but Book 1 forbids either as
  a first career, so the generators would have produced characters the
  rules do not allow. The tell was there before any code: the coverage
  "gap" was the absence of a second implementation, and the test in
  question exists only to catch two implementations drifting apart.
- **When a table matters, read the PDF's word coordinates, not the text
  extract.** `pdftotext -bbox-layout` gives every word an x/y box, which
  recovers the real grid outright instead of inferring it. It caught a
  live misreading in p.60/p.61: extract line 4547 reads "Command College
  | Begin vs Edu or Tra | (may not transfer to Citizen)" as one row, but
  by coordinate those are three different page columns and the middle
  phrase is the Scholar career's Begin check. It also resolved p.61's
  legend defining A, N and M twice — the column an M sits in is what
  distinguishes Medical School from Marine School. Cost about ten minutes
  and would have been a silently wrong transcription otherwise.
- **The book contradicts itself; printed tables settle it.** Masterpiece
  value is "over 40" on p.75 and "over 39" in the QREBS chapter — the
  printed value table proves 40. Marine ranks read "Coronel" where
  Soldier reads "Colonel", consistently, three times.
- **Check whether a mechanic can actually fire.** Craftsman QREBS is
  fully implemented and never once triggered in 6,000 characters.
- **A fixture that passes can still be wrong.** Four instances now, and
  every one surfaced only when an unrelated change moved the dice stream.
  Two Marine assertions passed by luck — one omitted an automatic skill
  and was saved by an unresolvable cell; another rested on a premise (C3
  stays 0) a later change quietly broke. Then Soldier's "fixture
  guarantees Risk always succeeds" turned out false, because fourteen
  terms reach age 74 and Aging erodes the fixture's 20s to End 10. And a
  Soldier skill-count equality held only because that seed lost exactly
  one draw to an unresolvable cell, cancelling out an automatic skill the
  assertion never counted. The pattern is always the same: an equality
  that happens to balance two errors. Prefer a derived bound.
