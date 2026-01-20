package types

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Scanner is an interface for custom string-to-type conversion.
type Scanner interface {
	// Scan assigns the value to the custom type.
	Scan(src any) error
}

// Convert converts a string value to the target reflect.Value's type.
func Convert(value string, target reflect.Value) error {
	if !target.CanSet() {
		return nil
	}

	// Handle custom types that implement Scanner
	if target.CanAddr() {
		if scanner, ok := target.Addr().Interface().(Scanner); ok {
			return scanner.Scan(value)
		}
	}

	//nolint:exhaustive // Only common types need explicit handling
	switch target.Kind() {
	case reflect.String:
		target.SetString(value)
	case reflect.Bool:
		return convertBool(value, target)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertInt(value, target)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertUint(value, target)
	case reflect.Float32, reflect.Float64:
		return convertFloat(value, target)
	case reflect.Slice:
		return convertSlice(value, target)
	case reflect.Map:
		return convertMap(value, target)
	case reflect.Struct:
		return convertStruct(value, target)
	case reflect.Pointer:
		return convertPointer(value, target)
	default:
		return fmt.Errorf("unsupported type: %s", target.Kind())
	}

	return nil
}

func convertBool(value string, target reflect.Value) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	target.SetBool(v)

	return nil
}

func convertInt(value string, target reflect.Value) error {
	// Special handling for Duration
	if target.Type() == reflect.TypeFor[time.Duration]() {
		d, err := parseDuration(value)
		if err != nil {
			return err
		}
		target.SetInt(int64(d))

		return nil
	}

	// 1. Try generic byte parsing (handles raw numbers and size strings)
	v, err := ParseBytes(value)
	if err != nil {
		return err
	}

	// 2. Check for overflow based on target bit size
	if target.OverflowInt(v) {
		return fmt.Errorf("value %s overflows %s", value, target.Type())
	}

	target.SetInt(v)

	return nil
}

// parseDuration extends time.ParseDuration to support days with 'd' suffix.
// Examples: "5d" -> 5 days, "1d12h" -> 1 day and 12 hours, "2d30m" -> 2 days and 30 minutes.
func parseDuration(s string) (time.Duration, error) {
	// Find and convert 'd' suffix for days to hours
	// We need to handle cases like "5d", "1d12h", "2d30m5s"
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		// Find the start of a number
		numStart := i
		for i < len(s) && (s[i] == '-' || s[i] == '+' || s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
			i++
		}
		if i == numStart {
			// No number found, just copy the character
			if i < len(s) {
				result.WriteByte(s[i])
				i++
			}
			continue
		}
		numStr := s[numStart:i]

		// Find the unit
		unitStart := i
		for i < len(s) && ((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z')) {
			i++
		}
		unit := s[unitStart:i]

		// Convert 'd' or 'D' to hours
		if unit == "d" || unit == "D" {
			// Parse the number and multiply by 24
			days, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid duration: %s", s)
			}
			hours := days * 24
			result.WriteString(strconv.FormatFloat(hours, 'f', -1, 64))
			result.WriteString("h")
		} else {
			result.WriteString(numStr)
			result.WriteString(unit)
		}
	}

	return time.ParseDuration(result.String())
}

func convertUint(value string, target reflect.Value) error {
	v, err := ParseBytesUint(value)
	if err != nil {
		return err
	}

	// Check for overflow
	if target.OverflowUint(v) {
		return fmt.Errorf("value %s overflows %s", value, target.Type())
	}

	target.SetUint(v)

	return nil
}

func convertFloat(value string, target reflect.Value) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	target.SetFloat(v)

	return nil
}

func convertSlice(value string, target reflect.Value) error {
	// Special case: []byte should receive raw bytes, not CSV-parsed
	if target.Type().Elem().Kind() == reflect.Uint8 {
		target.SetBytes([]byte(value))
		return nil
	}

	reader := csv.NewReader(strings.NewReader(value))
	reader.TrimLeadingSpace = true
	parts, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to parse csv slice: %w", err)
	}

	slice := reflect.MakeSlice(target.Type(), len(parts), len(parts))
	for i, part := range parts {
		if err := Convert(part, slice.Index(i)); err != nil {
			return err
		}
	}
	target.Set(slice)

	return nil
}

func convertMap(value string, target reflect.Value) error {
	// format: key:value,key2:value2 (supports quoting via CSV)
	reader := csv.NewReader(strings.NewReader(value))
	reader.TrimLeadingSpace = true
	parts, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to parse csv map: %w", err)
	}

	resultMap := reflect.MakeMap(target.Type())
	keyType := target.Type().Key()
	elemType := target.Type().Elem()

	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid map item format: %s", part)
		}
		keyStr := strings.TrimSpace(kv[0])
		valStr := strings.TrimSpace(kv[1])

		keyVal := reflect.New(keyType).Elem()
		if err := Convert(keyStr, keyVal); err != nil {
			return err
		}

		elemVal := reflect.New(elemType).Elem()
		if err := Convert(valStr, elemVal); err != nil {
			return err
		}

		resultMap.SetMapIndex(keyVal, elemVal)
	}
	target.Set(resultMap)

	return nil
}

func convertStruct(value string, target reflect.Value) error {
	// Attempt JSON unmarshal if value looks like JSON object
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		if err := json.Unmarshal([]byte(value), target.Addr().Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal json to struct: %w", err)
		}

		return nil
	}

	return fmt.Errorf("unsupported conversion to struct for value: %s", value)
}

func convertPointer(value string, target reflect.Value) error {
	if target.IsNil() {
		target.Set(reflect.New(target.Type().Elem()))
	}

	return Convert(value, target.Elem())
}
