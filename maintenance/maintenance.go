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
