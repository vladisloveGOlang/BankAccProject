package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type MetricsCounters struct {
	RepoCounter      *prometheus.CounterVec
	RepoHistogram    *prometheus.HistogramVec
	RequestHistogram *prometheus.HistogramVec
	DicGauge         *prometheus.GaugeVec
}

func NewMetricsCounters() *MetricsCounters {
	repoCounterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "repo_requests_total",
			Help: "How many requests repo was handlered.",
		},
		[]string{"repo"},
	)

	if err := prometheus.Register(repoCounterVec); err != nil {
		logrus.Fatal(err)
	}

	dicGaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dictionary_items",
			Help: "How many items in dictionary.",
		},
		[]string{"name"},
	)

	if err := prometheus.Register(dicGaugeVec); err != nil {
		logrus.Fatal(err)
	}

	requestHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_response_time",
			Help: "How long it took to get a response.",
		},
		[]string{"method", "name"},
	)

	if err := prometheus.Register(requestHistogram); err != nil {
		logrus.Fatal(err)
	}

	repoHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "repo_response_time",
			Help: "How long it took to get a response from repo.",
		},
		[]string{"name"},
	)

	if err := prometheus.Register(repoHistogram); err != nil {
		logrus.Fatal(err)
	}

	return &MetricsCounters{
		RepoCounter:      repoCounterVec,
		RepoHistogram:    repoHistogram,
		RequestHistogram: requestHistogram,
		DicGauge:         dicGaugeVec,
	}
}
