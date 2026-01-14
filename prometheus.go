package main

import (
	"net/http"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type StatsInternal struct {
	Registry     *prometheus.Registry
	OkCounter    prometheus.Counter
	ErrorCounter prometheus.Counter
}

func NewStatsInternal() *StatsInternal {
	si := &StatsInternal{
		Registry: prometheus.NewRegistry(),
	}

	// Custom Go Runtime collector
	goCollector := collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(
			collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")},
		),
	)
	si.Registry.MustRegister(goCollector)
	si.Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Counter for successful DB writes
	si.OkCounter = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "MECHA_WRITE_OK"})
	si.Registry.MustRegister(si.OkCounter)

	// Counter for errors
	si.ErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "MECHA_ERROR"})
	si.Registry.MustRegister(si.ErrorCounter)

	return si
}

// RecOkCounter counts successful DB writes
func (si *StatsInternal) RecOkCounter() {
	si.OkCounter.Inc()
}

// RecErrorCounter counts errors (of any kind)
func (si *StatsInternal) RecErrorCounter() {
	si.ErrorCounter.Inc()
}

// Handler is the stats web handler
func (si *StatsInternal) Handler() http.Handler {
	return promhttp.HandlerFor(si.Registry, promhttp.HandlerOpts{})
}
