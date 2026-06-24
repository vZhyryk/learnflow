package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

func RegisterDBMetrics(pool *pgxpool.Pool) {
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "db_total_connections",
			Help: "Total number of connections in the pool.",
		}, func() float64 { return float64(pool.Stat().TotalConns()) }),

		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "db_acquired_connections",
			Help: "Number of connections currently acquired.",
		}, func() float64 { return float64(pool.Stat().AcquiredConns()) }),

		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "db_idle_connections",
			Help: "Number of idle connections in the pool.",
		}, func() float64 { return float64(pool.Stat().IdleConns()) }),

		prometheus.NewCounterFunc(prometheus.CounterOpts{
			Name: "db_acquire_total",
			Help: "Total number of successful acquires from the pool.",
		}, func() float64 { return float64(pool.Stat().AcquireCount()) }),

		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "db_acquire_duration_seconds",
			Help: "Total duration spent acquiring connections (seconds).",
		}, func() float64 {
			return pool.Stat().AcquireDuration().Seconds()
		}),
	)
}
