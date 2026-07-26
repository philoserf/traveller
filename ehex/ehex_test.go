package ehex

import (
	"strings"
	"testing"
)

// TestAlphabetMatchesT5Convention pins the digit set itself, since every
// other behavior in this package derives from it: 0-9 then A-Z with I
// and O omitted, because T5 avoids the characters most easily misread as
// 1 and 0. A silent change here would quietly re-encode every UWP,
// characteristic and drive rating in the codebase, so it is asserted
// directly rather than only through round-trips.
func TestAlphabetMatchesT5Convention(t *testing.T) {
	t.Parallel()

	const want = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

	if alphabet != want {
		t.Fatalf("alphabet = %q, want %q", alphabet, want)
	}

	if len(alphabet) != 34 {
		t.Errorf("len(alphabet) = %d, want 34 (0-33 inclusive)", len(alphabet))
	}

	for _, skipped := range []string{"I", "O"} {
		if strings.Contains(alphabet, skipped) {
			t.Errorf("alphabet contains %q, want it skipped (confusable with %q)",
				skipped, map[string]string{"I": "1", "O": "0"}[skipped])
		}
	}

	if Max != Value(len(alphabet)-1) {
		t.Errorf("Max = %d, want %d (last index of the alphabet)", Max, len(alphabet)-1)
	}
}

// TestValidBoundary pins the exact edge rather than sampling: Max itself
// is representable and Max+1 is not. Valid exists so callers building a
// composite string can reject bad input before Byte silently substitutes
// '?', so an off-by-one here would defeat its whole purpose.
func TestValidBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v    Value
		want bool
	}{
		{0, true},
		{9, true},
		{10, true},
		{Max - 1, true},
		{Max, true},
		{Max + 1, false},
		{255, false},
	}

	for _, c := range cases {
		if got := c.v.Valid(); got != c.want {
			t.Errorf("Value(%d).Valid() = %v, want %v", c.v, got, c.want)
		}
	}
}

// TestStringAndByteAgreeAcrossEveryValidValue walks the full range
// rather than spot-checking, and checks the two accessors against each
// other. They have separate implementations — String reads a precomputed
// cache, Byte indexes the alphabet directly — so agreeing across all 34
// values is what rules out the cache having drifted from its source.
func TestStringAndByteAgreeAcrossEveryValidValue(t *testing.T) {
	t.Parallel()

	for v := Value(0); v <= Max; v++ {
		s := v.String()
		if len(s) != 1 {
			t.Errorf("Value(%d).String() = %q, want a single character", v, s)

			continue
		}

		if b := v.Byte(); b != s[0] {
			t.Errorf("Value(%d): Byte() = %q but String() = %q — the digit cache has drifted", v, b, s)
		}

		if want := alphabet[v]; s[0] != want {
			t.Errorf("Value(%d).String() = %q, want %q", v, s, string(want))
		}
	}
}

// TestRoundTripEveryValidValue is the package's central property: any
// value that can be rendered can be read back unchanged. Everything in
// this codebase that serializes a UWP or UPP and later parses it depends
// on it holding for the whole range, not just for common digits.
func TestRoundTripEveryValidValue(t *testing.T) {
	t.Parallel()

	for v := Value(0); v <= Max; v++ {
		got, err := Parse(v.String())
		if err != nil {
			t.Errorf("Parse(Value(%d).String()) returned error %v, want it to round-trip", v, err)

			continue
		}

		if got != v {
			t.Errorf("Parse(Value(%d).String()) = %d, want %d", v, got, v)
		}
	}
}

// TestInvalidValueIsReportedNotHidden covers the deliberate asymmetry
// between the two accessors: String has room to describe the problem and
// does, while Byte must return one byte and falls back to '?'. That
// fallback is why Valid exists, so both halves are pinned — a '?' that
// silently became a real digit would hide corrupt data in a UWP string.
func TestInvalidValueIsReportedNotHidden(t *testing.T) {
	t.Parallel()

	for _, v := range []Value{Max + 1, 100, 255} {
		if got := v.Byte(); got != '?' {
			t.Errorf("Value(%d).Byte() = %q, want '?'", v, got)
		}

		got := v.String()
		if !strings.Contains(got, "invalid") {
			t.Errorf("Value(%d).String() = %q, want it to describe the value as invalid", v, got)
		}

		if len(got) == 1 {
			t.Errorf("Value(%d).String() = %q — an invalid value must not look like a real digit", v, got)
		}
	}
}

func TestParseRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"two characters", "10"},
		{"a whole UWP", "A788899-C"},
		{"skipped letter I", "I"},
		{"skipped letter O", "O"},
		{"lowercase", "a"},
		{"punctuation", "-"},
		{"space", " "},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(c.in)
			if err == nil {
				t.Fatalf("Parse(%q) = %d with no error, want an error", c.in, got)
			}

			if got != 0 {
				t.Errorf("Parse(%q) returned %d alongside its error, want the zero Value", c.in, got)
			}

			if !strings.Contains(err.Error(), "ehex") {
				t.Errorf("Parse(%q) error = %q, want it to name the package", c.in, err)
			}
		})
	}
}

// TestParseRejectsLowercaseDeliberately records that case-sensitivity is
// a decision, not an oversight: T5 writes extended hex in uppercase, and
// accepting "a" for 10 would also have to decide what "i"/"o" mean —
// characters the alphabet skips precisely to avoid ambiguity.
func TestParseRejectsLowercaseDeliberately(t *testing.T) {
	t.Parallel()

	for v := Value(10); v <= Max; v++ {
		upper := v.String()

		lower := strings.ToLower(upper)
		if lower == upper {
			continue
		}

		if _, err := Parse(lower); err == nil {
			t.Errorf("Parse(%q) succeeded, want uppercase-only parsing", lower)
		}
	}
}
