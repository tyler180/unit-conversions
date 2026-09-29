package main

import (
	"math"
	"testing"
)

func TestConvertMillimeters(t *testing.T) {
	values, err := convertFrom("millimeters", 426, 4, 64)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"millimeters":       "426.0000",
		"centimeters":       "42.6000",
		"meters":            "0.4260",
		"decimal-inches":    "16.7717",
		"fractional-inches": "16 49/64",
		"feet":              "1.3976",
		"yards":             "0.4659",
	}
	for key, expected := range want {
		if values[key] != expected {
			t.Errorf("%s = %q; want %q", key, values[key], expected)
		}
	}
}

func TestFractionAccuracyAndReduction(t *testing.T) {
	tests := []struct {
		inches      float64
		denominator int
		want        string
	}{
		{16.7717, 4, "16 3/4"},
		{16.7717, 64, "16 49/64"},
		{2.5, 64, "2 1/2"},
		{0.999, 64, "1"},
		{-0.5, 16, "-1/2"},
	}
	for _, test := range tests {
		if got := formatImperialFraction(test.inches, test.denominator); got != test.want {
			t.Errorf("formatImperialFraction(%v, %d) = %q; want %q", test.inches, test.denominator, got, test.want)
		}
	}
}

func TestParseImperialFraction(t *testing.T) {
	tests := map[string]float64{
		"16 3/4": 16.75,
		"16-3/4": 16.75,
		"3/8":    0.375,
		"-1 1/2": -1.5,
		"2.625":  2.625,
	}
	for input, want := range tests {
		got, err := parseImperialFraction(input)
		if err != nil {
			t.Errorf("parseImperialFraction(%q): %v", input, err)
			continue
		}
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("parseImperialFraction(%q) = %v; want %v", input, got, want)
		}
	}
}
