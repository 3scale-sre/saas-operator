package config

import (
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/service"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func DefaultListenerPublishingStrategy() []service.ServiceDescriptor {
	return []service.ServiceDescriptor{
		{
			PublishingStrategy: saasv1alpha1.PublishingStrategy{
				Strategy:     saasv1alpha1.SimpleStrategy,
				EndpointName: "HTTP",
				Simple:       &saasv1alpha1.Simple{ServiceType: ptr.To(saasv1alpha1.ServiceTypeClusterIP)},
			},
			PortDefinitions: []corev1.ServicePort{{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromString("http"),
			}},
		},
	}
}
