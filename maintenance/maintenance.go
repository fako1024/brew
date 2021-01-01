package maintenance

// Type denotes a type of maintenance performed on the coffee machine
type Type = string

const (

	// BackFlush denotes a simple back flush maintenance using
	// Puly Caff or a similar claner
	BackFlush = "back_flush"

	// DescaleBrewGroup denotes a descaling of the brew group using
	// Puly Cleaner or a similar descaling agent
	DescaleBrewGroup = "descale_brew_group"

	// DescaleFull denotes a ful descaling of the whole machine,
	// including replacement of the respective seals
	DescaleFull = "descale_full"
)

// AllTypes provides a lookup table for all existing types
var AllTypes = map[Type]struct{}{
	BackFlush:        {},
	DescaleBrewGroup: {},
	DescaleFull:      {},
}

// IsValidType checks and returns if a type is valid
func IsValidType(t Type) bool {
	_, exists := AllTypes[t]

	return exists
}
