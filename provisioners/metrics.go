package provisioners

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricsNamespace = "cfssl_issuer"
)

var (
	signRequests = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Help:      "duration in seconds for signing requests",
		Namespace: metricsNamespace,
		Name:      "sign_request_seconds",
		Buckets:   []float64{0.05, 0.1, 0.5, 1.0, 5.0},
	}, []string{"profile"})
	signErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Help:      "upstream signing errors",
		Namespace: metricsNamespace,
		Name:      "sign_errors",
	}, []string{"profile"})
)

func init() {
	metrics.Registry.MustRegister(signRequests)
	metrics.Registry.MustRegister(signErrors)
}
