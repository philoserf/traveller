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

## 1. #93 — Noble rank and NobleTitle() disagree

A live defect, not a deferral: the two disagree for 25.7% of generated
Nobles, because Mustering Out raises Soc after the career ends and the
Soc-derived title outruns the ladder the career walked.

Do this first. It is small and self-contained, and it sits underneath
#36 — the largest remaining change — so leaving it means re-deriving
Noble fixtures twice.

Both halves are ruled above. Keep Soc-derivation for characters who
never walked the ladder: that is what makes p.68's Knighthood confer a
title from any career's Mustering Out.

## 2. #36a — Command College and ANM School

The in-career half of #36, separable from Education proper and much
smaller than its own deferral comments assume.

> p.61: "A Character must attend Command College in the first year of
> the term after he is promoted to Officer4, provided he successfully
> Continues. A character who fails Command College may not Continue in
> the service. Success at Command College awards two skill levels from
> the appropriate Military or Naval Academy."

Both grants are flat +1, so no multi-level machinery is needed yet. The
term is already four years and p.59's Duration is one year _inside_ it,
so no partial term is required either — the structural change the
deferral comments feared does not arise. Touches only Marine, Soldier
and Spacer, so it perturbs no pre-career dice.

Needs p.60's AVAILABLE SKILLS matrix transcribed first — five page
columns interleaved with category labels spliced into data rows, the
character-offset hazard in its worst form.

## 3. #36b — Education (CharGen step C)

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

## 4. #41 — Scholar Major/Minor selection and Waivers

Depends on #36. The payoff measured above is realized here.

## 5. #95 — Craftsman never reaches 40 Master Points

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

## 6. #96 — Land Grant scope deferrals

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
- **The book contradicts itself; printed tables settle it.** Masterpiece
  value is "over 40" on p.75 and "over 39" in the QREBS chapter — the
  printed value table proves 40. Marine ranks read "Coronel" where
  Soldier reads "Colonel", consistently, three times.
- **Check whether a mechanic can actually fire.** Craftsman QREBS is
  fully implemented and never once triggered in 6,000 characters.
- **A fixture that passes can still be wrong.** Two Marine assertions
  passed only by luck — one omitted an automatic skill and was saved by
  an unresolvable cell; another rested on a premise (C3 stays 0) that a
  later change quietly broke.
