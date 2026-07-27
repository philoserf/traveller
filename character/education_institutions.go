package character

import (
	"strings"

	"github.com/philoserf/traveller/dice"
)

// Book 1 p.92-93's own "Educational Institutions" chart, one of the
// Background Tables:
//
//	p.61: "Characters who attend schools, colleges, and universities can
//	determine the specific name of the school attended from the
//	Educational Institution Chart. The information adds to the
//	character's background; there may come a time when a character will
//	meet someone who also attended that school. School Rank. The chart
//	allows determination of the relative rank of schools when compared
//	with others."
//
// and p.72's own checklist step, "For Each School Attended: Note School
// Name and Rank".
//
// Twelve sub-tables, each a 1D name roll and a rank die that varies by
// institution. Transcribed from the PDF's word bounding boxes: the page
// prints two sub-tables side by side, so a visual read splices the
// left-hand table's entries into the right-hand one's.
//
// Every name is recorded verbatim, placeholders included, because most
// of them cannot be filled in yet — see resolveInstitutionName.
type educationInstitutionChart struct {
	// Names are the 1D entries in printed order.
	Names [6]string
	// RankDice is how many D6 the Rank column rolls. Zero is ED5's own
	// "Rank= Inconsequential", which prints no rank at all.
	RankDice int
}

// educationInstitutionCharts is the chart keyed by the institution names
// this codebase uses. Sub-tables 05-09 (Medical School, Law School,
// Naval Academy, Military Academy, Flight School) are transcribed too,
// even though nothing attends them yet, since the whole chart was read
// in one pass and leaving half of it out would mean reading the page
// again when #113 lands.
var educationInstitutionCharts = map[string]educationInstitutionChart{
	"ED5": {
		RankDice: 0, // "Rank= Inconsequential"
		Names: [6]string{
			"<City> School District",
			"<City> Independent School System",
			"<City> School District",
			"<Province> Academy",
			"<Province> Alternative Schools",
			"<Province> Education Company",
		},
	},
	"Trade School": {
		RankDice: 1,
		Names: [6]string{
			"Acme Trade School",
			"Institute of <Skill>",
			"<Company> School of <Skill>",
			"Standardized <Skill> Qualification Program",
			"<Company> Institute of <Skill>",
			"Certified <Skill> Course",
		},
	},
	"College": {
		RankDice: 2,
		Names: [6]string{
			"<City> College",
			"College of <World>",
			"<World> College",
			"<Province> Provincial College",
			"Imperial College of <World>",
			"Peoples College of <World>",
		},
	},
	"University": {
		RankDice: 3,
		Names: [6]string{
			"<World> University",
			"University of <World>",
			"Imperial University of <World>",
			"<World> Orbital University",
			"The <Color> Institute",
			"All-<World> University",
		},
	},
	"Medical School": {
		RankDice: 2,
		Names: [6]string{
			"<World> University Medical School",
			"University of <World> Medical School",
			"Imperial University of <World> School of Medicine",
			"<World> Orbital Medical University",
			"The <Color> Institute of Medicine",
			"All-<World> University Medical School",
		},
	},
	"Law School": {
		RankDice: 2,
		Names: [6]string{
			"<World> University Law School",
			"University of <World> School of Law",
			"Imperial University of <World> School of Advocacy",
			"<World> Orbital Legal University",
			"The <Color> Institute of Law",
			"All-<World> University Law School",
		},
	},
	"Naval Academy": {
		RankDice: 2,
		Names: [6]string{
			"<Sector> Naval Academy",
			"<Subsector> Naval Academy",
			"<World> Naval Academy",
			"<World> Continental Naval Academy",
			"<World> Continental Naval Academy",
			"<World> Reserve Naval Academy",
		},
	},
	"Military Academy": {
		RankDice: 2,
		Names: [6]string{
			"<Sector> Military Academy",
			"<Subsector> Military Academy",
			"<World> Military Academy",
			"<World> Continental Military Academy",
			"<World> Continental Military Academy",
			"<World> Reserve Military Academy",
		},
	},
	"Flight School": {
		RankDice: 2,
		Names: [6]string{
			"Imperial Navy Flight School (<Sector>)",
			"Imperial Navy Flight School (<Subsector>)",
			"Imperial Navy Flight School (<World>)",
			"Advanced Mission Flight School (<World>)",
			"Special Mission Flight Program (<Sector>)",
			"Imperial Navy Combat Flight School (<Subsector>)",
		},
	},
	"Command College": {
		RankDice: 2,
		Names: [6]string{
			"Imperial Command and General Staff College",
			"Imperial College of Military Command",
			"Imperial College of Naval Strategy",
			"Imperial Command College (<Armed Force>)",
			"Imperial War College",
			"Imperial College of Military Studies",
		},
	},
	"Training Course": {
		RankDice: 1,
		Names: [6]string{
			"<City> <Skill> Training Course",
			"<City> Education Company <Skill> Course",
			"<Color> Training System (<Skill>)",
			"<Province> Academy <Skill> Training",
			"<Province> School of <Skill>",
			"<Province> Education Company <Skill> Course",
		},
	},
	"ANM School": {
		RankDice: 1,
		Names: [6]string{
			"Subsector <Armed Force> <Skill> Course",
			"Imperial <Armed Force> <Skill> School",
			"Imperial <Armed Force> <Skill> Course",
			"Sector <Armed Force> <Skill> Course",
			"Imperial <Armed Force> <Skill> Training Course",
			"Imperial <Skill> Course",
		},
	},
}

// Placeholders the chart's names use.
const (
	placeholderSkill      = "<Skill>"
	placeholderArmedForce = "<Armed Force>"
)

// resolveInstitutionName fills in a chart name's placeholders, and
// returns false when any of them has no source in this codebase.
//
// Two of the seven can be filled in: <Skill> from whatever the school
// taught, and <Armed Force> from the service a character serves in. The
// other five — <World>, <Sector>, <Subsector>, <City>, <Province>,
// <Company>, <Color> — name places and organisations that do not exist
// here. world.Generate leaves world.World.Name zero-valued, sector names
// are supplied by the caller rather than generated, and there is no
// source at all for a city, province, company or colour.
//
// So a name is returned only when it comes out complete. Emitting
// "College of <World>" onto a character sheet would be worse than
// emitting nothing, and inventing a world name here would put
// name generation in the character package, where it does not belong.
// The roll that chose the template is kept either way (Education's own
// SchoolNameRoll), so filling these in later costs no dice and shifts no
// stream — see the follow-up issue this PR files.
func resolveInstitutionName(template, skill, armedForce string) (string, bool) {
	name := template

	if skill != "" {
		name = strings.ReplaceAll(name, placeholderSkill, skill)
	}

	if armedForce != "" {
		name = strings.ReplaceAll(name, placeholderArmedForce, armedForce)
	}

	if strings.ContainsRune(name, '<') {
		return "", false
	}

	return name, true
}

// rollInstitution is p.72's own "For Each School Attended: Note School
// Name and Rank" — the 1D name roll and the institution's own rank die,
// in that printed order.
//
// Both are drawn whenever a school is attended, including when the name
// they select cannot be rendered yet, so that the dice cost of this
// chart is paid once and does not move again when the remaining
// placeholders acquire sources.
//
// Only CharGen step C calls this so far. Command College and ANM School
// have chart entries too — and are the only two institutions whose names
// resolve completely today — but recording theirs means carrying a name
// on a Term rather than on Education, which is left with the follow-up.
func rollInstitution(r *dice.Roller, school, skill string) (int, string, int) {
	chart, ok := educationInstitutionCharts[school]
	if !ok {
		return 0, "", 0
	}

	roll := r.D6()

	// armedForce is empty here: no institution step C attends uses
	// <Armed Force>. It arrives with Command College and ANM School,
	// which are the two whose names resolve completely.
	name, _ := resolveInstitutionName(chart.Names[roll-1], skill, "")

	rank := 0
	if chart.RankDice > 0 {
		rank = r.SumD6(chart.RankDice)
	}

	return roll, name, rank
}
