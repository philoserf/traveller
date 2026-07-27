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
School.

**#100** (PR #106) established that p.60's "Bold= Knowledge-Only skill"
marking _is_ recoverable — from the PDF's font subsets, though from no
text extract — and that it marks nine parent skills rather than any
entry, which p.61 explains: "Education or Training can only impart the
Knowledges; the Skills themselves are not obtainable." p.61 also
enumerates each parent's Knowledges independently of p.60, and those
counts match the transcription row for row, which is the only
independent check the specialty blocks have.

**#36b** shipped in two stages. PR #108 built CharGen step C — the
academic spine, ED5 through University, with Apply / Pass-Fail / Waiver
/ Graduation and Major/Minor selection — inserted through a shared
`generateStart` so all twelve entry points draw it identically. PR #109
then resolved the 48 Major/Minor career-table cells, at final assembly
rather than where they are rolled, which costs no dice at all.

Their combined effect is larger than either alone, and that is worth
keeping: a Major leaves Education at 4 or 5, and it is the career cells
that then push it past 6, because they grant the same subject again
rather than a random one. Level-6 skill entries roughly doubled for
Scout and Merchant. Measuring step C by itself showed only +8% and was
misleading.

#36 stays open for what both stages deferred: Service Academy, OTC and
NOTC (they confer Commissions and oblige service), the degree-chained
schools (Masters, Professors, Medical, Law), Flight School, and the Tra
path.

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

## 1. #36 — the rest of Education

The academic spine shipped (see above). What is left, each deferred
because it pulls in a mechanic of its own rather than more of Education:

- **Service Academy, OTC, NOTC.** They confer an Army, Navy or Marine
  Commission and oblige a term of service — "The character is required
  to serve one term in the service... he is in the Reserves" — which is
  a career interaction. Read alongside #110, which is about exactly that
  sentence being applied too widely.
- **Masters, Professors, Medical School, Law School.** These gate on the
  degree step C now produces (BA, MA, Honors BA), so they chain off it.
  Masters shares the "Minor+1 per 2 Passes" cell already implemented.
- **Flight School.** Gates on an Honors BA, grants Pilot-3 and a Flight
  Branch.
- **The Tra path** — Apprenticeship, Mentor, Training Course. p.59 puts
  Sophonts with Tra there and lets Humans use Training Courses at Edu/2.
  `GenerateUPP` produces only Humans, whose C5 is always Edu, so none
  has a caller yet; this is really a "generate non-Humans" item wearing
  an Education hat.

Two smaller cells are also still unresolved and are cheap now.
"One Science" has an enumerable list at last — p.60's C-flagged Sciences
block — worth 15 cells. "Capital" is #101. Both draw a die, so both move
every character's stream; do them together and measure once.

## 2. #41 — Scholar Major/Minor selection and Waivers

Depends on #36. The payoff measured above is realized here.

## 3. #110 — career chains cap every non-final career at one term

Found while re-measuring #95, and now in front of it.

`GenerateCareerChainCharacter` gives every career but the last a single
term. Book 1 has a one-term obligation but p.61 scopes it to
commissioned academy graduates — "The character is required to serve one
term in the service" — not to every career a player lists. The cap has
no other rules basis I could find and no recorded rationale.

It is why Craftsman is unreachable: `BeginCraftsman` is a prerequisite,
not a roll ("IF two Skill-6+ and Craftsman-1"), and one term per prior
career cannot produce two level-6 skills. Zero Craftsman terms served
across ~18,000 characters and six chain shapes.

Expect it to move a lot of generated output, and to affect Functionary
in the opposite direction — its Begin target is Total Terms x3, so a
short chain makes it easier to enter than intended.

## 4. #95 — Craftsman never reaches 40 Master Points

Re-measured after #36b. The original diagnosis was right and #36b
confirmed it: characters satisfying both `BeginCraftsman` prerequisites
went from 4.32% to **10.18%**, and those holding two level-6 skills from
10.7% to 22.5%. Education was the missing mechanism.

Two things still sit in front of the Masterpiece gate. #110 above stops
the career beginning at all. And even past that, five _distinct_ level-6
skills remains out of reach: best Master Points seen is 38 against a
40-point gate, and no character in 6,000 holds more than four qualifying
skills.

Do not measure Masterpieces again until #110 is settled — the career
cannot run, so the figure would mean nothing.

## 5. #96 — Land Grant scope deferrals

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
- **Measure the halves together when they compound.** Step C on its own
  moved level-6 skill counts by 8%, which read as a disappointment and
  was reported as one. Resolving the Major/Minor career cells then moved
  them by 39-125%. Neither number is wrong; the mechanism is that
  Education leaves a Major at 4 or 5 and the career cells grant the same
  subject again. A staged change measured only stage by stage will
  understate itself.
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
