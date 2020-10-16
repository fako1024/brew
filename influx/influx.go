package influx

import (
	"fmt"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

// DataPoint denotes a data point with specific timings
type DataPoint struct {
	TimeStamp time.Time
	Data      map[string]interface{}
	Tags      map[string]string
}

// DataPoints denotes a list of data points
type DataPoints []DataPoint

// EventTracker is an event tracker for card and selfieid which writes to influxdb.
type EventTracker struct {
	config *client.HTTPConfig
}

// NewEventTracker creates a new event tracker which can transmit events to influxdb.
func NewEventTracker(endpointURL, username, password string) *EventTracker {
	return &EventTracker{
		config: &client.HTTPConfig{
			Addr:     endpointURL,
			Username: username,
			Password: password,
		},
	}
}

// EmitDataPoints creates data points and stores it in the underlying Influx database
func (t *EventTracker) EmitDataPoints(db, measurement string, data DataPoints) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*t.config)
	if err != nil {
		return fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, db, err)
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: "ms",
	})

	for _, v := range data {
		pt, err := client.NewPoint(measurement, v.Tags, v.Data, v.TimeStamp)
		if err != nil {
			return fmt.Errorf("Error creating InfluxDB Point for measurement %s on DB %s: %s", measurement, db, err)
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	if err = c.Write(bp); err != nil {
		return fmt.Errorf("Error writing InfluxDB Batch for measurement %s on DB %s: %s", measurement, db, err)
	}

	return nil
}
