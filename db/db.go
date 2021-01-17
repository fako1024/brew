package db

import (
	"time"
)

// DataPoint denotes a data point with specific timings
type DataPoint struct {
	TimeStamp time.Time
	Data      map[string]interface{}
	Tags      map[string]string
}

// DataPoints denotes a list of data points
type DataPoints []DataPoint

// DB is an generic DB interface, providing functionality to interact with a database
type DB interface {

	// EmitDataPoints creates data points and stores it in the underlying database
	EmitDataPoints(db, measurement string, data DataPoints) error

	// ModifyMeasurement allows to alter certain elements of a measurement
	ModifyMeasurement(db, measurement, selectTagName, selectTagValue, replaceTagName, replaceTagValue string, additionalData map[string]interface{}) error
}
