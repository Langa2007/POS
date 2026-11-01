package api

import "testing"

func TestPermissionMatches(t *testing.T) {
	cases := []struct {
		granted, required string
		want              bool
	}{
		{"sales.read", "sales.read", true}, {"sales.read", "sales.write", false},
		{"core.*", "core.manifest", true}, {"core.*", "core.sales.write", true},
		{"users.*", "user.manage", false}, {"*", "anything", true},
	}
	for _, tc := range cases {
		if got := permissionMatches(tc.granted, tc.required); got != tc.want {
			t.Errorf("permissionMatches(%q,%q)=%t want %t", tc.granted, tc.required, got, tc.want)
		}
	}
}
