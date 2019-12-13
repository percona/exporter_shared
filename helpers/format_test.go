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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestFormatParse(t *testing.T) {
	metrics := []prometheus.Metric{
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("counter", "metric description", []string{"instance"}, prometheus.Labels{"job": "test1"}),
			prometheus.CounterValue,
			36.6,
			"test2",
		),
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("gauge", "metric description", []string{"instance"}, prometheus.Labels{"job": "test1"}),
			prometheus.GaugeValue,
			36.6,
			"test2",
		),
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("untyped", "metric description", []string{"instance"}, prometheus.Labels{"job": "test1"}),
			prometheus.UntypedValue,
			36.6,
			"test2",
		),
	}
	expected := []string{
		`# HELP counter metric description`,
		`# TYPE counter counter`,
		`counter{instance="test2",job="test1"} 36.6`,
		`# HELP gauge metric description`,
		`# TYPE gauge gauge`,
		`gauge{instance="test2",job="test1"} 36.6`,
		`# HELP untyped metric description`,
		`# TYPE untyped untyped`,
		`untyped{instance="test2",job="test1"} 36.6`,
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
