package fuda

import "github.com/arloliu/fuda/internal/types"

// Setter is implemented by config structs that need dynamic defaults.
// SetDefaults is called after all tag processing (default, env, ref) completes.
//
// Example:
//
//	type Config struct {
//	    RequestID string
//	}
//
//	func (c *Config) SetDefaults() {
//	    if c.RequestID == "" {
//	        c.RequestID = uuid.New().String()
//	    }
//	}
type Setter = types.Setter
