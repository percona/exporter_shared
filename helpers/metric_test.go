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
	"sort"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestReadMetric(t *testing.T) {
	for expected, m := range map[*Metric]prometheus.Metric{
		{
			Name:   "counter",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test1"},
			Type:   dto.MetricType_COUNTER,
			Value:  36.6,
		}: prometheus.MustNewConstMetric(
			prometheus.NewDesc("counter", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.CounterValue,
			36.6,
			"test1",
		),

		{
			Name:   "gauge",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test2"},
			Type:   dto.MetricType_GAUGE,
			Value:  36.6,
		}: prometheus.MustNewConstMetric(
			prometheus.NewDesc("gauge", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.GaugeValue,
			36.6,
			"test2",
		),
		{
			Name:   "histogram",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test3"},
			Type:   dto.MetricType_HISTOGRAM,
			Value:  3.66, // mean: 36.6 / 10
			Count:  10,
			Sum:    36.6,
			Buckets: map[float64]uint64{
				0.1:         1,
				1.0:         2,
				10.0:        3,
				math.Inf(1): 4,
			},
		}: prometheus.MustNewConstHistogram(
			prometheus.NewDesc("histogram", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			10,
			36.6,
			map[float64]uint64{
				0.1:         1,
				1.0:         2,
				10.0:        3,
				math.Inf(1): 4,
			},
			"test3",
		),
		{
			Name:   "untyped",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test3"},
			Type:   dto.MetricType_UNTYPED,
			Value:  36.6,
		}: prometheus.MustNewConstMetric(
			prometheus.NewDesc("untyped", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.UntypedValue,
			36.6,
			"test3",
		),
	} {
		actual1 := ReadMetric(m)
		if !reflect.DeepEqual(expected, actual1) {
			t.Errorf("ReadMetric 1:\nexpected = %+v\nactual1 = %+v", expected, actual1)
		}

		m2 := actual1.Metric()
		actual2 := ReadMetric(m2)
		if !reflect.DeepEqual(expected, actual2) {
			t.Errorf("ReadMetric 2:\nexpected = %+v\actual2 = %+v", expected, actual2)
		}
	}
}

func TestLess(t *testing.T) {
	metrics := []prometheus.Metric{
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("counter", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.CounterValue,
			36.6,
			"test3",
		),
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
	}

	actual := ReadMetrics(metrics)
	sort.Slice(actual, func(i, j int) bool { return actual[i].Less(actual[j]) })

	expected := []*Metric{
		{
			Name:   "counter",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test1"},
			Type:   dto.MetricType_COUNTER,
			Value:  36.6,
		},
		{
			Name:   "counter",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test3"},
			Type:   dto.MetricType_COUNTER,
			Value:  36.6,
		},
		{
			Name:   "gauge",
			Help:   "metric description",
			Labels: prometheus.Labels{"job": "test", "instance": "test2"},
			Type:   dto.MetricType_GAUGE,
			Value:  36.6,
		},
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Less:\nexpected = %+v\actual = %+v", expected, actual)
	}

	if s := expected[0].String(); s != "{Name:counter Help:metric description Labels:map[instance:test1 job:test] Type:COUNTER Value:36.6 Count:0 Sum:0 Buckets:map[]}" {
		t.Errorf("Unexpected String(): %q for %#v", s, expected[0])
	}
}

func TestHistogramSupport(t *testing.T) {
	// Test creating a histogram metric
	histogram := &Metric{
		Name:   "test_histogram",
		Help:   "A test histogram",
		Labels: prometheus.Labels{"method": "GET", "status": "200"},
		Type:   dto.MetricType_HISTOGRAM,
		Value:  5.5, // mean
		Count:  100,
		Sum:    550.0,
		Buckets: map[float64]uint64{
			0.1:         10,
			0.5:         25,
			1.0:         50,
			5.0:         75,
			10.0:        90,
			math.Inf(1): 100,
		},
	}

	// Test converting to prometheus.Metric
	promMetric := histogram.Metric()

	// Test reading it back
	readBack := ReadMetric(promMetric)

	// Verify all fields match
	if readBack.Name != histogram.Name {
		t.Errorf("Name mismatch: expected %q, got %q", histogram.Name, readBack.Name)
	}
	if readBack.Help != histogram.Help {
		t.Errorf("Help mismatch: expected %q, got %q", histogram.Help, readBack.Help)
	}
	if !reflect.DeepEqual(readBack.Labels, histogram.Labels) {
		t.Errorf("Labels mismatch: expected %v, got %v", histogram.Labels, readBack.Labels)
	}
	if readBack.Type != histogram.Type {
		t.Errorf("Type mismatch: expected %v, got %v", histogram.Type, readBack.Type)
	}
	if readBack.Count != histogram.Count {
		t.Errorf("Count mismatch: expected %d, got %d", histogram.Count, readBack.Count)
	}
	if readBack.Sum != histogram.Sum {
		t.Errorf("Sum mismatch: expected %f, got %f", histogram.Sum, readBack.Sum)
	}
	if !reflect.DeepEqual(readBack.Buckets, histogram.Buckets) {
		t.Errorf("Buckets mismatch: expected %v, got %v", histogram.Buckets, readBack.Buckets)
	}
	if readBack.Value != histogram.Value {
		t.Errorf("Value mismatch: expected %f, got %f", histogram.Value, readBack.Value)
	}
}
