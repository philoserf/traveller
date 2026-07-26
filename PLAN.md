# Character generation plan

Remaining work on **character generation only**. Starship generation is
out of scope — see "Tabled" below.

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

## 1. #93 — Noble rank and NobleTitle() disagree

A live defect, not a deferral: the two disagree for 25.7% of generated
Nobles, because Mustering Out raises Soc after the career ends and the
Soc-derived title outruns the ladder the career walked.

Do this first. It is small and self-contained, and it sits underneath
#36 — the largest remaining change — so leaving it means re-deriving
Noble fixtures twice.

Carries a rules question: whether noble rank is Soc-derived or
ladder-tracked, and whether a Mustering Out Soc increase awards a Land
Grant (p.85: "Each increase in Soc during CharGen awards a Land Grant").

## 2. #94 — Scout Discovery Fame: +1 or x4

> p.79: a Discovery gives "a Land Grant, and Fame +1."
> p.91 Fame table: "Scout — Discoveries — x4"

Needs a decision, not research. The code does x4, which #82 rebuilt the
whole Fame system on. Cheap to settle and worth settling before #36,
since Education changes how characteristics move.

This codebase's own precedent cuts both ways — "a career's own box beats
the generic summary" favors p.79, but p.91 is the Fame chapter's
dedicated table with a Mult column, not a summary.

## 3. #36 — Command College and Education

The largest remaining chargen item, and the highest-value one.
Major/Minor cells are **8.8% of all skill-table cells** across the 13
careers (48 of 546), and every draw on one is discarded silently today.
Unresolvable cells total 12.1%.

Book 1 has the material despite the issue calling it unresearched: 110
Education references, 14 for Command College.

Scope needs a decision at the research stage. Education is not just a
characteristic — Command College is deferred at O4 in Marine, Soldier
and Spacer, and "Resolve ANM School as Education" sits in all three
Operations tables. Settle how far that runs before building.

## 4. #41 — Scholar Major/Minor selection and Waivers

Depends on #36. The payoff measured above is realized here.

## 5. #95 — Craftsman never reaches 40 Master Points

Zero Masterpieces across 6,000 generated chains, so QREBS and Vintage
never fire in practice.

Do this after #41, not before: the real question is skill-level
progression — every grant in this codebase is a flat +1, so a level-6
skill needs six grants and five of them needs thirty, while Book 1
casually assumes a Craftsman with 45 Master Points. #36/#41 change what
grants skills, so measure after they land.

## 6. #96 — Land Grant scope deferrals

Preferred World, geodesic hex maps, Moot proxies and voting, and grant
improvement. Independent of each other; none blocks anything above.
Preferred World and hex placement both want a world-selection concept
that does not exist yet.

---

## Tabled

**#6 — shipgen.** A separate subsystem against Book 2, not character
generation. The `starship` package is types and constants today. Out of
scope for this plan; re-plan it on its own terms when chargen closes,
scoped into a sequence of PRs rather than one issue.

---

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
