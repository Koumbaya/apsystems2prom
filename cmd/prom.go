package main

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

	totalGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ap_total",
			Help: "Total energy for the day, in kWh",
		})

	maxGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ap_max",
			Help: "Peak power of the day, in W",
		})

	lifeTimeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ap_lifetime",
			Help: "Lifetime generation in kWh",
		})

	panelGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ap_panels",
			Help: "Panel generation in W",
		},
		[]string{"panel_id"},
	)
)

func init() {
	prometheus.MustRegister(powerGauge)
	prometheus.MustRegister(energyGauge)
	prometheus.MustRegister(totalGauge)
	prometheus.MustRegister(maxGauge)
	prometheus.MustRegister(lifeTimeGauge)
	prometheus.MustRegister(panelGauge)
}

func updateMetrics(point dataPoint) {
	powerGauge.Set(float64(point.power))
	energyGauge.Set(point.energy)
	totalGauge.Set(point.total)
	maxGauge.Set(point.max)
	lifeTimeGauge.Set(point.lifetime)
	for _, panel := range point.panels {
		panelGauge.WithLabelValues(panel.id).Set(panel.power)
	}
}
