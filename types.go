package brew

import (
	"time"

	"github.com/fako1024/btscale/pkg/scale"
)

// ShotType denotes the type of brew (single or double)
type ShotType int

const (

	// UnknownShot denotes an invalid / unknown shot type
	UnknownShot ShotType = iota

	// SingleShot denotes a single shot brew
	SingleShot

	// DoubleShot denotes double shot brew
	DoubleShot
)

// String returns a string representation of the shot type
func (t ShotType) String() string {
	switch t {
	case SingleShot:
		return "single"
	case DoubleShot:
		return "double"
	default:
		return "unknown"
	}
}

// Brew denotes a brew process
type Brew struct {
	ID         string           // ID of brew
	Start      time.Time        // Start of the brewing process
	End        time.Time        // End of the brewing process
	ShotType   ShotType         // Type of brew (single / double / unknown)
	DataPoints scale.DataPoints // Data points collected as part of the brewing process
}
