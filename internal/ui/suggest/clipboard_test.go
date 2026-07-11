package suggest

import "testing"

func TestLooksLikeURL(t *testing.T) {
	cases := map[string]bool{
		"https://youtu.be/abc123":        true,
		"http://example.com/track":       true,
		"just some search text":          false,
		"":                               false,
		"  https://youtu.be/abc123  ":    true,
	}
	for in, want := range cases {
		if got := LooksLikeURL(in); got != want {
			t.Errorf("LooksLikeURL(%q) = %v, want %v", in, got, want)
		}
	}
}
