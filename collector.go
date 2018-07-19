package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type gnmiCollector struct {
	up   *prometheus.Desc
	ifdb *intfDB
}

func newGNMICollector(ifdb *intfDB) *gnmiCollector {
	return &gnmiCollector{
		up:   prometheus.NewDesc("arista_ceos_up", "Arista cEOS1 is up", nil, nil),
		ifdb: ifdb,
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *gnmiCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.up
}

//Collect implements required collect function.
func (collector *gnmiCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 1)
	for intfname, intf := range collector.ifdb.db {
		desc := prometheus.NewDesc("in_broadcast_packets_total", "Broadcast packets recieved", []string{"interface"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InBroadcastPkts), intfname, "cEOS1")
		desc = prometheus.NewDesc("in_discard_packets_total", "Discard packets recieved", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InDiscards), intfname, "cEOS1")
		desc = prometheus.NewDesc("in_error_packets_total", "Error packets recieved", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InErrors), intfname, "cEOS1")
		desc = prometheus.NewDesc("in_multicast_packets_total", "Multicast packets recieved", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InMulticastPkts), intfname, "cEOS1")
		desc = prometheus.NewDesc("in_bytes_total", "Total bytes recieved", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InOctets), intfname, "cEOS1")
		desc = prometheus.NewDesc("in_unicast_packets_total", "Total unicast recieved", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InUnicastPkts), intfname, "cEOS1")
		desc = prometheus.NewDesc("out_discard_packets_total", "Total discard packets sent", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.InBroadcastPkts), intfname, "cEOS1")
		desc = prometheus.NewDesc("out_unicast_packets_total", "Total unicast packets sent", []string{"interface", "switch"}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(intf.stats.OutUnicastPkts), intfname, "cEOS1")

	}

}
