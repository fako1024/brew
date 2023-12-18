package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fako1024/brew/db/influx"
	"github.com/fako1024/brew/scanner"
	"github.com/fako1024/btscale/pkg/api"
	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/fako1024/gatt"
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
	logger := scale.NewDefaultLogger(cfg.debug)

	if cfg.influxEndpoint == "" {
		logger.Fatalf("no InfluxDB endpoint specified")
	}

	btDevice, err := gatt.NewDevice([]gatt.Option{
		gatt.LnxMaxConnections(2),
		gatt.LnxDeviceID(-1, true),
		gatt.LnxMsgTimeout(10 * time.Second),
	}...)
	if err != nil {
		logger.Fatalf("failed to initialize bluetooth system device: %s", err)
	}

	s, err := felicita.New(felicita.WithDevice(btDevice), felicita.WithDeviceID("C8:FD:19:8E:3E:3C"), felicita.WithLogger(logger))
	if err != nil {
		logger.Fatalf("failed to initialize Felicita scale: %s", err)
	}
	sStateChan := make(chan scale.ConnectionStatus)
	s.SetStateChangeChannel(sStateChan)
	go func() {
		for st := range sStateChan {
			logger.Infof("scale state change: %v", st)
		}
	}()

	if cfg.apiEndpoint != "" {
		api.New(s, cfg.apiEndpoint)
	}

	sigChan := make(chan os.Signal, 3)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		logger.Infof("got signal, terminating connection to scale / meater")
		if err := s.Close(); err != nil {
			logger.Errorf("failed to close scale: %s", err)
		}
		if err := btDevice.Close(); err != nil {
			logger.Errorf("failed to stop bluetooth device: %s", err)
		}
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
		scanner.WithLogger(logger),
	)

	if err := scan.Run(); err != nil {
		logger.Fatalf("failed to scan for data: %s", err)
	}
}
