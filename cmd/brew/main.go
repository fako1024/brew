package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fako1024/brew/db/influx"
	"github.com/fako1024/brew/scanner"
	"github.com/fako1024/btscale/pkg/api"
	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/sirupsen/logrus"
)

type config struct {
	apiEndpoint string

	influxEndpoint string
	influxUser     string
	influxPassword string

	beansWeightSingle float64
	beansWeightDouble float64
	grindSetting      float64

	debug bool
}

func main() {

	var cfg config

	flag.StringVar(&cfg.apiEndpoint, "api", ":8099", "Endpoint for scale API")

	flag.StringVar(&cfg.influxEndpoint, "influxEndpoint", "", "Endpoint for InfluxDB emissions")
	flag.StringVar(&cfg.influxUser, "influxUser", "root", "User for InfluxDB emissions")
	flag.StringVar(&cfg.influxPassword, "influxPassword", "root", "Password for InfluxDB emissions")

	flag.Float64Var(&cfg.beansWeightSingle, "beansWeightSingle", scanner.DefaultSingleShotBeansWeight, "Weight of beans / grounds used for a single shot")
	flag.Float64Var(&cfg.beansWeightDouble, "beansWeightDouble", scanner.DefaultDoubleShotBeansWeight, "Weight of beans / grounds used for a double shot")
	flag.Float64Var(&cfg.grindSetting, "grindSetting", scanner.DefaultGrindSetting, "Relative grinder setting (0.0: Fine -> 1.0: Coarse)")

	flag.BoolVar(&cfg.debug, "debug", false, "Enable debugging mode (more verbose logging)")

	flag.Parse()
	if cfg.influxEndpoint == "" {
		logrus.StandardLogger().Fatalf("No InfluxDB endpoint specified")
	}
	if cfg.debug {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	}

	s, err := felicita.New()
	if err != nil {
		logrus.StandardLogger().Fatalf("Failed to initialize Felicita scale: %s", err)
	}

	if cfg.apiEndpoint != "" {
		api.New(s, cfg.apiEndpoint)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		logrus.StandardLogger().Infof("Got signal, terminating connection to scale")
		s.Close()
		os.Exit(0)
	}()

	influxDB := influx.New(
		cfg.influxEndpoint,
		cfg.influxUser,
		cfg.influxPassword,
	)
	scan := scanner.New(s, influxDB,
		scanner.WithSingleShotBeansWeight(cfg.beansWeightSingle),
		scanner.WithDoubleShotBeansWeight(cfg.beansWeightDouble),
		scanner.WithGrindSetting(cfg.grindSetting),
	)

	if err := scan.Run(); err != nil {
		logrus.StandardLogger().Fatalf("Failed to scan for data: %s", err)
	}
}
