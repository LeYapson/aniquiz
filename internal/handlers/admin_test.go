package handlers

import "testing"

func TestIsAdmin(t *testing.T) {
	t.Setenv("ADMIN_USERNAMES", "yatokishi, LeYapson")

	cases := []struct {
		username string
		want     bool
	}{
		{"yatokishi", true},
		{"LeYapson", true},
		{"YATOKISHI", true}, // insensible à la casse
		{"leyapson", true},
		{"  yatokishi  ", false}, // l'entrée utilisateur n'est pas trimée, seule l'allowlist l'est
		{"intrus", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsAdmin(c.username); got != c.want {
			t.Errorf("IsAdmin(%q) = %v, want %v", c.username, got, c.want)
		}
	}
}

func TestIsAdminEmptyEnv(t *testing.T) {
	t.Setenv("ADMIN_USERNAMES", "")
	if IsAdmin("yatokishi") {
		t.Error("sans ADMIN_USERNAMES, personne ne doit être admin")
	}
}
