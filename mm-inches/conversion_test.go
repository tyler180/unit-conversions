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

func TestFeetInches(t *testing.T) {
	cases := []struct {
		input     string
		inches    float64
		formatted string
	}{
		{`5' 6 1/2"`, 66.5, `5' 6 1/2"`},
		{`5 ft 6.5 in`, 66.5, `5' 6 1/2"`},
		{`-0' 6"`, -6, `-0' 6"`},
		{`0' 0"`, 0, `0' 0"`},
		{`2'`, 24, `2' 0"`},
		{`1' 11.99"`, 23.99, `2' 0"`},
	}
	for _, tt := range cases {
		got, err := parseValue("feet-inches", tt.input)
		if err != nil || math.Abs(got-tt.inches) > 1e-9 {
			t.Errorf("parseValue(%q) = %v, %v; want %v", tt.input, got, err, tt.inches)
			continue
		}
		values, err := convertFrom("feet-inches", got, 4, 16)
		if err != nil || values["feet-inches"] != tt.formatted {
			t.Errorf("convertFrom(%q) = %q, %v; want %q", tt.input, values["feet-inches"], err, tt.formatted)
		}
	}
	for _, input := range []string{`abc`, `5' -6"`, `5' 1/0"`, `5' NaN"`, `5' 6"junk`} {
		if _, err := parseValue("feet-inches", input); err == nil {
			t.Errorf("parseValue(%q) should fail", input)
		}
	}
}
