package backend

import (
	"strings"

	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pod"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/twemproxy"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func (gen *WorkerGenerator) deployment() *appsv1.Deployment {
	dep := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: gen.WorkerSpec.Replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: ptr.To(intstr.FromInt(0)),
					MaxSurge:       ptr.To(intstr.FromInt(1)),
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ImagePullSecrets: pod.ImagePullSecrets(gen.Image.PullSecretName),
					Containers: []corev1.Container{
						{
							Name:  strings.Join([]string{component, worker}, "-"),
							Image: pod.Image(gen.Image),
							Args:  []string{"bin/3scale_backend_worker", "run"},
							Ports: pod.ContainerPorts(
								pod.ContainerPortTCP("metrics", 9421),
							),
							Env:             gen.Options.BuildEnvironment(),
							Resources:       corev1.ResourceRequirements(*gen.WorkerSpec.Resources),
							ImagePullPolicy: *gen.Image.PullPolicy,
							LivenessProbe:   pod.HTTPProbe("/metrics", intstr.FromString("metrics"), corev1.URISchemeHTTP, *gen.WorkerSpec.LivenessProbe),
							ReadinessProbe:  pod.HTTPProbe("/metrics", intstr.FromString("metrics"), corev1.URISchemeHTTP, *gen.WorkerSpec.ReadinessProbe),
						},
					},
					Affinity:                      pod.Affinity(gen.GetSelector(), gen.WorkerSpec.NodeAffinity),
					Tolerations:                   gen.WorkerSpec.Tolerations,
					TerminationGracePeriodSeconds: ptr.To[int64](30),
				},
			},
		},
	}

	if gen.TwemproxySpec != nil {
		dep.Spec.Template = twemproxy.AddTwemproxySidecar(dep.Spec.Template, gen.TwemproxySpec)
	}

	return dep
}
