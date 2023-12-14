package main

import (
	"encoding/csv"
	"flag"
	"os"

	"github.com/fako1024/brew/db/influx"
	"github.com/fako1024/btscale/pkg/scale"
)

const timestampLayout = "2006-01-02T15:04:05"

type config struct {
	id string

	csvFile string

	influxEndpoint string
	influxUser     string
	influxPassword string
}

func main() {

	var (
		cfg config
	)

	// Basic flags for InfluxDB communication
	flag.StringVar(&cfg.csvFile, "csv", "", "Path to CSV file")
	flag.StringVar(&cfg.influxEndpoint, "influxEndpoint", "", "Endpoint for InfluxDB emissions")
	flag.StringVar(&cfg.influxUser, "influxUser", "root", "User for InfluxDB emissions")
	flag.StringVar(&cfg.influxPassword, "influxPassword", "root", "Password for InfluxDB emissions")

	flag.Parse()
	logger := scale.NewDefaultLogger(false)
	if cfg.influxEndpoint == "" {
		logger.Fatalf("no InfluxDB endpoint specified")
	}
	influxDB := influx.New(
		cfg.influxEndpoint,
		cfg.influxUser,
		cfg.influxPassword,
	)

	// Open the file
	csvData, err := os.OpenFile(cfg.csvFile, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		logger.Fatalf("failed to open CSV file: %s", err)
	}
	defer csvData.Close()

	// Retrieve the measurements
	rows, err := influxDB.FetchMeasurementsTable("brews", "summary", "id", "shot_type", "start", "end", "end_weight", "unit", "battery_level", "beans_weight", "grind_setting")
	if err != nil {
		logger.Fatalf("failed to perform query: %s", err)
	}

	w := csv.NewWriter(csvData)

	// Iterate through the records
	for _, row := range rows {

		if len(row) != 10 {
			logger.Fatalf("unexpected number of columns in measurement list")
		}

		// As the query returns the timestamp first, move it to end of csv
		var x string
		x, row = row[0], row[1:]
		row = append(row, x)

		if err := w.Write(row); err != nil {
			logger.Fatalf("failed to write record %v: %s", row, err)
		}
	}

	w.Flush()
}
