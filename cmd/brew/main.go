package main

import (
	"os"
	"os/signal"
	"syscall"

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

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		logrus.StandardLogger().Infof("Got signal, terminating connection to scale")
		s.Close()
		os.Exit(0)
	}()

	influxDB := influx.NewEventTracker("http://nas.f-ko.eu:8086", "root", "root")
	scanner := brew.NewScanner(s, influxDB)

	if err := scanner.Run(); err != nil {
		logrus.StandardLogger().Fatalf("Failed to scan for data: %s", err)
	}
}
