package starship

// Configuration is a hull's shape/streamlining, which drives friction,
// agility, max G, stability, and land-capability.
type Configuration byte

const (
	ConfigCluster       Configuration = 'C'
	ConfigBracedCluster Configuration = 'B'
	ConfigPlanetoid     Configuration = 'P'
	ConfigUnstreamlined Configuration = 'U'
	ConfigStreamlined   Configuration = 'S'
	ConfigAirframe      Configuration = 'A'
	ConfigLiftingBody   Configuration = 'L'
)

// Structure is a hull's construction technique, which determines the base
// Armor Value formula and required Tech Level.
type Structure byte

const (
	StructurePlate   Structure = 'A'
	StructureShell   Structure = 'S'
	StructurePolymer Structure = 'P'
	StructureFeNi    Structure = 'F'
	StructureOrganic Structure = 'O'
	StructureCharged Structure = 'C'
)

// JumpFieldType affects safe jump distance, armor modifier, and jump flash size.
type JumpFieldType int

const (
	JumpFieldBubble JumpFieldType = iota
	JumpFieldGrid
	JumpFieldPlates
)

// Fitting is an optional hull add-on.
type Fitting string

const (
	FlotationHull   Fitting = "FlotationHull"
	SubmergenceHull Fitting = "SubmergenceHull"
	Fins            Fitting = "Fins"
	FoldingFins     Fitting = "FoldingFins"
	Wings           Fitting = "Wings"
	FoldingWings    Fitting = "FoldingWings"
	LandingSkids    Fitting = "LandingSkids"
	LandingLegs     Fitting = "LandingLegs"
	LandingWheels   Fitting = "LandingWheels"
	Lifters         Fitting = "Lifters"
)

// ArmorLayer is one applied layer of hull armor.
type ArmorLayer struct {
	Type      string
	Value     int
	Coating   string
	AntiLayer bool
}

// Hull is a ship's structural airframe.
type Hull struct {
	ID            string // hull size code, e.g. "A".."Z", "N2".."Z2"
	Tons          int    // displacement tons, 100-2400
	Configuration Configuration
	Structure     Structure
	ArmorValue    int
	ArmorLayers   []ArmorLayer
	Fittings      []Fitting
	JumpFieldType JumpFieldType
	MaxG          int
}
