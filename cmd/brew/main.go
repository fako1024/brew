package main

import (
	"github.com/fako1024/brew"
	"github.com/fako1024/brew/influx"
	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/sirupsen/logrus"
)

func main() {

	s, err := felicita.New()
	if err != nil {
		logrus.StandardLogger().Fatalf("Failed to initialize Felicita scale: %s", err)
	}
	defer s.Close()

	influxDB := influx.NewEventTracker("http://nas.f-ko.eu:8086", "root", "root")
	scanner := brew.NewScanner(s, influxDB)

	if err := scanner.Run(); err != nil {
		logrus.StandardLogger().Fatalf("Failed to scan for data: %s", err)
	}
}
