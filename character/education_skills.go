package character

// Book 1 p.60's own "AVAILABLE SKILLS" matrix — which skills and
// knowledges each kind of educational institution can teach.
//
// The page is four side-by-side skill columns, each carrying its own
// bank of school flags, and the pdftotext extract interleaves all four
// into single lines. This transcription was recovered from the PDF's own
// word bounding boxes rather than from the extract's visual layout,
// because the two disagree: extract line 4547 reads "Command College |
// Begin vs Edu or Tra | (may not transfer to Citizen)" as though the
// middle phrase belonged to Command College, when by coordinate those
// are three different page columns and "Begin vs Edu or Tra" is the
// Scholar career's own Begin check.
//
// Reading the flags needs p.61's legend, which defines A, N and M twice:
//
//	College   C  College or University, including Masters and Professors
//	          L  Law School
//	          M  Medical School
//	Academy   A  Military Academy and Military Command College
//	          N  Naval Academy and Naval Command College
//	School    A  Army School
//	          N  Naval School
//	          M  Marine School
//	          S  School including Apprentice, Training Course, Trade School
//
// The two M codes are told apart by which column the letter sits in, and
// the page's own layout is what makes that unambiguous. The two
// left-hand skill groups (Base Skills, and the Ship/Soldier/Arts/Trades
// block) have four flag columns headed College, then two Academy ones,
// then School; the two right-hand groups have five headed College, Army,
// Navy, Marine, Trade. So an M in a left group's first column is Medical
// School (Forensics, Medic) while an M in its third is Marine School
// (Leader, Recon, Sapper) — and the right-hand groups never print L or a
// Medical M at all, which is the cross-check that the column reading is
// right rather than merely self-consistent.
//
// One consequence worth recording because it looks like a transcription
// error and is not: the two left-hand groups have no A column whatever.
// No Base Skill and nothing in the Ship/Soldier/Arts/Trades block is
// available at the Military Academy; the Army's own skills all sit in
// the right-hand groups (Battle Dress, Blades, Slug Throw, Artillery,
// WMD, the ground vehicles). Verified against the bounding boxes twice.
type educationSkill struct {
	Name string
	// Parent is the p.60 specialty grouping a skill is printed under —
	// "Driver" over ACV/Automotive/Grav, "Seafarer" over Boat/Ship/Sub.
	// Recorded because three separate rows are printed "Grav" (under
	// Driver, Flyer and Seafarer) and nothing else distinguishes them.
	Parent  string
	Schools schoolSet
}

// schoolSet is the set of institutions that teach one skill — p.60's own
// flag columns, one bit per legend code.
type schoolSet uint8

const (
	schoolCollege         schoolSet = 1 << iota // C
	schoolLaw                                   // L
	schoolMedical                               // M, in a College column
	schoolMilitaryAcademy                       // A — Military Academy, and Army School
	schoolNavalAcademy                          // N — Naval Academy, and Naval School
	schoolMarine                                // M, in a School column — Marine School
	schoolTrade                                 // S
)

// educationSkills is p.60's matrix in printed order, group by group:
// Base Skills, then the Ship/Soldier/Arts/Trades block, then the two
// right-hand groups. The trailing comment on each row is the group and
// the flag string exactly as printed, so a row can be checked against
// the page without re-deriving the bitmask.
var educationSkills = []educationSkill{
	{"Admin", "", schoolTrade},                // G1 ...S
	{"Advocate", "", schoolLaw},               // G1 L...
	{"Animals", "", 0},                        // G1 ....
	{"Athlete", "", schoolCollege},            // G1 C...
	{"Broker", "", schoolCollege},             // G1 C...
	{"Bureaucrat", "", schoolCollege},         // G1 C...
	{"Comms", "", schoolTrade},                // G1 ...S
	{"Computer", "", schoolTrade},             // G1 ...S
	{"Counsellor", "", schoolCollege},         // G1 C...
	{"Designer", "", schoolCollege},           // G1 C...
	{"Diplomat", "", schoolLaw},               // G1 L...
	{"Driver", "", 0},                         // G1 ....
	{"Explosives", "", schoolTrade},           // G1 ...S
	{"Fleet Tactics", "", schoolNavalAcademy}, // G1 .N..
	{"Flyer", "", 0},                          // G1 ....
	{"Forensics", "", schoolMedical},          // G1 M...
	{"Gambler", "", 0},                        // G1 ....
	{"High-G", "", schoolTrade},               // G1 ...S
	{"Hostile Env", "", schoolTrade},          // G1 ...S
	{"JOT", "", 0},                            // G1 ....
	{
		"Language",
		"",
		schoolCollege | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G1 CNMS
	{
		"Leader",
		"",
		schoolNavalAcademy | schoolMarine,
	}, // G1 .NM.
	{
		"Liaison",
		"",
		schoolNavalAcademy | schoolMarine,
	}, // G1 .NM.
	{
		"Naval Arch",
		"",
		schoolNavalAcademy,
	}, // G1 .N..
	{
		"Seafarer",
		"",
		0,
	}, // G1 ....
	{
		"Stealth",
		"",
		0,
	}, // G1 ....
	{
		"Strategy",
		"",
		schoolNavalAcademy | schoolMarine,
	}, // G1 .NM.
	{
		"Streetwise",
		"",
		0,
	}, // G1 ....
	{
		"Survey",
		"",
		schoolTrade,
	}, // G1 ...S
	{
		"Survival",
		"",
		schoolTrade,
	}, // G1 ...S
	{
		"Tactics",
		"",
		schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G1 .NMS
	{
		"Teacher",
		"",
		schoolCollege,
	}, // G1 C...
	{
		"Trader",
		"",
		schoolTrade,
	}, // G1 ...S
	{
		"Vacc Suit",
		"",
		schoolTrade,
	}, // G1 ...S
	{
		"Zero-G",
		"",
		schoolTrade,
	}, // G1 ...S
	{
		"Astrogator",
		"",
		schoolCollege | schoolNavalAcademy,
	}, // G2 CN..
	{
		"Engineer",
		"",
		0,
	}, // G2 ....
	{
		"Gunner",
		"",
		schoolNavalAcademy,
	}, // G2 .N..
	{
		"Medic",
		"",
		schoolMedical | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G2 MNMS
	{
		"Pilot",
		"",
		0,
	}, // G2 ....
	{
		"Sensors",
		"",
		schoolNavalAcademy | schoolTrade,
	}, // G2 .N.S
	{
		"Steward",
		"",
		schoolNavalAcademy | schoolTrade,
	}, // G2 .N.S
	{
		"Fighter",
		"",
		schoolMarine,
	}, // G2 ..M.
	{
		"Fwd Obs",
		"",
		schoolMarine | schoolTrade,
	}, // G2 ..MS
	{
		"Hvy Wpns",
		"",
		schoolMarine,
	}, // G2 ..M.
	{
		"Navigation",
		"",
		schoolMarine | schoolTrade,
	}, // G2 ..MS
	{
		"Recon",
		"",
		schoolMarine | schoolTrade,
	}, // G2 ..MS
	{
		"Sapper",
		"",
		schoolMarine | schoolTrade,
	}, // G2 ..MS
	{
		"Actor",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Artist",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Author",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Chef",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Dancer",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Musician",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Biologics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Craftsman",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Electronics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Fluidics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Gravitics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Magnetics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Mechanic",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Photonics",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Polymers",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"Program",
		"",
		schoolCollege | schoolTrade,
	}, // G2 C..S
	{
		"ACV",
		"Driver",
		schoolMilitaryAcademy | schoolTrade,
	}, // G3 .A..S
	{
		"Automotive",
		"Driver",
		schoolCollege | schoolMilitaryAcademy | schoolTrade,
	}, // G3 CA..S
	{
		"Grav",
		"Driver",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G3 .ANMS
	{
		"Legged",
		"Driver",
		schoolMilitaryAcademy | schoolTrade,
	}, // G3 .A..S
	{
		"Mole",
		"Driver",
		schoolMilitaryAcademy | schoolTrade,
	}, // G3 .A..S
	{
		"Tracked",
		"Driver",
		schoolMilitaryAcademy | schoolMarine | schoolTrade,
	}, // G3 .A.MS
	{
		"Wheeled",
		"Driver",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G3 .ANMS
	{
		"Battle Dress",
		"Fighter",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine,
	}, // G3 .ANM.
	{
		"Beams",
		"Fighter",
		schoolMilitaryAcademy | schoolMarine,
	}, // G3 .A.M.
	{
		"Blades",
		"Fighter",
		schoolMilitaryAcademy | schoolMarine | schoolTrade,
	}, // G3 .A.MS
	{
		"Exotics",
		"Fighter",
		schoolMilitaryAcademy | schoolMarine,
	}, // G3 .A.M.
	{
		"Slug Throw",
		"Fighter",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G3 .ANMS
	{
		"Sprays",
		"Fighter",
		schoolMilitaryAcademy | schoolMarine,
	}, // G3 .A.M.
	{
		"Unarmed",
		"Fighter",
		schoolMilitaryAcademy | schoolMarine | schoolTrade,
	}, // G3 .A.MS
	{
		"J-Drives",
		"Engineer",
		schoolNavalAcademy | schoolTrade,
	}, // G3 ..N.S
	{
		"Life Support",
		"Engineer",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolTrade,
	}, // G3 .AN.S
	{
		"M-Drive",
		"Engineer",
		schoolNavalAcademy | schoolTrade,
	}, // G3 ..N.S
	{
		"P-Systems",
		"Engineer",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G3 .ANMS
	{
		"Archeology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Biology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Chemistry",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"History",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Linguistics",
		"Sciences",
		schoolCollege | schoolTrade,
	}, // G3 C...S
	{
		"Philosophy",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Physics",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Planetology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Psionicology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Psychohistory",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Psychology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Robotics",
		"Sciences",
		schoolCollege | schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G3 CANMS
	{
		"Sophontology",
		"Sciences",
		schoolCollege,
	}, // G3 C....
	{
		"Aeronautics",
		"Flyer",
		schoolCollege | schoolMilitaryAcademy | schoolNavalAcademy | schoolTrade,
	}, // G4 CAN.S
	{
		"Flapper",
		"Flyer",
		schoolMilitaryAcademy | schoolTrade,
	}, // G4 .A..S
	{
		"Grav",
		"Flyer",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G4 .ANMS
	{
		"LTA",
		"Flyer",
		schoolMilitaryAcademy | schoolTrade,
	}, // G4 .A..S
	{
		"Rotor",
		"Flyer",
		schoolMilitaryAcademy | schoolTrade,
	}, // G4 .A..S
	{
		"Winged",
		"Flyer",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolTrade,
	}, // G4 .AN.S
	{
		"Bay Wpns",
		"Gunner",
		schoolNavalAcademy,
	}, // G4 ..N..
	{
		"Ortillery",
		"Gunner",
		schoolNavalAcademy,
	}, // G4 ..N..
	{
		"Screens",
		"Gunner",
		schoolMilitaryAcademy | schoolNavalAcademy,
	}, // G4 .AN..
	{
		"Spines",
		"Gunner",
		schoolNavalAcademy,
	}, // G4 ..N..
	{
		"Turrets",
		"Gunner",
		schoolNavalAcademy | schoolMarine,
	}, // G4 ..NM.
	{
		"Small Craft",
		"Pilot",
		schoolMilitaryAcademy | schoolMarine | schoolTrade,
	}, // G4 .A.MS
	{
		"Pilot-ACS",
		"Pilot",
		schoolNavalAcademy,
	}, // G4 ..N..
	{
		"Pilot-BCS",
		"Pilot",
		schoolNavalAcademy,
	}, // G4 ..N..
	{
		"Rider",
		"Animals",
		schoolMilitaryAcademy | schoolTrade,
	}, // G4 .A..S
	{
		"Teamster",
		"Animals",
		schoolMilitaryAcademy | schoolTrade,
	}, // G4 .A..S
	{
		"Trainer",
		"Animals",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G4 .ANMS
	{
		"Aquanautics",
		"Seafarer",
		schoolCollege | schoolTrade,
	}, // G4 C...S
	{
		"Grav",
		"Seafarer",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine | schoolTrade,
	}, // G4 .ANMS
	{
		"Boat",
		"Seafarer",
		schoolMarine | schoolTrade,
	}, // G4 ...MS
	{
		"Ship",
		"Seafarer",
		schoolMarine | schoolTrade,
	}, // G4 ...MS
	{
		"Sub",
		"Seafarer",
		schoolMarine | schoolTrade,
	}, // G4 ...MS
	{
		"Artillery",
		"Hvy Wpns",
		schoolMilitaryAcademy | schoolMarine,
	}, // G4 .A.M.
	{
		"Launcher",
		"Hvy Wpns",
		schoolMilitaryAcademy | schoolMarine,
	}, // G4 .A.M.
	{
		"Ordnance",
		"Hvy Wpns",
		schoolMilitaryAcademy | schoolMarine,
	}, // G4 .A.M.
	{
		"WMD",
		"Hvy Wpns",
		schoolMilitaryAcademy | schoolNavalAcademy | schoolMarine,
	}, // G4 .ANM.
}

// isKnowledge reports whether an entry is a Knowledge rather than a
// plain skill, which Book 1 p.61 defines structurally rather than by a
// marking on p.60:
//
//	"Some skills (Animals, Driver, Engineer, Fighter, Flyer, Gunner,
//	Heavy Weapons, Pilot, Seafarer) include within them several
//	Knowledges. Education or Training can only impart the Knowledges;
//	the Skills themselves are not obtainable in Education or Training."
//
// and, on the stand-alone ones:
//
//	"Some Knowledges are subsets of a Skill ... other Knowledges are
//	stand-alone sciences (Archeology is ...)."
//
// So a Knowledge is either a specialty printed under one of those parent
// skills or one of the Sciences — exactly the entries this table records
// a Parent for. p.61 enumerates every parent's list, and those counts
// match this transcription's groups row for row: Animals 3, Driver 7,
// Engineer 4, Fighter 7, Flyer 6, Gunner 5, Heavy Weapons 4, Pilot 3,
// Seafarer 5, plus the 13 Sciences.
//
// The parents themselves are the entries p.60 sets in bold, per its own
// "Bold= Knowledge-Only skill." legend — which is recoverable after all,
// from the PDF's font subsets, though not from any text extract. Two
// things about that marking are worth recording:
//
//   - p.61's prose names nine parents, but its own table prints a tenth,
//     Musician, with Instrument and Other Instrument under it. Musician
//     is not treated as a parent here, following the prose.
//   - Three bold parents carry school flags on p.60 anyway — Gunner .N..,
//     Fighter ..M., Hvy Wpns ..M. — while Engineer and Pilot carry none.
//     Verified against the rendered page, so it is the book contradicting
//     itself rather than a transcription slip. p.61's rule is followed
//     over p.60's flags: it is a stated rule, where the flags are an
//     unexplained inconsistency in three rows out of nine.
func (s educationSkill) isKnowledge() bool {
	return s.Parent != ""
}

// skillsForSchool is every skill p.60 marks as available at one
// institution, deduplicated by name and in printed order.
//
// Deduplicated because "Grav" is printed three times — as a Driver
// specialty, a Flyer specialty and a Seafarer one — and all three carry
// the same flags in every column, so a pool that listed each row
// separately would draw Grav three times as often as anything else
// without the page saying so.
func skillsForSchool(school schoolSet, knowledgesOnly bool) []string {
	var (
		names []string
		seen  = make(map[string]bool, len(educationSkills))
	)

	for _, skill := range educationSkills {
		if skill.Schools&school == 0 || seen[skill.Name] {
			continue
		}

		if knowledgesOnly && !skill.isKnowledge() {
			continue
		}

		seen[skill.Name] = true

		names = append(names, skill.Name)
	}

	return names
}
