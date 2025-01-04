package apsystems2prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	powerGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ap_power",
			Help: "Power generation in Watt",
		},
	)

	energyGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ap_energy",
			Help: "Energy generation in kWh",
		},
	)
)

func init() {
	prometheus.MustRegister(powerGauge)
	prometheus.MustRegister(energyGauge)
}

func update(point dataPoint) {
	powerGauge.Set(float64(point.power))
	energyGauge.Set(point.energy)
}
