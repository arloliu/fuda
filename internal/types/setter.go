package types

// Setter is an interface for setting default values.
// Only pointer receivers should implement this interface.
type Setter interface {
	// SetDefaults sets default values for the struct.
	SetDefaults()
}
