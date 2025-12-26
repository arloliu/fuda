package fuda

import "github.com/arloliu/fuda/internal/types"

// Scanner is implemented by custom types to define string-to-value conversion.
// When processing the default tag, if the target type implements Scanner,
// its Scan method is called instead of the built-in conversion.
//
// Example:
//
//	type LogLevel int
//
//	func (l *LogLevel) Scan(src any) error {
//	    s, ok := src.(string)
//	    if !ok {
//	        return fmt.Errorf("expected string, got %T", src)
//	    }
//	    switch strings.ToLower(s) {
//	    case "debug":
//	        *l = 0
//	    case "info":
//	        *l = 1
//	    case "warn":
//	        *l = 2
//	    case "error":
//	        *l = 3
//	    default:
//	        return fmt.Errorf("unknown log level: %s", s)
//	    }
//	    return nil
//	}
type Scanner = types.Scanner
