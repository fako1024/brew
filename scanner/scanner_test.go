package scanner

import (
	"testing"
	"time"

	"github.com/fako1024/brew"
	"github.com/fako1024/brew/buffer"
	"github.com/fako1024/btscale/pkg/mock"
	"github.com/fako1024/btscale/pkg/scale"
	jsoniter "github.com/json-iterator/go"
)

func TestParseDataPoints(t *testing.T) {

	var dataPoints scale.DataPoints
	if err := jsoniter.Unmarshal([]byte(standardBrewSingle1JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}
	if err := jsoniter.Unmarshal([]byte(standardBrewSingle2JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}
	if err := jsoniter.Unmarshal([]byte(standardBrewDouble1JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}
	if err := jsoniter.Unmarshal([]byte(standardBrewDouble2JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}
}

func TestLastNIncreasing(t *testing.T) {

	buf := buffer.NewDataBuffer(1)
	buf.Append(scale.DataPoint{
		Weight: 1.0,
	})
	if lastNIncreasing(buf.LastN(1), 5) {
		t.Fatalf("Unexpected detection of increasing slope for invalid data")
	}

	buf = buffer.NewDataBuffer(100)
	for i := 1; i <= 10; i++ {
		buf.Append(scale.DataPoint{
			Weight: float64(i),
		})
	}
	for i := 2; i <= 10; i++ {
		if !lastNIncreasing(buf.LastN(i), i-1) {
			t.Fatalf("Unexpected failure to detect increasing slope for valid data")
		}
	}

	var dataPoints scale.DataPoints
	if err := jsoniter.Unmarshal([]byte(standardBrewSingle1JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}

	buf = buffer.NewDataBuffer(1024)
	for i := 0; i < len(dataPoints); i++ {
		buf.Append(dataPoints[i])

		lastN := buf.LastN(i + 2)
		lastNIncreasing(lastN, i+1)
	}

	for i := 1; i <= len(dataPoints); i++ {
		lastN := buf.LastN(i)
		lastNIncreasing(lastN, i-1)
	}
}

func TestScanDataPointsTable(t *testing.T) {

	expectedSingleShotWeight = 45.
	expectedDoubleShotWeight = 90.

	testTable := []struct {
		name                string
		data                string
		nExpectedDataPoints int
		expectedWeight      float64
		expectedShotType    brew.ShotType
	}{
		{"standardBrewSingle1", standardBrewSingle1JSON, 242, 47.81, brew.SingleShot},
		{"standardBrewSingle2", standardBrewSingle2JSON, 242, 61.64, brew.SingleShot},
		{"standardBrewDouble1", standardBrewDouble1JSON, 235, 86.72, brew.DoubleShot},
		{"standardBrewDouble2", standardBrewDouble2JSON, 231, 82.23, brew.DoubleShot},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			s, err := mock.New()
			if err != nil {
				t.Fatalf("Failed to initialize mock scale: %s", err)
			}

			scanner := New(s, nil)
			go scanner.Run()

			var dataPoints scale.DataPoints
			if err := jsoniter.Unmarshal([]byte(test.data), &dataPoints); err != nil {
				t.Fatalf("Failed to parse JSON: %s", err)
			}

			for _, dataPoint := range dataPoints {
				scanner.dataChan <- dataPoint
			}

			for i := 0; i < 3; i++ {
				time.Sleep(100 * time.Millisecond)
				if len(scanner.currentBrew.DataPoints) != test.nExpectedDataPoints {
					continue
				}

				if weight := scanner.currentBrew.DataPoints[len(scanner.currentBrew.DataPoints)-1].Weight; weight != test.expectedWeight {
					t.Fatalf("Unexpected weight, want %.2f, have %.2f", test.expectedWeight, weight)
				}

				if shotType := scanner.currentBrew.ShotType; shotType != test.expectedShotType {
					t.Fatalf("Unexpected shot type, want %s, have %s", test.expectedShotType, shotType)
				}

				return
			}

			t.Fatalf("Unexpected brew data points detected: %d", len(scanner.currentBrew.DataPoints))
		})
	}
}

//////////////////////

func BenchmarkLastNIncreasing(b *testing.B) {

	var dataPoints scale.DataPoints
	if err := jsoniter.Unmarshal([]byte(standardBrewSingle1JSON), &dataPoints); err != nil {
		b.Fatalf("Failed to parse JSON: %s", err)
	}

	buf := buffer.NewDataBuffer(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(dataPoints); i++ {
			buf.Append(dataPoints[i])

			lastN := buf.LastN(i + 2)
			lastNIncreasing(lastN, i+1)
		}

		for i := 1; i <= len(dataPoints); i++ {
			lastN := buf.LastN(i)
			lastNIncreasing(lastN, i-1)
		}
	}
}
