package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/fako1024/brew/db"
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
	csvData, err := os.Open(cfg.csvFile)
	if err != nil {
		logger.Fatalf("failed to open CSV file: %s", err)
	}
	defer csvData.Close()

	// Parse the file
	r := csv.NewReader(csvData)

	// Iterate through the records
	for {

		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Fatalf("failed to read record from CSV file: %s", err)
		}
		if len(record) != 10 {
			logger.Fatalf("unexpected number of columns in CSV file")
		}

		// Field order: ID, shot_type, start, end, end_weight, unit, battery_level, beans_weight, grind_setting, start_ns
		// Example:
		// 84e1ffa1-07fa-4d25-9af7-0a50debe1921,double,1603958992000,1603959019000,55.66,g,0.57,16.0,0.208695652,1603958992245000000
		id, shotType, unit := record[0], record[1], record[5]
		start, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			logger.Fatalf("failed to parse `start` column from CSV file: %s", err)
		}
		end, err := strconv.ParseInt(record[3], 10, 64)
		if err != nil {
			logger.Fatalf("failed to parse `end` column from CSV file: %s", err)
		}
		endWeight, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			logger.Fatalf("failed to parse `end_weight` column from CSV file: %s", err)
		}
		batterylevel, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			logger.Fatalf("failed to parse `battery_level` column from CSV file: %s", err)
		}
		beansWeight, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			logger.Fatalf("failed to parse `beans_weight` column from CSV file: %s", err)
		}
		grindSetting, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			logger.Fatalf("failed to parse `grind_setting` column from CSV file: %s", err)
		}
		startNS, err := strconv.ParseInt(record[9], 10, 64)
		if err != nil {
			logger.Fatalf("failed to parse `start_ns` column from CSV file: %s", err)
		}

		// Generate tags
		tags := map[string]string{
			"id":        id,
			"shot_type": shotType,
		}

		// Emit the summary to the influxDB
		if err := influxDB.EmitDataPoints("brews", "summary", db.DataPoints{
			{
				TimeStamp: time.Unix(0, startNS),
				Tags:      tags,
				Data: map[string]interface{}{
					"start":         start,
					"end":           end,
					"end_weight":    endWeight,
					"unit":          unit,
					"battery_level": batterylevel,
					"beans_weight":  beansWeight,
					"grind_setting": grindSetting,
				},
			},
		}); err != nil {
			logger.Errorf("failed to emit brew summary to influxDB: %s", err)
		}
	}

}
