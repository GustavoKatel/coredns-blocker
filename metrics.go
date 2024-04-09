package blocker

import (
	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// processedEntries is number of requests processed by domain
	processedEntries = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: PluginName,
		Name:      "entries_processed",
		Help:      "The number of requests blocked",
	}, []string{"status"})

	blocklistSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: PluginName,
		Name:      "blocklist_size",
		Help:      "The combined number of entries in hosts and Corefile.",
	}, []string{"source"})
)
