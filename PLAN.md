# Traveller backlog plan

Ranked working order for the 17 open issues. Every entry below was checked
against `reference/Traveller5 Core Rules Book 1 Characters and Combat.txt`
(and Book 2 for shipgen) — the page text is quoted where it decides scope.

**Working rule for each item:** read the cited rule text first, implement
against it, and quote it in the PR. Do not implement from the issue text
alone — several issues paraphrase the rules inaccurately.

---

## Phase 1 — Mustering Out cluster — DONE

All four merged: #55 (PR #77), #57 (#78), #45 (#81), #56 (#80).

<details>
<summary>Original plan detail</summary>

These four all live in `career_muster_out.go` / `muster_out_apply.go`.
Doing them adjacently avoids four rounds of churn through the same code.
Rule text for all four is confirmed and unambiguous.

### 1. #55 — extra Mustering Out rolls

> p.68: "one Mustering Out roll for each term served… one additional roll
> per Commendation, MCG, or SEH. He is allowed one additional roll if Fame 19+."

`scoutMusterOutRollCount` implements terms + Disability doubling only.
Note the rule names **MCG or SEH specifically** — not XS or MCUF, which the
Armed Forces careers grant far more often. Agent Commendations are already
recorded on terms.
Also confirms: Fame 19+ grants its roll, and p.68 lets that roll pick any
eligible career table.
_Size: medium. Start here — it changes roll counts everything else builds on._

### 2. #57 (remainder) — Knighthood and Forbidden Knowledge

> p.68: "Knighthood. The character receives a Knighthood (= Soc B if the
> character has C6= Soc)." … "If the improvement is C6+1 and for the
> character C6= Caste, the benefit is lost."
> p.69: "Forbidden Knowledge… Each receipt provides skill-1."

The characteristic cap half shipped in #76. Knighthood is fully specified:

> p.68: "A Knighthood raises any value of Soc to B; if the character is
> already Soc 11+, he receives Soc +1 instead."
> p.68: "In the Spacer, Soldier, and Marine careers, Knighthood is only
> available to Officers. A non-officer receives Soc +1 (even if it advances
> Soc to 11 or beyond)."

The enlisted restriction the issue described is real — I initially reported
it as unfound, having stopped reading a paragraph too early. Both halves are
confirmed; no open question remains.

Forbidden Knowledge grants skill-1 but the text names no table, so decide
and document how the skill is chosen.
_Size: small–medium._

### 3. #45 — reroll duplicate Benefits

> p.68: "Duplicate benefits may be rerolled."
> p.69: "A result that duplicates a previous (unwanted or unusable) benefit
> may be rerolled until a different benefit is received, for example: Wafer
> Jack, TAS Member, Knighthood."

The p.69 examples effectively define the "unique" set; cumulative results
(Ship Share, characteristic +1, Fame +2, cash) stay repeatable. "May" is a
player option — resolve it with this codebase's established convention for
open choices, and guard against an exhausted table looping forever.
_Size: medium._

### 4. #56 — pensions and retirement

> p.70: Citizen Cr5,000/yr; Functionary Cr15,000/yr (replaces Citizen's);
> Reserve Cr100/yr; tenured Professor's pension; Enlisted retirement
> Cr2,000/term and Officer Cr3,000/term, both requiring 4+ terms.
> "A pension begins when a character reaches Life Stage 9 Retirement (= age
> 66 for Humans)." Duplicate entitlements are allowed. "Any Entitlement can
> be cashed out for a lump sum equal to five years of payments."

`MusteringOut.Pension` and `.RetirementPay` already exist and are never set.
Note the Life-Stage-9 start is a real gate — most generated characters never
reach 66. Render annual amounts as income, not cash.
_Size: medium._

</details>

---

## Phase 2 — per-career mechanics — DONE

#43 (PR #83), #39 (verified correct, closed without change), #42 (#84),
#40 (#85), #59 (#86), #44 (#89).

<details>
<summary>Original plan detail</summary>

Self-contained; each touches one career's own file. Order within the phase
is by confidence in the rule text, highest first.

### 5. #43 — Merchant Ship Owner Fame

> p.91 Fame table: "Merchant — Ship Owner = 1D"

Fame award is confirmed. **Open question:** the threshold at which
accumulated Ship Shares become ownership — find it before implementing.

### 6. #39 — Functionary F6 rank titles

Locate p.87's F6 title table and cover every preceding career this codebase
supports; the generic "Director" fallback stays only for genuinely unnamed
combinations.

### 7. #42 — Entertainer optional Flux rolls and Comeback

### 8. #40 — Rogue multi-term prison sentences

### 9. #59 — Rogue Scheme Flux adjustment and previous-career selection

### 10. #44 — Agent Undercover Assignment table and A/B/C mechanic

Largest of this phase: a full table plus a three-die mechanic.

</details>

---

## Phase 3 — cross-cutting structures — DONE

#58+#37 (PR #90), #54 (#91), #35 (#92).

<details>
<summary>Original plan detail</summary>

### 11. #58 + #37 — Land Grants (do together)

42 references in Book 1. Scout Discoveries and Noble both need the same
structured Land Grant, and `noble_generate.go`'s "Capital" skill cell is
blocked on it too. Building a Scout-only version first would make the shared
one harder — this is why #58's remaining half was deferred rather than
half-built.

### 12. #54 — Armed Forces term skills restricted to received Operations columns

Restructures how term skills are drawn: retain all four Operations results
per term for eligibility while still using only the highest Mod for
Risk/Reward. Expect fixture churn.

### 13. #35 — Craftsman QREBS and Vintage appreciation

Needs a structured item record and a time-since-creation concept, neither of
which exists yet.

---

</details>

## Phase 4 — foundational and large

### 14. #36 — Command College and Education

Currently the only issue the codebase itself calls unresearched. Education is
a prerequisite for #41's Major/Minor, and Command College is deferred in
Marine/Soldier/Spacer at O4. Do this before #41.

### 15. #41 — Scholar Major/Minor selection and Waivers

Depends on #36.

### 16. #6 — shipgen

A whole subsystem against Book 2 (20k lines). The `starship` package is
types and constants only — one function today. Scope into its own sequence
of PRs rather than treating it as one issue.

---

## Found along the way

### #82 — Fame Stacks cap — DONE (PR #88)

Resolved via p.91's own Fame descriptor scale (0 Unknown … 19 Subsector,
20 Sector, 21 Domain … 36 All Reality): Fame is a scale of *reach*, so
local fames accumulate to Sector-wide and beyond that only the greatest
reach counts — `max(min(sum, 20), highest)`.

The lesson worth keeping: the first implementation summed each career's
Fame into one award, and measuring showed the cap was **inert** (4.2% of
characters over 20 both before and after). That was the signal the
granularity was wrong, not that the rule was minor. p.91 counts "Fame
points *received*", so awards are per-instance. Measure before believing
a rules fix did anything.

## Open rules questions

Two conflicts found in Phase 3, flagged rather than silently resolved:

- **Scout Discovery Fame.** p.79 says a Discovery gives "Fame +1";
  p.91's Fame table says "Scout — Discoveries — x4". The code implements
  x4, which #82/#88 rebuilt the whole Fame system on. Unresolved; no
  behavior changed.
- **Craftsman Master Points are effectively unreachable.** p.75 needs
  CC + Craftsman + five level-6 skills to total 40. Across 6,000
  generated chains none reached it, so QREBS and Vintage never fire in
  practice. Faithful to the rules as written; open question whether
  skill progression should make it reachable.

## Standing checks for every PR

- Read and quote the governing rule text; the issue's paraphrase is not the source.
- Verify each new test fails against the pre-fix code before trusting it.
- When a fix changes generated output, measure the before/after distribution
  and state it in the PR rather than letting review discover it.
- Prefer assertions derived from generated data over pinned magic numbers —
  pinned fixtures have needed re-deriving after nearly every rules change.
- `task check` clean before opening.
