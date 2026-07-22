// Package starship models Traveller5 starship designs: hulls, drives,
// power plants, computers, fuel, hardpoints, and crew.
package starship

import "github.com/philoserf/traveller/ehex"

// Cargo is a ship's cargo capacity, broken down by hold type.
type Cargo struct {
	GeneralTons   float64
	BulkGasLiquid float64
	BulkSolid     float64
	Specialized   float64
	Sophisticated float64
	Vault         float64
	ShipsLocker   float64
}

// Accommodations is a ship's passenger/crew berthing capacity.
type Accommodations struct {
	Staterooms int
	Suites     int
	LowBerths  int
	Steerage   int
}

// Ship is a full starship design.
type Ship struct {
	Name      string
	Mission   string // 6-part mission code: Service/Activity/Type/Qualifier/Mission/Modifier
	QSP       string // Quick Ship Profile summary code
	HullTons  int
	TechLevel ehex.Value
	Cost      float64 // MCr

	Hull               Hull
	ManeuverDrive      Drive
	JumpDrive          Drive
	PowerPlant         Drive
	SupplementalDrives []Drive
	Fuel               Fuel
	Computer           Computer
	Sensors            []Sensor
	Hardpoints         []Hardpoint
	Crew               Crew
	Cargo              Cargo
	Accommodations     Accommodations
	VehiclesCarried    []string

	AnnualMaintenanceCost float64 // MCr
}
