package config

import (
	"github.com/3scale-sre/basereconciler/util"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/pkg/resource_builders/service"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DefaultPublishingStrategy() []service.ServiceDescriptor {
	return []service.ServiceDescriptor{
		{
			PublishingStrategy: saasv1alpha1.PublishingStrategy{
				Strategy:     saasv1alpha1.SimpleStrategy,
				EndpointName: "HTTP",
				Simple:       &saasv1alpha1.Simple{ServiceType: util.Pointer(saasv1alpha1.ServiceTypeClusterIP)},
			},
			PortDefinitions: []corev1.ServicePort{
				{
					Name:       "mapping",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromString("mapping"),
				},
				{
					Name:       "management",
					Protocol:   corev1.ProtocolTCP,
					Port:       8090,
					TargetPort: intstr.FromString("management"),
				},
			},
		},
	}
}
