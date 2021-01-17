package main

import (
	"flag"
	"strings"
	"time"

	"github.com/fako1024/brew"
	"github.com/fako1024/brew/action"
	"github.com/fako1024/brew/db"
	"github.com/fako1024/brew/db/influx"
	"github.com/sirupsen/logrus"
)

const timestampLayout = "2006-01-02T15:04:05"

type config struct {
	id          string
	shotType    brew.ShotType
	beansWeight float64

	actionTS   time.Time
	actionType action.Type

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
	flag.Float64Var(&cfg.beansWeight, "beansWeight", 0., "Beans weight to set")

	// Flags to perform / set action (e.g. maintenance) parameters
	flag.StringVar(&timestampStr, "action-time", time.Now().Format(timestampLayout), "Timestamp at which an action was performed")
	flag.StringVar(&cfg.actionType, "action-type", "", "Type of performed action")

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

			// Check if the beans weight has been overridden
			additionalFields := make(map[string]interface{})
			if cfg.beansWeight > 0. {
				additionalFields["beans_weight"] = cfg.beansWeight
			}

			if err := influxDB.ModifyMeasurement("brews", "brew", "id", cfg.id, "shot_type", cfg.shotType.String(), additionalFields); err != nil {
				logrus.StandardLogger().Fatalf("Failed to alter measurement: %s", err)
			}
			if err := influxDB.ModifyMeasurement("brews", "summary", "id", cfg.id, "shot_type", cfg.shotType.String(), additionalFields); err != nil {
				logrus.StandardLogger().Fatalf("Failed to alter measurement summary: %s", err)
			}
			logrus.StandardLogger().Infof("Successfully changed shot type for ID %s to %s", cfg.id, cfg.shotType)
		}
	}

	if cfg.actionType != "" {

		// Attempt to parse the action timestamp
		var err error
		if cfg.actionTS, err = time.Parse(timestampLayout, timestampStr); err != nil {
			logrus.StandardLogger().Fatalf("Failed to parse time stamp for action: %s", err)
		}

		// Check if the ation type is supported
		actionCategory, isValid := action.Categorize(cfg.actionType)
		if !isValid {
			logrus.StandardLogger().Fatalf("Invalid action type: %s (supported: %v)", cfg.actionType, action.Categories())
		}

		if err := influxDB.EmitDataPoints("brews", "actions", db.DataPoints{
			{
				TimeStamp: cfg.actionTS,
				Data: map[string]interface{}{
					"type":     strings.Title(strings.Replace(cfg.actionType, "_", " ", -1)),
					"category": strings.Title(strings.Replace(actionCategory, "_", " ", -1)),
				},
				Tags: map[string]string{
					"action_type":     cfg.actionType,
					"action_category": actionCategory,
				},
			},
		}); err != nil {
			logrus.StandardLogger().Fatalf("Failed to add action: %s", err)
		}
	}
}
