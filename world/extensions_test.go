package world_test

import (
	"testing"

	"github.com/philoserf/traveller/world"
)

func TestTravelZoneString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		zone world.TravelZone
		want string
	}{
		{world.ZoneGreen, "Green"},
		{world.ZoneAmber, "Amber"},
		{world.ZoneRed, "Red"},
		{world.TravelZone(0), ""}, // zero value: world.Generate doesn't set this yet
		{world.TravelZone('X'), ""},
	}

	for _, c := range cases {
		if got := c.zone.String(); got != c.want {
			t.Errorf("TravelZone(%q).String() = %q, want %q", byte(c.zone), got, c.want)
		}
	}
}
