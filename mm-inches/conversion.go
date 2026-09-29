package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var unitMillimeters = map[string]float64{
	"millimeters":       1,
	"centimeters":       10,
	"meters":            1000,
	"decimal-inches":    25.4,
	"fractional-inches": 25.4,
	"feet":              304.8,
	"yards":             914.4,
}

func convertFrom(unit string, value float64, decimalPlaces, denominator int) (map[string]string, error) {
	factor, ok := unitMillimeters[unit]
	if !ok {
		return nil, fmt.Errorf("Choose a supported unit")
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, fmt.Errorf("Enter a finite value")
	}

	mm := value * factor
	inches := mm / 25.4
	return map[string]string{
		"millimeters":       formatDecimal(mm, decimalPlaces),
		"centimeters":       formatDecimal(mm/10, decimalPlaces),
		"meters":            formatDecimal(mm/1000, decimalPlaces),
		"decimal-inches":    formatDecimal(inches, decimalPlaces),
		"fractional-inches": formatImperialFraction(inches, denominator),
		"feet":              formatDecimal(inches/12, decimalPlaces),
		"yards":             formatDecimal(inches/36, decimalPlaces),
	}, nil
}

func formatDecimal(value float64, places int) string {
	if math.Abs(value) < 0.5/math.Pow10(places) {
		value = 0
	}
	return strconv.FormatFloat(value, 'f', places, 64)
}

func validDenominator(value int) bool {
	switch value {
	case 2, 4, 8, 16, 32, 64:
		return true
	default:
		return false
	}
}

func formatImperialFraction(inches float64, denominator int) string {
	negative := inches < 0
	ticks := int64(math.Round(math.Abs(inches) * float64(denominator)))
	whole := ticks / int64(denominator)
	numerator := ticks % int64(denominator)

	var result string
	switch {
	case numerator == 0:
		result = strconv.FormatInt(whole, 10)
	default:
		divisor := gcd(numerator, int64(denominator))
		numerator /= divisor
		reducedDenominator := int64(denominator) / divisor
		if whole == 0 {
			result = fmt.Sprintf("%d/%d", numerator, reducedDenominator)
		} else {
			result = fmt.Sprintf("%d %d/%d", whole, numerator, reducedDenominator)
		}
	}
	if negative && ticks != 0 {
		return "-" + result
	}
	return result
}

func parseImperialFraction(raw string) (float64, error) {
	value := strings.TrimSpace(raw)
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed, nil
	}

	negative := strings.HasPrefix(value, "-")
	value = strings.TrimSpace(strings.TrimPrefix(value, "-"))
	value = strings.Replace(value, "-", " ", 1)
	parts := strings.Fields(value)
	if len(parts) == 0 || len(parts) > 2 {
		return 0, fmt.Errorf("invalid fraction")
	}

	whole := 0.0
	fractionPart := parts[0]
	if len(parts) == 2 {
		var err error
		whole, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid whole number")
		}
		fractionPart = parts[1]
	}

	fractionPieces := strings.Split(fractionPart, "/")
	if len(fractionPieces) != 2 {
		return 0, fmt.Errorf("invalid fraction")
	}
	numerator, err := strconv.ParseFloat(fractionPieces[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numerator")
	}
	denominator, err := strconv.ParseFloat(fractionPieces[1], 64)
	if err != nil || denominator == 0 {
		return 0, fmt.Errorf("invalid denominator")
	}

	result := whole + numerator/denominator
	if negative {
		result = -result
	}
	return result, nil
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
