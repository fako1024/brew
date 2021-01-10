package influx

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fako1024/brew/db"
	client "github.com/influxdata/influxdb1-client/v2"
)

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
func (d *DB) EmitDataPoints(dbName, measurement string, data db.DataPoints) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
	if err != nil {
		return fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, dbName, err)
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "ms",
	})

	for _, v := range data {
		pt, err := client.NewPoint(measurement, v.Tags, v.Data, v.TimeStamp)
		if err != nil {
			return fmt.Errorf("Error creating InfluxDB Point for measurement %s on DB %s: %s", measurement, dbName, err)
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	if err = c.Write(bp); err != nil {
		return fmt.Errorf("Error writing InfluxDB Batch for measurement %s on DB %s: %s", measurement, dbName, err)
	}

	return nil
}

// FetchMeasurementsTable retrieves a measurement
func (d *DB) FetchMeasurementsTable(dbName, measurement string, field ...string) ([][]string, error) {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
	if err != nil {
		return nil, fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, dbName, err)
	}
	defer c.Close()

	for i := 0; i < len(field); i++ {
		field[i] = "\"" + field[i] + "\""
	}

	// Get the requested measuremet values
	q := client.NewQueryWithParameters(fmt.Sprintf("SELECT %s FROM $m", strings.Join(field, ",")), dbName, "ns", client.Params{
		"m": client.Identifier(measurement),
	})
	response, err := c.Query(q)
	if err != nil || response.Error() != nil {
		return nil, fmt.Errorf("Failed to query measurement: %s, %s", err, response.Error())
	}

	var entries [][]string

	for _, result := range response.Results {
		for _, ser := range result.Series {
			for _, row := range ser.Values {
				var rowFields []string
				for _, column := range row {
					if colStr, ok := column.(string); ok {
						rowFields = append(rowFields, colStr)
					} else if colStringer, ok := column.(fmt.Stringer); ok {
						rowFields = append(rowFields, colStringer.String())
					} else {
						return nil, fmt.Errorf("Failed to parse column value: %s", column)
					}
				}
				entries = append(entries, rowFields)
			}
		}
	}

	return entries, nil
}

// ModifyMeasurement allows to alter certain elements of a measurement
func (d *DB) ModifyMeasurement(dbName, measurement, selectTagName, selectTagValue, replaceTagName, replaceTagValue string) error {

	// Create a new InfluxDB client
	c, err := client.NewHTTPClient(*d.config)
	if err != nil {
		return fmt.Errorf("Error creating InfluxDB Client for measurement %s on DB %s: %s", measurement, dbName, err)
	}
	defer c.Close()

	// Get the column types
	q := client.NewQueryWithParameters("SHOW FIELD KEYS ON $d FROM $m", dbName, "ms", client.Params{
		"d": client.Identifier(dbName),
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

	// Get the requested measurement values
	q = client.NewQueryWithParameters("SELECT * FROM $m WHERE $tag_name = $tag_value", dbName, "ms", client.Params{
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
	var dataPoints db.DataPoints
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
				dataPoints = append(dataPoints, db.DataPoint{
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
	q = client.NewQueryWithParameters("DROP SERIES FROM $m WHERE $tag_name = $tag_value", dbName, "ms", client.Params{
		"m":         client.Identifier(measurement),
		"tag_name":  client.Identifier(selectTagName),
		"tag_value": client.StringValue(selectTagValue),
	})
	response, err = c.Query(q)
	if err != nil || response.Error() != nil {
		return err
	}

	// Insert new data points for the same measuremet / tag combination
	return d.EmitDataPoints(dbName, measurement, dataPoints)
}
