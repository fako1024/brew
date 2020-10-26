package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fako1024/brew/influx"
	"github.com/fako1024/brew/scanner"
	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/sirupsen/logrus"
)

type config struct {
	influxEndpoint string
	influxUser     string
	influxPassword string
}

func main() {

	var cfg config
	flag.StringVar(&cfg.influxEndpoint, "influxEndpoint", "", "Endpoint for InfluxDB emissions")
	flag.StringVar(&cfg.influxUser, "influxUser", "root", "User for InfluxDB emissions")
	flag.StringVar(&cfg.influxPassword, "influxPassword", "root", "Password for InfluxDB emissions")
	flag.Parse()
	if cfg.influxEndpoint == "" {
		logrus.StandardLogger().Fatalf("No InfluxDB endpoint specified")
	}

	s, err := felicita.New()
	if err != nil {
		logrus.StandardLogger().Fatalf("Failed to initialize Felicita scale: %s", err)
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
	scan := scanner.New(s, influxDB)

	if err := scan.Run(); err != nil {
		logrus.StandardLogger().Fatalf("Failed to scan for data: %s", err)
	}
}
