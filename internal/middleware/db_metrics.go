package middleware

import (
	"database/sql"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type dbPoolCollector struct {
	db *sql.DB
}

func newDBPoolCollector(db *sql.DB) *dbPoolCollector {
	return &dbPoolCollector{db: db}
}

var (
	dbPoolOpenConns = prometheus.NewDesc(
		"db_pool_open_connections",
		"Number of open connections to the database.",
		nil, nil,
	)
	dbPoolInUseConns = prometheus.NewDesc(
		"db_pool_in_use_connections",
		"Number of connections currently in use.",
		nil, nil,
	)
	dbPoolIdleConns = prometheus.NewDesc(
		"db_pool_idle_connections",
		"Number of idle connections in the pool.",
		nil, nil,
	)
	dbPoolWaitCount = prometheus.NewDesc(
		"db_pool_wait_count_total",
		"Total number of times a connection was waited for.",
		nil, nil,
	)

	registerDBMetricsOnce sync.Once
)

// RegisterDBMetrics registers DB pool metrics on the default Prometheus registry.
func RegisterDBMetrics(db *sql.DB) {
	if db == nil {
		return
	}
	registerDBMetricsOnce.Do(func() {
		prometheus.MustRegister(newDBPoolCollector(db))
	})
}

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- dbPoolOpenConns
	ch <- dbPoolInUseConns
	ch <- dbPoolIdleConns
	ch <- dbPoolWaitCount
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	if c.db == nil {
		return
	}

	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(dbPoolOpenConns, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(dbPoolInUseConns, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(dbPoolIdleConns, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(dbPoolWaitCount, prometheus.CounterValue, float64(stats.WaitCount))
}
