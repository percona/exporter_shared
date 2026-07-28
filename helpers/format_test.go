// Copyright 2017 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestFormatParse(t *testing.T) {
	metrics := []prometheus.Metric{
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("counter", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.CounterValue,
			36.6,
			"test1",
		),
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("gauge", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.GaugeValue,
			36.6,
			"test2",
		),
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("untyped", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.UntypedValue,
			36.6,
			"test3",
		),
	}
	expected := []string{
		`# HELP counter metric description`,
		`# TYPE counter counter`,
		`counter{instance="test1",job="test"} 36.6`,
		`# HELP gauge metric description`,
		`# TYPE gauge gauge`,
		`gauge{instance="test2",job="test"} 36.6`,
		`# HELP untyped metric description`,
		`# TYPE untyped untyped`,
		`untyped{instance="test3",job="test"} 36.6`,
	}

	actual := Format(metrics)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected:\n%#v\n\nactual:\n%#v", expected, actual)
		t.Errorf("expected:\n%s\n\nactual:\n%s", strings.Join(expected, "\n"), strings.Join(actual, "\n"))
	}

	metrics2 := Parse(actual)
	actual2 := Format(metrics2)
	if !reflect.DeepEqual(expected, actual2) {
		t.Errorf("expected:\n%#v\n\nactual2:\n%#v", expected, actual2)
		t.Errorf("expected:\n%s\n\nactual2:\n%s", strings.Join(expected, "\n"), strings.Join(actual2, "\n"))
	}
}

func TestHistogramFormat(t *testing.T) {
	metrics := []prometheus.Metric{
		prometheus.MustNewConstHistogram(
			prometheus.NewDesc("http_request_duration_seconds", "HTTP request latency", []string{"method"}, prometheus.Labels{"service": "api"}),
			100,
			550.0,
			map[float64]uint64{
				0.1:         10,
				0.5:         25,
				1.0:         50,
				5.0:         75,
				10.0:        90,
				math.Inf(1): 100,
			},
			"GET",
		),
	}

	// Format the histogram
	formatted := Format(metrics)

	// Check that the format includes histogram-specific lines
	expectedLines := []string{
		`# HELP http_request_duration_seconds HTTP request latency`,
		`# TYPE http_request_duration_seconds histogram`,
	}

	foundHelp := false
	foundType := false
	for _, line := range formatted {
		if line == expectedLines[0] {
			foundHelp = true
		}
		if line == expectedLines[1] {
			foundType = true
		}
		// Check for histogram bucket lines
		if strings.Contains(line, "http_request_duration_seconds_bucket") {
			if !strings.Contains(line, `method="GET"`) || !strings.Contains(line, `service="api"`) {
				t.Errorf("Histogram bucket line missing expected labels: %s", line)
			}
		}
		// Check for sum line
		if strings.Contains(line, "http_request_duration_seconds_sum") {
			if !strings.Contains(line, "550") {
				t.Errorf("Histogram sum line has unexpected value: %s", line)
			}
		}
		// Check for count line
		if strings.Contains(line, "http_request_duration_seconds_count") {
			if !strings.Contains(line, "100") {
				t.Errorf("Histogram count line has unexpected value: %s", line)
			}
		}
	}

	if !foundHelp {
		t.Error("Missing HELP line for histogram")
	}
	if !foundType {
		t.Error("Missing TYPE line for histogram")
	}

	// Test parsing the formatted histogram
	parsed := Parse(formatted)
	if len(parsed) == 0 {
		t.Fatal("Failed to parse formatted histogram")
	}

	// Verify the parsed metric
	readMetric := ReadMetric(parsed[0])
	if readMetric.Type != dto.MetricType_HISTOGRAM {
		t.Errorf("Expected HISTOGRAM type, got %v", readMetric.Type)
	}
	if readMetric.Count != 100 {
		t.Errorf("Expected count 100, got %d", readMetric.Count)
	}
	if readMetric.Sum != 550.0 {
		t.Errorf("Expected sum 550.0, got %f", readMetric.Sum)
	}
}
