package types

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/arloliu/fuda/internal/bytesize"
)

// ParseBytes parses a string representation of bytes into an int64.
// It supports both pure numbers and size suffixes (IEC and SI).
// Examples: "1024", "1KiB", "2.5MB", "1G".
func ParseBytes(s string) (int64, error) {
	return bytesize.Parse(s)
}

// ByteMultiplier exposes the normalized multiplier lookup for other packages.
func ByteMultiplier(unit string) (int64, bool) {
	return bytesize.LookupMultiplier(unit)
}

// ParseBytesUint parses a string representation of bytes into a uint64.
// It supports both pure numbers and size suffixes (IEC and SI).
// Uses big.Int arithmetic for precision with large values.
func ParseBytesUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("cannot assign negative value %s to uint", s)
	}

	// Try parsing as raw uint64 first (fast path for large raw numbers)
	if val, err := strconv.ParseUint(s, 10, 64); err == nil {
		return val, nil
	}

	// Use big.Int-based parsing for precision
	val, err := bytesize.Parse(s)
	if err != nil {
		// Try parsing as big.Int for values > MaxInt64
		numStr, unitStr, splitErr := splitNumberUnit(s)
		if splitErr != nil {
			return 0, err // Return original error
		}

		bigVal, ok := bytesize.ParseToBigInt(numStr, unitStr)
		if !ok {
			return 0, err
		}

		// Check uint64 range
		if bigVal.Sign() < 0 {
			return 0, fmt.Errorf("cannot assign negative value %s to uint", s)
		}
		if bigVal.Cmp(new(big.Int).SetUint64(^uint64(0))) > 0 {
			return 0, fmt.Errorf("value out of range for uint64: %s", s)
		}

		return bigVal.Uint64(), nil
	}

	if val < 0 {
		return 0, fmt.Errorf("cannot assign negative value %s to uint", s)
	}

	return uint64(val), nil
}

// splitNumberUnit splits a size string into number and unit parts.
func splitNumberUnit(s string) (numStr, unitStr string, err error) {
	unitStart := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] >= '0' && s[i] <= '9' || s[i] == '.' {
			break
		}
		unitStart = i
	}

	if unitStart == len(s) || unitStart == 0 {
		return "", "", fmt.Errorf("invalid size format: %s", s)
	}

	return s[:unitStart], s[unitStart:], nil
}
