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
				EndpointName: "Proxy",
				Simple:       &saasv1alpha1.Simple{ServiceType: util.Pointer(saasv1alpha1.ServiceTypeClusterIP)},
			},
			PortDefinitions: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromString("http"),
				},
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromString("https"),
				},
			},
		},
	}
}
