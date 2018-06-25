package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type gnmiCollector struct {
	fooMetric *prometheus.Desc
}

func newGNMICollector() *gnmiCollector {
	return &gnmiCollector{
		fooMetric: prometheus.NewDesc("foo_metric",
			"Shows whether a foo has occurred in our cluster",
			nil, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *gnmiCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.fooMetric
}

//Collect implements required collect function.
func (collector *gnmiCollector) Collect(ch chan<- prometheus.Metric) {

	var metricValue float64
	if 1 == 1 {
		metricValue = 1
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.fooMetric, prometheus.CounterValue, metricValue)

}
