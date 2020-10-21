package brew

import (
	"math"
	"time"

	"github.com/fako1024/brew/influx"
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultDataChanDepth     = 256
	expectedSingleShotWeight = 35.
	expectedDoubleShotWeight = 70.

	minBrewTime = 10 * time.Second
	maxBrewTime = 60 * time.Second
)

// Scanner denotes a brew scanner that constantly analyzes weight data from a scale
// and automatically creates / tracks brews
type Scanner struct {
	scale    scale.Scale          // The scale to use for measurement
	influxDB *influx.EventTracker // The InfluxDb endpoint for data submission

	dataChan    chan scale.DataPoint // The data channel to receive measurements on
	dataBuf     *DataBuffer          // The ring buffer to keep the last n measurements
	currentBrew *Brew                // The currently ongoing brew process
}

// NewScanner initializes a new brew scanner instance
func NewScanner(s scale.Scale, influxDB *influx.EventTracker) *Scanner {
	return &Scanner{
		scale:    s,
		influxDB: influxDB,
		dataBuf:  NewDataBuffer(1024),
		dataChan: make(chan scale.DataPoint, defaultDataChanDepth),
	}
}

// Run starts to continuously scan for data and process it (blocking method)
func (s *Scanner) Run() error {

	// Set the data channel
	s.scale.SetDataChannel(s.dataChan)

	var currentlyTrackingBrew bool

	// Loop over channel and process each arriving data point
	for dataPoint := range s.dataChan {

		s.dataBuf.Append(dataPoint)
		last5 := s.dataBuf.LastN(5)

		if !currentlyTrackingBrew {
			logrus.StandardLogger().Infof("%.1f %.1f %.1f %.1f %.1f", last5[0].Weight, last5[1].Weight, last5[2].Weight, last5[3].Weight, last5[4].Weight)
			if last5[0].Weight > last5[1].Weight && last5[1].Weight > last5[2].Weight && last5[2].Weight > last5[3].Weight && last5[3].Weight > last5[4].Weight {
				s.currentBrew = &Brew{
					ID:         uuid.New().String(),
					Start:      last5[4].TimeStamp,
					DataPoints: scale.DataPoints{last5[4], last5[3], last5[2], last5[1], last5[0]},
				}
				logrus.StandardLogger().Infof("Starting tracking brew: %v", last5[0])
				currentlyTrackingBrew = true
			}
		} else {
			s.currentBrew.DataPoints = append(s.currentBrew.DataPoints, last5[0])
			if last5[0].Weight-last5[1].Weight < 0.1 && last5[1].Weight-last5[2].Weight < 0.1 && last5[2].Weight-last5[3].Weight < 0.1 && last5[3].Weight-last5[4].Weight < 0.1 {
				s.currentBrew.End = last5[0].TimeStamp

				if s.currentBrew.End.Sub(s.currentBrew.Start) < minBrewTime {
					logrus.StandardLogger().Errorf("Brew time too short, ignoring data points")
					currentlyTrackingBrew = false
					continue
				} else if s.currentBrew.End.Sub(s.currentBrew.Start) > maxBrewTime {
					logrus.StandardLogger().Errorf("Brew time too long, ignoring data points")
					currentlyTrackingBrew = false
					continue
				}

				if math.Abs(expectedSingleShotWeight-last5[0].Weight) < math.Abs(expectedDoubleShotWeight-last5[0].Weight) {
					s.currentBrew.ShotType = SingleShot
					s.scale.Buzz(1)
				} else {
					s.currentBrew.ShotType = DoubleShot
					s.scale.Buzz(2)
				}

				// If brew was successfully tracked, store data into InfluxDB
				logrus.StandardLogger().Infof("Got brew: %#v", s.currentBrew)
				if s.influxDB != nil {

					// Generate tags
					tags := map[string]string{
						"id":        s.currentBrew.ID,
						"shot_type": s.currentBrew.ShotType.String(),
					}

					// Generate data points from brew data
					var dataPoints influx.DataPoints
					for _, v := range s.currentBrew.DataPoints {
						dataPoints = append(dataPoints, influx.DataPoint{
							TimeStamp: v.TimeStamp,
							Tags:      tags,
							Data: map[string]interface{}{
								"unit":   v.Unit,
								"weight": v.Weight,
							},
						})
					}

					// Emit the summary to the influxDB
					if err := s.influxDB.EmitDataPoints("brews", "summary", influx.DataPoints{
						{
							TimeStamp: s.currentBrew.Start,
							Tags:      tags,
							Data: map[string]interface{}{
								"start":         s.currentBrew.Start.Unix() * 1000,
								"end":           s.currentBrew.End.Unix() * 1000,
								"end_weight":    s.currentBrew.DataPoints[len(s.currentBrew.DataPoints)-1].Weight,
								"unit":          s.currentBrew.DataPoints[len(s.currentBrew.DataPoints)-1].Unit,
								"battery_level": s.scale.BatteryLevel(),
							},
						},
					}); err != nil {
						logrus.StandardLogger().Errorf("Failed to emit brew summary to influxDB: %s", err)
					}

					// Emit the data points to the influxDB
					if err := s.influxDB.EmitDataPoints("brews", "brew", dataPoints); err != nil {
						logrus.StandardLogger().Errorf("Failed to emit brew data points to influxDB: %s", err)
					}
				}

				currentlyTrackingBrew = false
			}
		}

		// jsonData, err := jsoniter.Marshal(dataPoint)
		// if err != nil {
		// 	logrus.StandardLogger().Errorf("Error parsing data point %#v: %s", *dataPoint, err)
		// 	continue
		// }
		//
		// logrus.StandardLogger().Infof("%s", string(jsonData))
	}

	return nil
}
