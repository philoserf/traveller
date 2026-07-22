// Package ehex implements Traveller5's "extended hex" digit: a single
// character encoding values 0-33, used throughout T5 for UWP fields,
// characteristics, armor ratings, and drive/component sizes.
package ehex

import (
	"fmt"
	"strings"
)

// Value is an extended-hex digit in the range 0-33.
type Value uint8

// alphabet is the digit set 0-9 then A-Z with I and O skipped, matching
// T5's convention of avoiding characters easily confused with 1 and 0.
const alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// Max is the highest value a single extended-hex digit can represent.
const Max = Value(len(alphabet) - 1)

// digits precomputes each valid Value's single-character string, so String
// returns a cached string instead of allocating one on every call.
var digits = func() [Max + 1]string {
	var d [Max + 1]string
	for i := range d {
		d[i] = string(alphabet[i])
	}

	return d
}()

// Byte returns the single-character extended-hex representation as a byte.
func (v Value) Byte() byte {
	if v > Max {
		return '?'
	}

	return alphabet[v]
}

// String returns the single-character extended-hex representation.
func (v Value) String() string {
	if v > Max {
		return fmt.Sprintf("<invalid ehex %d>", uint8(v))
	}

	return digits[v]
}

// Parse converts a single extended-hex character into its Value.
func Parse(s string) (Value, error) {
	if len(s) != 1 {
		return 0, fmt.Errorf("ehex: %q is not a single character", s)
	}

	i := strings.IndexByte(alphabet, s[0])
	if i < 0 || i > int(Max) {
		return 0, fmt.Errorf("ehex: %q is not a valid extended-hex digit", s)
	}

	return Value(i), nil
}
