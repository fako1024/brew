package influx

import (
	"encoding/json"
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
func (d *DB) ModifyMeasurement(db, measurement, selectTagName, selectTagValue, replaceTagName, replaceTagValue string) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
	if err != nil {
		return fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, db, err)
	}
	defer c.Close()

	// Get the column types
	q := client.NewQueryWithParameters("SHOW FIELD KEYS ON $d FROM $m", db, "ms", client.Params{
		"d": client.Identifier(db),
		"m": client.Identifier(measurement),
	})
	response, err := c.Query(q)
	if err != nil || response.Error() != nil {
		return fmt.Errorf("Failed to query measurement: %s", err)
	}
	columnTypes := make(map[string]string)
	for _, result := range response.Results {
		for _, ser := range result.Series {
			for _, row := range ser.Values {
				columnTypes[row[0].(string)] = row[1].(string)
			}
		}
	}

	// Get the requested measuremet values
	q = client.NewQueryWithParameters("SELECT * FROM $m WHERE $tag_name = $tag_value", db, "ms", client.Params{
		"m":         client.Identifier(measurement),
		"tag_name":  client.Identifier(selectTagName),
		"tag_value": client.StringValue(selectTagValue),
	})
	response, err = c.Query(q)
	if err != nil || response.Error() != nil {
		return fmt.Errorf("Failed to query measurement: %s", err)
	}
	if len(response.Results) != 1 || len(response.Results[0].Series) != 1 {
		return fmt.Errorf("Unexpected number of results / series returned, want exactly one")
	}

	// Generate tags for updated data points
	tags := map[string]string{
		selectTagName:  selectTagValue,
		replaceTagName: replaceTagValue,
	}

	// Process existing data points
	var dataPoints DataPoints
	for _, result := range response.Results {
		for _, ser := range result.Series {

			for _, row := range ser.Values {

				var (
					data = make(map[string]interface{})
					ts   time.Time
				)

				// Process individual columns
				for i, col := range ser.Columns {

					// If the column is a tag, ignore it
					if _, tagExists := tags[col]; tagExists {
						continue
					}

					// If the column is the timestamp, parse and convert it
					if col == "time" {
						tsParse, err := row[i].(json.Number).Int64()
						if err != nil {
							return fmt.Errorf("Failed to convert timestamp: %s", err)
						}
						ts = time.Unix(0, tsParse*int64(time.Millisecond))
						data[col] = ts
					} else {

						// Check the column type and perform parsing
						switch columnTypes[col] {
						case "integer":
							if intValue, err := row[i].(json.Number).Int64(); err == nil {
								data[col] = intValue
							} else {
								return fmt.Errorf("Failed to parse integer value for column %s", col)
							}
						case "float":
							if floatValue, err := row[i].(json.Number).Float64(); err == nil {
								data[col] = floatValue
							} else {
								return fmt.Errorf("Failed to parse floating point value for column %s", col)
							}
						case "string":
							data[col] = row[i].(string)
						default:
							return fmt.Errorf("Failed to find valid type assertion for column %s", col)
						}
					}
				}

				// Append the new data point
				dataPoints = append(dataPoints, DataPoint{
					TimeStamp: ts,
					Data:      data,
					Tags:      tags,
				})
			}

			// Cross check generated data points
			if len(dataPoints) != len(ser.Values) {
				return fmt.Errorf("Unexpected length of data points, want %d, have %d", len(ser.Values), len(dataPoints))
			}
		}
	}

	//Drop existing measurement with the provided tag from the DB
	q = client.NewQueryWithParameters("DROP SERIES FROM $m WHERE $tag_name = $tag_value", db, "ms", client.Params{
		"m":         client.Identifier(measurement),
		"tag_name":  client.Identifier(selectTagName),
		"tag_value": client.StringValue(selectTagValue),
	})
	response, err = c.Query(q)
	if err != nil || response.Error() != nil {
		return err
	}

	// Insert new data points for the same measuremet / tag combination
	return d.EmitDataPoints(db, measurement, dataPoints)
}
