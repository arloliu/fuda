// Package bytesize provides parsing utilities for human-readable byte size strings.
package bytesize

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

// Multipliers maps unit suffixes to their byte multipliers.
// Keys are lowercase for case-insensitive lookup.
var Multipliers = map[string]int64{
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

// LookupMultiplier returns the byte multiplier for a unit suffix (case-insensitive).
func LookupMultiplier(unit string) (int64, bool) {
	val, ok := Multipliers[strings.ToLower(unit)]
	return val, ok
}

// Parse parses a string representation of bytes into an int64.
// It supports both pure numbers and size suffixes (IEC and SI).
// Uses big.Int arithmetic for precision with large values.
// Examples: "1024", "1KiB", "2.5MB", "1G".
func Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}

	// 1. Try parsing as a raw integer first (fast path)
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, nil
	}

	// 2. Split number and unit
	numStr, unitStr, err := splitNumberUnit(s)
	if err != nil {
		return 0, err
	}

	// 3. Parse using big.Int for precision
	result, ok := ParseToBigInt(numStr, unitStr)
	if !ok {
		return 0, fmt.Errorf("invalid size string: %s", s)
	}

	// 4. Check int64 range
	if !result.IsInt64() {
		return 0, fmt.Errorf("value out of range for int64: %s", s)
	}

	return result.Int64(), nil
}

// ParseToBigInt converts number+unit to bytes as big.Int.
// Using big.Int/big.Float to safely handle large numbers.
// Returns (nil, false) if parsing fails or result has fractional bytes.
func ParseToBigInt(numStr, unitStr string) (*big.Int, bool) {
	// Parse number with arbitrary precision
	val, _, err := big.ParseFloat(numStr, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, false
	}

	// Get multiplier
	multiplier, ok := LookupMultiplier(unitStr)
	if !ok {
		return nil, false
	}

	// Multiply
	mult := new(big.Float).SetInt64(multiplier)
	val.Mul(val, mult)

	// Convert to Int, reject fractional bytes
	i, _ := val.Int(nil)
	if val.Cmp(new(big.Float).SetInt(i)) != 0 {
		return nil, false
	}

	return i, true
}

// splitNumberUnit splits a size string into number and unit parts.
func splitNumberUnit(s string) (numStr, unitStr string, err error) {
	unitStart := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		r := rune(s[i])
		if unicode.IsDigit(r) || r == '.' {
			break
		}
		unitStart = i
	}

	if unitStart == len(s) || unitStart == 0 {
		return "", "", fmt.Errorf("invalid size format: %s", s)
	}

	return s[:unitStart], s[unitStart:], nil
}
