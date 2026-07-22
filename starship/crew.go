package starship

// CrewPosition is a shipboard crew role.
type CrewPosition string

const (
	Captain       CrewPosition = "Captain"
	Exec          CrewPosition = "Exec"
	Pilot         CrewPosition = "Pilot"
	Astrogator    CrewPosition = "Astrogator"
	ChiefEngineer CrewPosition = "ChiefEngineer"
	Engineer      CrewPosition = "Engineer"
	DriveTech     CrewPosition = "DriveTech"
	RadTech       CrewPosition = "RadTech"

	Sensop CrewPosition = "Sensop"
	Comms  CrewPosition = "Comms"
	ITTech CrewPosition = "ITTech"

	Gunner CrewPosition = "Gunner"
	Loader CrewPosition = "Loader"

	LifeSupportTech CrewPosition = "LifeSupportTech"
	Driver          CrewPosition = "Driver"
	Cook            CrewPosition = "Cook"
	Security        CrewPosition = "Security"

	Purser        CrewPosition = "Purser"
	Steward       CrewPosition = "Steward"
	Freightmaster CrewPosition = "Freightmaster"
	Medic         CrewPosition = "Medic"

	Counsellor       CrewPosition = "Counsellor"
	PoliticalOfficer CrewPosition = "PoliticalOfficer"
	Specialist       CrewPosition = "Specialist"
)

// StaffingLevel is the crew multiplier applied per staffed console.
type StaffingLevel int

const (
	StaffingMinimal     StaffingLevel = iota // 1/3
	StaffingOneShift                         // 1x
	StaffingTwoShifts                        // 2x
	StaffingThreeShifts                      // 3x
)

// Crew is a ship's complement.
type Crew struct {
	Positions         []CrewPosition
	TotalCrew         int
	StaffingLevel     StaffingLevel
	AccommodationTons float64
}
