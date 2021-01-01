package main

import (
	"flag"
	"strings"
	"time"

	"github.com/fako1024/brew"
	"github.com/fako1024/brew/db"
	"github.com/fako1024/brew/db/influx"
	"github.com/fako1024/brew/maintenance"
	"github.com/sirupsen/logrus"
)

const timestampLayout = "2006-01-02T15:04:05"

type config struct {
	id       string
	shotType brew.ShotType

	maintenanceTS   time.Time
	maintenanceType maintenance.Type

	influxEndpoint string
	influxUser     string
	influxPassword string
}

func main() {

	var (
		cfg          config
		shotTypeStr  string
		timestampStr string
	)

	// Basic flags for InfluxDB communication
	flag.StringVar(&cfg.influxEndpoint, "influxEndpoint", "", "Endpoint for InfluxDB emissions")
	flag.StringVar(&cfg.influxUser, "influxUser", "root", "User for InfluxDB emissions")
	flag.StringVar(&cfg.influxPassword, "influxPassword", "root", "Password for InfluxDB emissions")

	// Flags to perform changes to existing brews
	flag.StringVar(&cfg.id, "id", "", "Brew ID to perform change on")
	flag.StringVar(&shotTypeStr, "shotType", "", "Shot type to set")

	// Flags to perform / set maintenance parameters
	flag.StringVar(&timestampStr, "maintenance-time", time.Now().Format(timestampLayout), "Timestamp at which a maintenance was performed")
	flag.StringVar(&cfg.maintenanceType, "maintenance-type", "", "Type of performed maintenance")

	flag.Parse()
	if cfg.influxEndpoint == "" {
		logrus.StandardLogger().Fatalf("No InfluxDB endpoint specified")
	}
	influxDB := influx.New(
		cfg.influxEndpoint,
		cfg.influxUser,
		cfg.influxPassword,
	)

	// Shot type change requested
	if cfg.id != "" {
		if shotTypeStr != "" {
			cfg.shotType = brew.ShotTypeFromString(shotTypeStr)
			if cfg.shotType == brew.UnknownShot {
				logrus.StandardLogger().Fatalf("Invalid shot type specified: %s", shotTypeStr)
			}
		} else {
			logrus.StandardLogger().Fatal("No action specified")
		}

		if cfg.shotType != brew.UnknownShot {
			if err := influxDB.ModifyMeasurement("brews", "brew", "id", cfg.id, "shot_type", cfg.shotType.String()); err != nil {
				logrus.StandardLogger().Fatalf("Failed to alter measurement: %s", err)
			}
			if err := influxDB.ModifyMeasurement("brews", "summary", "id", cfg.id, "shot_type", cfg.shotType.String()); err != nil {
				logrus.StandardLogger().Fatalf("Failed to alter measurement summary: %s", err)
			}
			logrus.StandardLogger().Infof("Successfully changed shot type for ID %s to %s", cfg.id, cfg.shotType)
		}
	}

	if cfg.maintenanceType != "" {

		// Attempt to parse the maintenance timestamp
		var err error
		if cfg.maintenanceTS, err = time.Parse(timestampLayout, timestampStr); err != nil {
			logrus.StandardLogger().Fatalf("Failed to parse time stamp for maintenance task: %s", err)
		}

		// Check if the maintenance type is supported
		if !maintenance.IsValidType(cfg.maintenanceType) {
			logrus.StandardLogger().Fatalf("Invalid maintenance type: %s (supported: %v)", cfg.maintenanceType, maintenance.AllTypes)
		}

		if err := influxDB.EmitDataPoints("brews", "maintenance", db.DataPoints{
			{
				TimeStamp: cfg.maintenanceTS,
				Data:      map[string]interface{}{"type": strings.Title(strings.Replace(cfg.maintenanceType, "_", " ", -1))},
				Tags:      map[string]string{"maintenance_type": cfg.maintenanceType},
			},
		}); err != nil {
			logrus.StandardLogger().Fatalf("Failed to add maintenance task: %s", err)
		}
	}
}
