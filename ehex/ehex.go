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

// Valid reports whether v is within the representable extended-hex range.
// Callers building a composite string from several Values should check this
// first: Byte, unlike String, has no room to signal an invalid digit and
// falls back to '?', which would otherwise hide corrupt data.
func (v Value) Valid() bool {
	return v <= Max
}

// Byte returns the single-character extended-hex representation as a byte.
// For an invalid Value it returns '?' — callers that need to detect and
// report invalid digits should check Valid first and use String instead.
func (v Value) Byte() byte {
	if !v.Valid() {
		return '?'
	}

	return alphabet[v]
}

// String returns the single-character extended-hex representation.
func (v Value) String() string {
	if !v.Valid() {
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
	if i < 0 {
		return 0, fmt.Errorf("ehex: %q is not a valid extended-hex digit", s)
	}

	// i is in [0, len(alphabet)-1] by IndexByte's contract on a fixed 34-byte
	// alphabet, so this always fits in Value (uint8); gosec can't see that.
	return Value(i), nil //nolint:gosec // bounded by len(alphabet)==34, see comment above
}
