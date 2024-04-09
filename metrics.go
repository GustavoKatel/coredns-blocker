package blocker

import (
	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// blockedEntries is number of requests blocked
	blockedEntries = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "blocker",
		Name:      "entries_blocked",
		Help:      "The number of requests blocked",
	}, []string{"domain"})
)
