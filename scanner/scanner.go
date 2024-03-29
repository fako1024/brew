package scanner

import (
	"math"
	"time"

	"github.com/fako1024/brew"
	"github.com/fako1024/brew/buffer"
	"github.com/fako1024/brew/db"
	"github.com/fako1024/brew/db/influx"
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/google/uuid"
)

const (
	defaultDataChanDepth = 256

	minBrewTime = 10 * time.Second
	maxBrewTime = 60 * time.Second

	defaultExpectedSingleShotWeight = 30.
	defaultExpectedDoubleShotWeight = 65.

	// DefaultSingleShotBeansWeight denotes the default weight of beans
	// / grounds used for a single shot
	DefaultSingleShotBeansWeight = 8.75

	// DefaultDoubleShotBeansWeight denotes the default weight of beans
	// / grounds used for a double shot
	DefaultDoubleShotBeansWeight = 16.0

	// DefaultGrindSetting denotes the relative grinder setting
	// 0.0: Fine
	// 1.0: Coarse
	DefaultGrindSetting = 0.208695652 // Mahlkönig Vario V2: (23*2 + 2) / 230 (3B setting)
)

// Scanner denotes a brew scanner that constantly analyzes weight data from a scale
// and automatically creates / tracks brews
type Scanner struct {
	scale    scale.Scale // The scale to use for measurement
	influxDB db.DB       // The database endpoint for data submission

	dataChan    chan scale.DataPoint // The data channel to receive measurements on
	dataBuf     *buffer.DataBuffer   // The ring buffer to keep the last n measurements
	currentBrew *brew.Brew           // The currently ongoing brew process

	expectedSingleShotWeight float64
	expectedDoubleShotWeight float64
	singleShotBeansWeight    float64
	doubleShotBeansWeight    float64
	grindSetting             float64

	logger scale.Logger
}

// New initializes a new brew scanner instance
func New(s scale.Scale, influxDB *influx.DB, options ...func(*Scanner)) *Scanner {
	scanner := &Scanner{
		scale:    s,
		dataBuf:  buffer.NewDataBuffer(1024),
		dataChan: make(chan scale.DataPoint, defaultDataChanDepth),

		expectedSingleShotWeight: defaultExpectedSingleShotWeight,
		expectedDoubleShotWeight: defaultExpectedDoubleShotWeight,

		singleShotBeansWeight: DefaultSingleShotBeansWeight,
		doubleShotBeansWeight: DefaultDoubleShotBeansWeight,
		grindSetting:          DefaultGrindSetting,
		logger:                &scale.NullLogger{},
	}

	if influxDB != nil {
		scanner.influxDB = influxDB
	}

	// Execute functional options, if any
	for _, opt := range options {
		opt(scanner)
	}

	return scanner
}

// Run starts to continuously scan for data and process it (blocking method)
func (s *Scanner) Run() error {

	// Set the data channel
	s.scale.SetDataChannel(s.dataChan)

	var currentlyTrackingBrew bool

	// Loop over channel and process each arriving data point
	for dataPoint := range s.dataChan {

		s.logger.Debugf("tracking data point %#v (Scale Battery Level: %.2f (raw %d)", dataPoint, s.scale.BatteryLevel(), s.scale.BatteryLevelRaw())

		s.dataBuf.Append(dataPoint)
		last5 := s.dataBuf.LastN(5)

		if !currentlyTrackingBrew {
			if lastNIncreasing(last5, 4) {
				s.currentBrew = &brew.Brew{
					ID:         uuid.New().String(),
					Start:      last5[0].(scale.DataPoint).TimeStamp,
					DataPoints: scale.DataPoints{last5[0].(scale.DataPoint), last5[1].(scale.DataPoint), last5[2].(scale.DataPoint), last5[3].(scale.DataPoint), last5[4].(scale.DataPoint)},
				}
				s.logger.Infof("starting tracking brew: %v", last5[0])
				currentlyTrackingBrew = true
			}
		} else {
			s.currentBrew.DataPoints = append(s.currentBrew.DataPoints, last5[4].(scale.DataPoint))
			if lastNStatic(last5, 4, 0.05) {
				s.currentBrew.End = last5[4].(scale.DataPoint).TimeStamp

				if elapsed := s.currentBrew.End.Sub(s.currentBrew.Start); elapsed < minBrewTime {
					s.logger.Warnf("brew time too short (%v), ignoring data points", elapsed)
					currentlyTrackingBrew = false
					continue
				} else if elapsed > maxBrewTime {
					s.logger.Warnf("brew time too long (%v), ignoring data points", elapsed)
					currentlyTrackingBrew = false
					continue
				}

				if math.Abs(s.expectedSingleShotWeight-last5[4].Value()) < math.Abs(s.expectedDoubleShotWeight-last5[4].Value()) {
					s.currentBrew.ShotType = brew.SingleShot
					s.scale.Buzz(1)
				} else {
					s.currentBrew.ShotType = brew.DoubleShot
					s.scale.Buzz(2)
				}

				// If brew was successfully tracked, store data into InfluxDB
				s.logger.Infof("finished tracking brew: %#v", s.currentBrew)
				if s.influxDB != nil {

					// Generate tags
					tags := map[string]string{
						"id":        s.currentBrew.ID,
						"shot_type": s.currentBrew.ShotType.String(),
					}

					// Generate data points from brew data
					var dataPoints db.DataPoints
					for _, v := range s.currentBrew.DataPoints {
						dataPoints = append(dataPoints, db.DataPoint{
							TimeStamp: v.TimeStamp,
							Tags:      tags,
							Data: map[string]interface{}{
								"unit":   v.Unit,
								"weight": v.Weight,
							},
						})
					}

					// Define the weight of the beans / grounds used for the
					// single  or double shot, respectively
					beansWeight := s.doubleShotBeansWeight
					if s.currentBrew.ShotType == brew.SingleShot {
						beansWeight = s.singleShotBeansWeight
					}

					// Emit the summary to the influxDB
					if s.influxDB != nil {
						if err := s.influxDB.EmitDataPoints("brews", "summary", db.DataPoints{
							{
								TimeStamp: s.currentBrew.Start,
								Tags:      tags,
								Data: map[string]interface{}{
									"start":         s.currentBrew.Start.Unix() * 1000,
									"end":           s.currentBrew.End.Unix() * 1000,
									"end_weight":    s.currentBrew.DataPoints[len(s.currentBrew.DataPoints)-1].Weight,
									"unit":          s.currentBrew.DataPoints[len(s.currentBrew.DataPoints)-1].Unit,
									"battery_level": s.scale.BatteryLevel(),
									"beans_weight":  beansWeight,
									"grind_setting": s.grindSetting,
								},
							},
						}); err != nil {
							s.logger.Errorf("failed to emit brew summary to influxDB: %s", err)
						}

						// Emit the data points to the influxDB
						if err := s.influxDB.EmitDataPoints("brews", "brew", dataPoints); err != nil {
							s.logger.Errorf("failed to emit brew data points to influxDB: %s", err)
						}
					}
				}

				currentlyTrackingBrew = false
			}
		}
	}

	return nil
}

func lastNIncreasing(data buffer.DataPoints, n int) bool {
	return lastNIncreasingBy(data, n, 0.0)
}

func lastNIncreasingBy(data buffer.DataPoints, n int, change float64) bool {

	// Validate data buffer length is sufficient
	if len(data) < n {
		return false
	}

	for i := 0; i < n; i++ {

		// Check if data point is valid
		if data[i] == nil || data[i+1] == nil {
			return false
		}

		// Check if data has increased from step i to i+1 by change
		if data[i+1].Value() <= data[i].Value()+change {
			return false
		}
	}

	return true
}

func lastNStatic(data buffer.DataPoints, n int, maxChange float64) bool {

	// Validate data buffer length is sufficient
	if len(data) < n {
		return false
	}

	for i := 0; i < n; i++ {

		// Check if data point is valid
		if data[i] == nil || data[i+1] == nil {
			return false
		}

		// Check if data has increased from step i to i+1 by change
		if data[i+1].Value() >= data[i].Value()+maxChange {
			return false
		}
	}

	return true
}
