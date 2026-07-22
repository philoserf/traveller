// Package ehex implements Traveller5's "extended hex" digit: a single
// character encoding values 0-33, used throughout T5 for UWP fields,
// characteristics, armor ratings, and drive/component sizes.
package ehex

import "fmt"

// Value is an extended-hex digit in the range 0-33.
type Value uint8

// alphabet is the digit set 0-9 then A-Z with I and O skipped, matching
// T5's convention of avoiding characters easily confused with 1 and 0.
const alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// Max is the highest value a single extended-hex digit can represent.
const Max = Value(len(alphabet) - 1)

// String returns the single-character extended-hex representation.
func (v Value) String() string {
	if v > Max {
		return fmt.Sprintf("<invalid ehex %d>", uint8(v))
	}

	return string(alphabet[v])
}

// Parse converts a single extended-hex character into its Value.
func Parse(s string) (Value, error) {
	if len(s) != 1 {
		return 0, fmt.Errorf("ehex: %q is not a single character", s)
	}

	c := s[0]
	for i := range len(alphabet) {
		if alphabet[i] == c {
			return Value(i), nil
		}
	}

	return 0, fmt.Errorf("ehex: %q is not a valid extended-hex digit", s)
}
