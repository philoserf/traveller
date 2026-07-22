package world

// TradeCode is a standard T5 trade classification abbreviation.
type TradeCode string

const (
	AsteroidBelt TradeCode = "As"
	Desert       TradeCode = "De"
	Fluid        TradeCode = "Fl"
	Garden       TradeCode = "Ga"
	Hellworld    TradeCode = "He"
	IceCapped    TradeCode = "Ic"
	Ocean        TradeCode = "Oc"
	Vacuum       TradeCode = "Va"
	WaterWorld   TradeCode = "Wa"
	Satellite    TradeCode = "Sa"
	Locked       TradeCode = "Lk"
	Dieback      TradeCode = "Di"

	Barren          TradeCode = "Ba"
	LowPopulation   TradeCode = "Lo"
	NonIndustrial   TradeCode = "Ni"
	PreHigh         TradeCode = "Ph"
	HighPopulation  TradeCode = "Hi"
	PreAgricultural TradeCode = "Pa"
	Agricultural    TradeCode = "Ag"
	NonAgricultural TradeCode = "Na"

	PrisonExileCamp TradeCode = "Px"
	PreIndustrial   TradeCode = "Pi"
	Industrial      TradeCode = "In"
	Poor            TradeCode = "Po"
	PreRich         TradeCode = "Pr"
	Rich            TradeCode = "Ri"
	Frozen          TradeCode = "Fr"

	Hot          TradeCode = "Ho"
	Cold         TradeCode = "Co"
	Tropic       TradeCode = "Tr"
	Tundra       TradeCode = "Tu"
	TwilightZone TradeCode = "Tz"

	Farming      TradeCode = "Fa"
	Mining       TradeCode = "Mi"
	MilitaryRule TradeCode = "Mr"
	PenalColony  TradeCode = "Pe"
	Reserve      TradeCode = "Re"

	SubsectorCapital TradeCode = "Cp"
	SectorCapital    TradeCode = "Cs"
	Capital          TradeCode = "Cx"
	Colony           TradeCode = "Cy"
	Forbidden        TradeCode = "Fo"

	Puzzle         TradeCode = "Pz"
	Dangerous      TradeCode = "Da"
	DataRepository TradeCode = "Ab"
	AncientSite    TradeCode = "An"
)
