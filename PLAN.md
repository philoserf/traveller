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

#36 closed with them; #113 carries what they deferred.

**#41** (PR #115) gave Scholars their Major, their Minor and their
Waivers, all three stated outright on p.76 — including the rule that
makes Scholar the one career whose Major/Minor cells always resolve:
"Every Scholar has a Major and a Minor. If no degree... then select any
Skill or Knowledge from the Skills List." 461 of 3,000 Scholars reach
the career without a degree, and were losing 1,830 cells to a rule
saying they have both.

**#95** (PR #116) closed with no production change: the mechanic was
never broken, only unreachable, and #110 fixed that. 126 Craftsmen now
serve across 6,000 chains and one creates a Masterpiece — 42 Master
Points, QREBS allocated, sold at Cr170,000 and Vintage-appreciated to
Cr176,800. Every line of PR #92's code runs in generated output at last.

Correcting what this plan said twice: the gate does NOT need five
distinct level-6 skills. The formula sums the levels of _up to_ five
qualifying skills — "up to FIVE" is a cap, not a requirement — and the
Masterpiece above clears 40 with four (Trader-10, Electronics-9,
Designer-8, Computer-7, totalling 34). High levels substitute for count.

**#110** (PR #112) let each listed career run to its own natural end
instead of being cut to a single term. The cap it removed had no rules
basis — Book 1's own one-term obligation is p.61's, and belongs to
commissioned academy graduates. Removing it exposed a real defect
underneath: with no -age target, segmentBudget handed every career a
fresh fourteen terms rather than one budget for a life, so an uncapped
two-career chain reached age 130. "No target" now means "the end of
p.89's table".

Two things about that change are worth keeping in mind while reading
generated output. Long chains rarely complete now — a five-career chain
reaches its last entry 13% of the time, because a first career that runs
well consumes most of a life — and that is the rule working. And the old
behaviour had an artifact it is easy to miss in hindsight: every
citizen,craftsman character served exactly one term and stopped at age
22, at 100% consistency.

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

**#101** (PR #117) resolved the last two career-table cells still being
discarded: "One Science" now draws from p.60's C-flagged Sciences block
(#36a), and "Capital" — p.85's "World Knowledge (of world of highest
held noble Land Grant) (value= 1D)" — records the 1D at the roll and
substitutes the world UWP at final assembly, the same marker-then-
substitute pattern #36b used for Major/Minor. World Knowledge entries
appear in generated output for the first time (Noble, Agent). Worlds
are identified by UWP (`World: A867A69-B`), the only identity this
codebase gives them; #118 tracks giving worlds real names.

**#102** (PR #119) transcribed Book 1 p.92-93's Educational Institution
Chart (twelve sub-tables, a 1D name roll plus a rank value — a die for
eleven of them, ED5's own "Rank= Inconsequential" for the twelfth) and
wired School Rank into every institution step C attends — 252 of 307
graduates carry one, 500-seed sample. School _names_ mostly do not
render: the chart's templates need a world/city/province/company/colour
this codebase has no source for. The 1D that picks the template is
still recorded (`Education.SchoolNameRoll`), so filling names in later
costs no dice. Command College and ANM School are the two institutions
whose names resolve completely today but aren't recorded yet (they'd
need a name on `Term`, not `Education`); place names generally are
#118.

**#113, items 1-3** shipped in three PRs, the largest remaining chargen
work at the time:

PR #120 built **Service Academy, OTC, NOTC** — each confers a military
Commission (p.61), entering Marine/Soldier/Spacer term 1 already
Officer1 and bypassing `BeginMarine`/`BeginSoldier`/`BeginSpacer`.
NOTC dominates OTC in this codebase's model (Navy-or-Marine vs.
Army-only, identical cost) and is always taken, the open-choice
convention `chooseInstitution` already used. `commissionAppliesTo`
enforces "consumed once" across a career chain. 6,000-seed sweep:
Marine 61.6%, Soldier 29.3%, Spacer 64.1% commissioned-entry — a large
shift, driven by University's own pre-existing ~53% share combined
with NOTC succeeding often for that population.

PR #121 built **Flight School** — Pilot-3 and a Flight Branch, p.61's
own worked example implemented literally down to the shortened first
term ("Because Flight School took a year, this first Term is reduced
to three years"), the first real use of `Term.Length` off its
hardcoded `4`. Also fixed a live bug PR #120 had shipped: a
Commissioned entry's own term 1 could still spuriously fire an
internal Enlisted Promotion roll, granting a second wrong-track
auto-skill and computing the wrong `Rank` string — invisible in #120's
own all-zero-UPP test, where that internal roll always failed too.
6,000-seed sweep of commissioned entrants: Marine 60.8%, Soldier
16.0%, Spacer 62.9% also reach Flight School — Soldier's Commission
path is Service Academy-only (NOTC doesn't reach Army), so it never
gets NOTC's higher-Edu Honors population.

PR #122 built **Masters, Professors, Medical School, Law School** —
University-only (both governing sentences name University
specifically), gated on the `Degree` step C already produces, and
Waiver-eligible like every other Prerequisite. A first draft ordered
Medical before Law as a fixed preference, mirroring `chooseInstitution`'s
demanding-first convention; measurement caught it immediately (1,791
Doctors vs. 12 Attorneys across 6,000 graduates, since Medical's
prerequisite is checked, and taken, first) — unlike Trade-vs-College or
OTC-vs-NOTC, Medical and Law don't actually dominate each other, so the
pick is uniform-random instead, `rollScoutDuty`'s own convention for
this shape of choice. Re-measured: Doctor 926, Attorney 948.

**#113, item 5** (PR #160, #161, #162 — three stages) shipped Later
Education — p.59: "At the beginning of any term, the character may
apply for any Educational Institution or Training, and if accepted
substitutes that process for the entire term." Distinct from ANM
School/Command College (assigned, don't consume a whole term) —
`character/command_college.go`'s own doc comment already named this
rule as the reason it avoided partial-term modeling, effectively a
forward-note for this work.

Staged to keep the risky part isolated: stage 1 added `termsServed`
(Later Education terms are elapsed, not served — every DM/roll-count/
cash-amount keyed on career length needed the distinction, 13 files),
stage 2 built the shared `beforeTerm` hook on `resolveCareerLoop` plus
the gate (`shouldAttemptLaterEducation`: retry after a non-graduation,
or escalate to a strictly better institution now reachable) fully
tested but unwired, stage 3 wired the real mechanism into Rogue as the
pilot. Both of the first two stages measured as truly inert before the
third one changed anything.

Rogue was the pilot because it has its own dedicated segment resolver
— Scout/Marine/Soldier/Spacer/Agent share `resolveRiskCareerSegment`'s
generic signature, which every implementor would have needed an unused
`Education` parameter to accommodate just to pilot one of them.

The pilot PR caught a real bug before it shipped: `careerSegment` never
carried a segment's own `Education` back out of a career chain, so a
chain-resolved Rogue's Later Education attendance vanished at final
assembly — `TestCareerChainSingleEntryMatchesLegacyGenerator` caught it
immediately (chain and standalone diverged on the same seed). Fixed by
returning `Education` from all nine `careerSegment` construction sites
and threading it forward through the chain loop, which also mostly
closed the "doesn't propagate to the next segment" gap rather than
leaving it as a documented limitation.

6,000-seed sweep of `GenerateRogueCharacter`: 32.0% elect Later
Education at least once (~1.9 terms each, among those who do). Mean
Skills 15.88 vs. 12.45 for those who don't (+27.5%); mean Fame 10.76
vs. 5.52. Not purely causal — the gate's retry branch preferentially
fires for characters who failed step C admission, a different starting
population — but the magnitude is real.

Three follow-ups split out rather than folded into #113 itself, so the
issue could close on the four pieces that actually shipped:

- **#163** — item 4, the Tra path (Apprenticeship, Mentor, Training
  Course). Needs non-Human generation, which doesn't exist anywhere in
  this codebase; #113's own text already said this "probably should not
  be built" until it does.
- **#164** — wire Later Education into Entertainer/Citizen/Noble, which
  hand-roll their own career loops instead of sharing
  `resolveCareerLoop`.
- **#165** — step C itself attending more than one institution in
  sequence (e.g. ED5 raising Edu to 5, then continuing straight into
  Trade School), distinct from Later Education's own mid-career
  mechanism.

## 1. #96 — Land Grant scope deferrals

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
- **A mechanic can be correct, tested, and still never run.** QREBS and
  Vintage were fully implemented and unit-tested, and fired zero times in
  generated output for months, because the career that uses them could
  not be entered. No unit test could have caught it — the code was right.
  What catches it is a test that walks the whole path and asserts the
  thing happens at all. #95 now has one.
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
