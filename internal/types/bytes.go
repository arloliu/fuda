package types

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// parseBytes parses a string representation of bytes into an int64.
// It supports both pure numbers and size suffixes (IEC and SI).
// Examples: "1024", "1KiB", "2.5MB", "1G".
func parseBytes(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}

	// 1. Try parsing as a raw integer first
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, nil
	}

	// 2. Parse unit string
	// Find where the number ends and unit starts
	unitStart := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		r := rune(s[i])
		if unicode.IsDigit(r) || r == '.' {
			break
		}
		unitStart = i
	}

	if unitStart == len(s) || unitStart == 0 {
		return 0, fmt.Errorf("invalid size format: %s", s)
	}

	numStr := s[:unitStart]
	unitStr := s[unitStart:]

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size string: %s", s)
	}

	multiplier, ok := lookupByteMultiplier(unitStr)
	if !ok {
		return 0, fmt.Errorf("unknown unit: %s", unitStr)
	}

	// Calculate bytes
	bytesVal := val * float64(multiplier)

	// Reject fractional bytes to avoid silent truncation
	if math.Mod(bytesVal, 1) != 0 {
		return 0, fmt.Errorf("fractional bytes not allowed: %s", s)
	}

	// Check for overflow (int64 limit)
	if bytesVal > float64(math.MaxInt64) || bytesVal < float64(math.MinInt64) {
		return 0, fmt.Errorf("value out of range for int64: %f", bytesVal)
	}

	return int64(bytesVal), nil
}

func lookupByteMultiplier(unit string) (int64, bool) {
	unit = strings.ToLower(unit)
	val, ok := byteMultipliers[unit]
	return val, ok
}

// ByteMultiplier exposes the normalized multiplier lookup for other packages.
func ByteMultiplier(unit string) (int64, bool) {
	return lookupByteMultiplier(unit)
}

var byteMultipliers = map[string]int64{
	// IEC (Binary)
	"kib": 1024,
	"mib": 1024 * 1024,
	"gib": 1024 * 1024 * 1024,
	"tib": 1024 * 1024 * 1024 * 1024,
	"pib": 1024 * 1024 * 1024 * 1024 * 1024,
	"eib": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,

	// SI (Decimal)
	"kb": 1000,
	"mb": 1000 * 1000,
	"gb": 1000 * 1000 * 1000,
	"tb": 1000 * 1000 * 1000 * 1000,
	"pb": 1000 * 1000 * 1000 * 1000 * 1000,
	"eb": 1000 * 1000 * 1000 * 1000 * 1000 * 1000,

	// Byte
	"b": 1,
}

func parseBytesUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("cannot assign negative value %s to uint", s)
	}

	// 1. Try raw integer parse first (full uint64 range)
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return v, nil
	}

	// 2. Parse unit string
	unitStart := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		r := rune(s[i])
		if unicode.IsDigit(r) || r == '.' || r == '+' || r == '-' {
			break
		}
		unitStart = i
	}

	if unitStart == len(s) || unitStart == 0 {
		return 0, fmt.Errorf("invalid size format: %s", s)
	}

	numStr := s[:unitStart]
	unitStr := s[unitStart:]

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size string: %s", s)
	}
	if val < 0 {
		return 0, fmt.Errorf("cannot assign negative value %s to uint", s)
	}

	multiplier, ok := lookupByteMultiplier(unitStr)
	if !ok {
		return 0, fmt.Errorf("unknown unit: %s", unitStr)
	}

	bytesVal := val * float64(multiplier)

	if math.Mod(bytesVal, 1) != 0 {
		return 0, fmt.Errorf("fractional bytes not allowed: %s", s)
	}

	if bytesVal > float64(math.MaxUint64) {
		return 0, fmt.Errorf("value out of range for uint64: %f", bytesVal)
	}

	return uint64(bytesVal), nil
}
