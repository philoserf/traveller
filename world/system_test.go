package world_test

import (
	"testing"

	"github.com/philoserf/traveller/world"
)

func TestStellarRoleString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role world.StellarRole
		want string
	}{
		{world.Primary, "Primary"},
		{world.Close, "Close"},
		{world.Near, "Near"},
		{world.Far, "Far"},
		{world.StellarRole(99), ""},
	}

	for _, c := range cases {
		if got := c.role.String(); got != c.want {
			t.Errorf("StellarRole(%d).String() = %q, want %q", c.role, got, c.want)
		}
	}
}
