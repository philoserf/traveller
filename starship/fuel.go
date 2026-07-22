package starship

// FuelType is the commodity a power or drive system consumes.
type FuelType string

// FuelType values.
const (
	FuelLiquidHydrogen  FuelType = "LiquidHydrogen"  // Power Plant
	FuelRadioactiveRods FuelType = "RadioactiveRods" // Fission
	FuelAntiMatterSlugs FuelType = "AntiMatterSlugs"
	FuelExoticParticles FuelType = "ExoticParticles" // Collector
)

// FuelFitting is equipment for collecting, processing, or storing fuel.
type FuelFitting string

// FuelFitting values.
const (
	FuelScoop    FuelFitting = "FuelScoop"
	FuelIntake   FuelFitting = "FuelIntake"
	FuelBin      FuelFitting = "FuelBin"
	FuelPurifier FuelFitting = "FuelPurifier"
	TransferPump FuelFitting = "TransferPump"
)

// Fuel is a ship's fuel tankage and consumption profile.
type Fuel struct {
	Type           FuelType
	Capacity       float64 // tons
	JumpFuel       float64 // tons consumed per jump
	OperationsFuel float64 // tons consumed per month of power plant operation
	Fittings       []FuelFitting
}
