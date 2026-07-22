package starship

// Configuration is a hull's shape/streamlining, which drives friction,
// agility, max G, stability, and land-capability.
type Configuration byte

// Configuration values, roughly ordered from least to most
// atmosphere-capable: Cluster/BracedCluster/Planetoid/Unstreamlined
// generally can't fly in atmosphere; Streamlined through LiftingBody can.
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

// Structure values and their base Armor Value formula: Plate AV=TL,
// Shell/Polymer/Organic AV=TL/2, FeNi AV=20, Charged AV=TL*2.
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

// JumpFieldType values; Bubble is the standard/default field type.
const (
	JumpFieldBubble JumpFieldType = iota
	JumpFieldGrid
	JumpFieldPlates
)

// Fitting is an optional hull add-on.
type Fitting string

// Fitting values, grouped by function: Flotation/SubmergenceHull for water
// operations, Fins/Wings (plus folding variants) for aerodynamic lift,
// Landing* for touching down, Lifters for VTOL.
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
