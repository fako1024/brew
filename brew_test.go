package brew

import (
	"testing"

	"github.com/fako1024/brew/buffer"
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

// func TestScanDataPoints(t *testing.T) {
//
// 	scanner := NewScanner(nil, nil)
// 	go scanner.Run()
//
// 	var dataPoints scale.DataPoints
// 	if err := jsoniter.Unmarshal([]byte(standardBrewSingle1JSON), &dataPoints); err != nil {
// 		t.Fatalf("Failed to parse JSON: %s", err)
// 	}
//
// 	for _, dataPoint := range dataPoints {
// 		scanner.dataChan <- dataPoint
// 	}
//
// 	select {}
// }
