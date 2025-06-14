package podmonitor

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New returns a basereconciler_types.GeneratorFunction function that will return a PodMonitor
// resource when called
func New(key types.NamespacedName, labels map[string]string, selector map[string]string,
	endpoints ...monitoringv1.PodMetricsEndpoint) func(client.Object) (*monitoringv1.PodMonitor, error) {
	return func(client.Object) (*monitoringv1.PodMonitor, error) {
		return &monitoringv1.PodMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels:    labels,
			},
			Spec: monitoringv1.PodMonitorSpec{
				PodMetricsEndpoints: endpoints,
				Selector: metav1.LabelSelector{
					MatchLabels: selector,
				},
			},
		}, nil
	}
}

// PodMetricsEndpoint returns a monitoringv1.PodMetricsEndpoint
func PodMetricsEndpoint(path, port string, interval int32) monitoringv1.PodMetricsEndpoint {
	return monitoringv1.PodMetricsEndpoint{
		Interval: monitoringv1.Duration(fmt.Sprintf("%ds", interval)),
		Path:     path,
		Port:     ptr.To(port),
		Scheme:   "http",
	}
}
