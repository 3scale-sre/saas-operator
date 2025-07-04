package deployment

import (
	"github.com/3scale-sre/basereconciler/resource"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	descriptor "github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/descriptor"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/service"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

/* Each of the workload types can be composed
of several of the features, each one of them
described by one of the following interfaces */

type WithKey interface {
	GetKey() types.NamespacedName
}

type WithLabels interface {
	GetLabels() map[string]string
}

type WithSelector interface {
	GetSelector() map[string]string
}

type WithWorkloadMeta interface {
	WithKey
	WithLabels
	WithSelector
}

type WithMonitoring interface {
	MonitoredEndpoints() []monitoringv1.PodMetricsEndpoint
}

type WithPodDisruptionBadget interface {
	PDBSpec() *saasv1alpha1.PodDisruptionBudgetSpec
}

type WithHorizontalPodAutoscaler interface {
	HPASpec() *saasv1alpha1.HorizontalPodAutoscalerSpec
}

type WithCanary interface {
	WithWorkloadMeta
	WithSelector
	SendTraffic() bool
	TrafficSelector() map[string]string
}

type WithMarin3rSidecar interface {
	WithWorkloadMeta
	EnvoyDynamicConfigurations() []descriptor.EnvoyDynamicConfigDescriptor
}

type WithPublishingStrategies interface {
	WithWorkloadMeta
	WithSelector
	WithCanary
	PublishingStrategies() ([]service.ServiceDescriptor, error)
}

type DeploymentWorkload interface {
	WithWorkloadMeta
	WithMonitoring
	WithHorizontalPodAutoscaler
	WithPodDisruptionBadget
	Deployment() *resource.Template[*appsv1.Deployment]
}
