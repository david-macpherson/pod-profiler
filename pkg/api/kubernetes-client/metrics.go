package kubernetesclient

import (
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Metrics struct {
	clientSet *metrics.Clientset
	namespace string
}

// Convenience function to return the pod metrics interface
func (m *Metrics) Pod() v1beta1.PodMetricsInterface {
	return m.clientSet.MetricsV1beta1().PodMetricses(m.namespace)
}

// Convenience function to return the node metrics interface
func (m *Metrics) Node() v1beta1.NodeMetricsInterface {
	return m.clientSet.MetricsV1beta1().NodeMetricses()
}
