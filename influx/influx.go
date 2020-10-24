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

// DB is an InfluxDB interface, providing functionality to interact with the database
type DB struct {
	config *client.HTTPConfig
}

// New creates a new InfluxDB instance
func New(addr, username, password string) *DB {
	return &DB{
		config: &client.HTTPConfig{
			Addr:     addr,
			Username: username,
			Password: password,
		},
	}
}

// EmitDataPoints creates data points and stores it in the underlying Influx database
func (d *DB) EmitDataPoints(db, measurement string, data DataPoints) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
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

// ModifyMeasurement allows to alter certain elements of a measurement
func (d *DB) ModifyMeasurement(db, measurement, tag string) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
	if err != nil {
		return fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, db, err)
	}
	defer c.Close()

	q := client.NewQueryWithParameters("SELECT * FROM ", db, "ns", map[string]interface{}{
		"M": measurement,
	})
	response, err := c.Query(q)
	if err != nil || response.Error() != nil {
		return err
	}

	fmt.Println(response.Results)

	return nil
}
