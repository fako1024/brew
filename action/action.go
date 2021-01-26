package action

// Category denotes a category of action performed on the coffee machine
type Category = string

const (

	// Generic denotes a generic action
	Generic = "generic"

	// Maintenance denotes a maintenance action
	Maintenance = "maintenance"
)

// Type denotes a type of action performed on the coffee machine
type Type = string

const (

	// Generic actions

	// NewCoffeePack denotes opening a new pack of coffee
	NewCoffeePack = "new_coffee_pack"

	// Maintenance actions

	// BackFlush denotes a simple back flush maintenance using
	// Puly Caff or a similar claner
	BackFlush = "back_flush"

	// DescaleBrewGroup denotes a descaling of the brew group using
	// Puly Cleaner or a similar descaling agent
	DescaleBrewGroup = "descale_brew_group"

	// DescalePressureReliefValve denotes a descaling of the pressure relief valve
	// for the boiler
	DescalePressureReliefValve = "descale_pressure_relief_valve"

	// DescaleExpansionValve denotes a descaling of the expansion valve
	// for the brew cycle
	DescaleExpansionValve = "descale_expansion_valve"

	// DescaleFull denotes a ful descaling of the whole machine,
	// including replacement of the respective seals
	DescaleFull = "descale_full"
)

var categories = map[Type]Category{
	NewCoffeePack:              Generic,
	BackFlush:                  Maintenance,
	DescaleBrewGroup:           Maintenance,
	DescalePressureReliefValve: Maintenance,
	DescaleExpansionValve:      Maintenance,
	DescaleFull:                Maintenance,
}

// Categories returns a list of all types and their categories
func Categories() map[Type]Category {
	return categories
}

// Categorize checks and returns if a type is valid (and its category)
func Categorize(t Type) (category Category, isValid bool) {
	category, isValid = categories[t]

	return
}
