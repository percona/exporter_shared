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
	"reflect"
	"sort"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestReadMetric(t *testing.T) {
	for expected, m := range map[*Metric]prometheus.Metric{
		{
			"counter",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test1"},
			dto.MetricType_COUNTER,
			36.6,
		}: prometheus.MustNewConstMetric(
			prometheus.NewDesc("counter", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.CounterValue,
			36.6,
			"test1",
		),

		{
			"gauge",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test2"},
			dto.MetricType_GAUGE,
			36.6,
		}: prometheus.MustNewConstMetric(
			prometheus.NewDesc("gauge", "metric description", []string{"instance"}, prometheus.Labels{"job": "test"}),
			prometheus.GaugeValue,
			36.6,
			"test2",
		),

		{
			"untyped",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test3"},
			dto.MetricType_UNTYPED,
			36.6,
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
			"counter",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test1"},
			dto.MetricType_COUNTER,
			36.6,
		},
		{
			"counter",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test3"},
			dto.MetricType_COUNTER,
			36.6,
		},
		{
			"gauge",
			"metric description",
			prometheus.Labels{"job": "test", "instance": "test2"},
			dto.MetricType_GAUGE,
			36.6,
		},
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Less:\nexpected = %+v\actual = %+v", expected, actual)
	}

	if s := expected[0].String(); s != "{Name:counter Help:metric description Labels:map[instance:test1 job:test] Type:COUNTER Value:36.6}" {
		t.Errorf("Unexpected String(): %q for %#v", s, expected[0])
	}
}
