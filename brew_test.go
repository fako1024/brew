package brew

import (
	"testing"

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

func TestScanDataPoints(t *testing.T) {

	scanner := NewScanner(nil, nil)
	go scanner.Run()

	var dataPoints scale.DataPoints
	if err := jsoniter.Unmarshal([]byte(standardBrewSingle1JSON), &dataPoints); err != nil {
		t.Fatalf("Failed to parse JSON: %s", err)
	}

	for _, dataPoint := range dataPoints {
		scanner.dataChan <- dataPoint
	}

	select {}
}
