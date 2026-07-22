package starship

import "github.com/philoserf/traveller/ehex"

// Computer is a ship's onboard computer system.
type Computer struct {
	Model   string // "0", "0bis", "1" .. "9", "9bis"
	Cells   int
	Tons    float64
	Squares int
	Cost    float64 // MCr
	BaseTL  ehex.Value
}
