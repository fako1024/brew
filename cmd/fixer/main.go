package main

import (
	"flag"

	"github.com/fako1024/brew"
	"github.com/fako1024/brew/db/influx"
	"github.com/sirupsen/logrus"
)

type config struct {
	id       string
	shotType brew.ShotType

	influxEndpoint string
	influxUser     string
	influxPassword string
}

func main() {

	var (
		cfg         config
		shotTypeStr string
	)
	flag.StringVar(&cfg.id, "id", "", "Brew ID to perform change on")
	flag.StringVar(&shotTypeStr, "shotType", "", "Shot type to set")
	flag.StringVar(&cfg.influxEndpoint, "influxEndpoint", "", "Endpoint for InfluxDB emissions")
	flag.StringVar(&cfg.influxUser, "influxUser", "root", "User for InfluxDB emissions")
	flag.StringVar(&cfg.influxPassword, "influxPassword", "root", "Password for InfluxDB emissions")
	flag.Parse()
	if cfg.id == "" {
		logrus.StandardLogger().Fatalf("No brew ID specified")
	}
	if cfg.influxEndpoint == "" {
		logrus.StandardLogger().Fatalf("No InfluxDB endpoint specified")
	}

	// Shot type change requested
	if shotTypeStr != "" {
		cfg.shotType = brew.ShotTypeFromString(shotTypeStr)
		if cfg.shotType == brew.UnknownShot {
			logrus.StandardLogger().Fatalf("Invalid shot type specified: %s", shotTypeStr)
		}
	} else {
		logrus.StandardLogger().Fatal("No action specified")
	}

	influxDB := influx.New(
		cfg.influxEndpoint,
		cfg.influxUser,
		cfg.influxPassword,
	)

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
