package game

import "testing"

func TestBuzzerPoints(t *testing.T) {
	cases := []struct {
		name      string
		placement int
		buzzMs    int64
		want      int
	}{
		{"1er + buzz très rapide", 0, 1500, 5 + 3},    // placement 5, vitesse +3
		{"2e + buzz rapide", 1, 4000, 3 + 2},          // placement 3, vitesse +2
		{"3e + buzz moyen", 2, 8000, 2 + 1},           // placement 2, vitesse +1
		{"4e (hors podium) + buzz lent", 3, 12000, 1}, // placement 1, vitesse 0
		{"1er mais buzz lent", 0, 15000, 5},           // placement 5, vitesse 0
		{"limite vitesse 3s exclue", 0, 3000, 5 + 2},  // 3000 n'est pas < 3000
		{"limite vitesse 6s exclue", 1, 6000, 3 + 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buzzerPoints(c.placement, c.buzzMs); got != c.want {
				t.Errorf("buzzerPoints(%d, %d) = %d, want %d", c.placement, c.buzzMs, got, c.want)
			}
		})
	}
}
