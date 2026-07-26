package character

import "testing"

// TestResolveFameStacksBelowCapIsPlainSum records that the cap is
// inert in the ordinary case: Book 1 p.91 sums Fame points received,
// and only the "to 20" clause makes it anything other than addition.
func TestResolveFameStacksBelowCapIsPlainSum(t *testing.T) {
	t.Parallel()

	cases := [][]int{nil, {}, {0}, {7}, {1, 2, 3}, {4, 4, 4, 4, 4}, {20}, {10, 10}}

	for _, awards := range cases {
		if got, want := resolveFameStacks(awards), sumInts(awards); got != want {
			t.Errorf("resolveFameStacks(%v) = %d, want %d (sum is <= the cap)", awards, got, want)
		}
	}
}

// TestResolveFameStacksCapsManySmallAwards is the clause's whole point,
// and the case a plain sum gets wrong: twelve Scout Discoveries are
// twelve awards of 4, and p.91 sums "Fame points received to 20".
func TestResolveFameStacksCapsManySmallAwards(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		awards []int
	}{
		{"twelve Discoveries at 4 each", []int{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}},
		{"just over the cap", []int{20, 1}},
		{"many ones", []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveFameStacks(c.awards); got != fameStackCap {
				t.Errorf("resolveFameStacks(%v) = %d, want %d", c.awards, got, fameStackCap)
			}

			if sum := sumInts(c.awards); sum <= fameStackCap {
				t.Fatalf("test case is not actually over the cap: sum = %d", sum)
			}
		})
	}
}

// TestResolveFameStacksLetsASingleLargeAwardStand covers p.91's second
// clause — "beyond 20, only the highest Fame applies". Without it an
// Entertainer whose p.77 Fame track reached 47 would be reported as 20,
// which is the cap lowering a Fame the rules awarded outright.
func TestResolveFameStacksLetsASingleLargeAwardStand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		awards []int
		want   int
	}{
		{[]int{47}, 47},
		{[]int{47, 2}, 47},
		{[]int{2, 47, 3}, 47},
		{[]int{21, 21}, 21},
	}

	for _, c := range cases {
		if got := resolveFameStacks(c.awards); got != c.want {
			t.Errorf("resolveFameStacks(%v) = %d, want %d", c.awards, got, c.want)
		}
	}
}

// TestResolveFameStacksIsMonotonic is the property that rules out the
// reading this rule was filed against: under a bare min(sum, 20),
// receiving *another* Fame award could never lower a character's Fame,
// but under a bare "highest applies" it could. Neither clause alone is
// safe, so the combination is asserted directly — an extra award must
// never make a character less famous.
func TestResolveFameStacksIsMonotonic(t *testing.T) {
	t.Parallel()

	base := [][]int{nil, {3}, {4, 4, 4}, {19}, {25}, {30, 1}, {5, 5, 5, 5, 5}}

	for _, awards := range base {
		before := resolveFameStacks(awards)

		for _, extra := range []int{0, 1, 2, 5, 19, 20, 40} {
			grown := append(append([]int(nil), awards...), extra)

			if after := resolveFameStacks(grown); after < before {
				t.Errorf("resolveFameStacks(%v) = %d then %v = %d — an award lowered Fame",
					awards, before, grown, after)
			}
		}
	}
}

// TestResolveFameStacksIgnoresOrder guards the accumulator plumbing:
// awards reach this function in whatever order careers resolved, and a
// running-total implementation that clamped as it went would not be
// order-independent.
func TestResolveFameStacksIgnoresOrder(t *testing.T) {
	t.Parallel()

	forward := []int{4, 30, 1, 4, 2}
	reverse := []int{2, 4, 1, 30, 4}

	if a, b := resolveFameStacks(forward), resolveFameStacks(reverse); a != b {
		t.Errorf("resolveFameStacks(%v) = %d but reversed = %d", forward, a, b)
	}
}
