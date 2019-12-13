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

// Package helpers provides test helpers for Prometheus exporters.
//
// It contains workarounds for the following issues:
//  * https://github.com/prometheus/client_golang/issues/322
//  * https://github.com/prometheus/client_golang/issues/323
package helpers

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var nameAndHelpRE = regexp.MustCompile(`fqName: "(\w+)", help: "([^"]+)"`)

func getNameAndHelp(d *prometheus.Desc) (string, string) {
	m := nameAndHelpRE.FindStringSubmatch(d.String())
	if len(m) != 3 {
		panic(fmt.Sprintf("failed to get metric name and help from %#q: %#v", d.String(), m))
	}
	return m[1], m[2]
}

// Metric contains Prometheus metric details.
type Metric struct {
	Name   string
	Help   string
	Labels prometheus.Labels
	Type   dto.MetricType
	Value  float64
}

// Metric returns Prometheus metric with same information.
func (m *Metric) Metric() prometheus.Metric {
	var valueType prometheus.ValueType
	switch m.Type {
	case dto.MetricType_GAUGE:
		valueType = prometheus.GaugeValue
	case dto.MetricType_COUNTER:
		valueType = prometheus.CounterValue
	case dto.MetricType_UNTYPED:
		valueType = prometheus.UntypedValue
	default:
		panic(fmt.Sprintf("Unsupported metric type %#v", m.Type))
	}

	return prometheus.MustNewConstMetric(prometheus.NewDesc(m.Name, m.Help, nil, m.Labels), valueType, m.Value)
}

func readDTOMetric(m *dto.Metric) (labels prometheus.Labels, typ dto.MetricType, value float64) {
	labels = make(prometheus.Labels, len(m.Label))
	for _, pair := range m.Label {
		labels[pair.GetName()] = pair.GetValue()
	}

	switch {
	case m.Gauge != nil:
		typ = dto.MetricType_GAUGE
		value = m.GetGauge().GetValue()
	case m.Counter != nil:
		typ = dto.MetricType_COUNTER
		value = m.GetCounter().GetValue()
	case m.Untyped != nil:
		typ = dto.MetricType_UNTYPED
		value = m.GetUntyped().GetValue()
	default:
		panic("unhandled metric type")
	}

	return
}

// ReadMetric extracts details from Prometheus metric.
func ReadMetric(metric prometheus.Metric) *Metric {
	var m dto.Metric
	if err := metric.Write(&m); err != nil {
		panic(err)
	}

	name, help := getNameAndHelp(metric.Desc())
	labels, typ, value := readDTOMetric(&m)
	return &Metric{name, help, labels, typ, value}
}

// ReadMetrics extracts details from Prometheus metrics.
func ReadMetrics(metrics []prometheus.Metric) []*Metric {
	res := make([]*Metric, len(metrics))
	for i, m := range metrics {
		res[i] = ReadMetric(m)
	}
	return res
}
