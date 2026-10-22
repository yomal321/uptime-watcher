package main

import "testing"

func TestValidURL(t *testing.T) {
	cases := map[string]bool{
		"https://example.com":      true,
		"http://localhost:3000/ok": true,
		"ftp://example.com":        false,
		"example.com":              false,
		"https://":                 false,
		"":                         false,
	}
	for in, want := range cases {
		if got := validURL(in); got != want {
			t.Errorf("validURL(%q) = %v, want %v", in, got, want)
		}
	}
}
