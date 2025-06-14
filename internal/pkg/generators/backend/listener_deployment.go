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

func (gen *ListenerGenerator) deployment() *appsv1.Deployment {
	dep := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: gen.ListenerSpec.Replicas,
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
							Name:  strings.Join([]string{component, listener}, "-"),
							Image: pod.Image(gen.Image),
							Args: func() (args []string) {
								if *gen.ListenerSpec.Config.RedisAsync {
									args = []string{"bin/3scale_backend", "-s", "falcon", "start"}
								} else {
									args = []string{"bin/3scale_backend", "start"}
								}
								args = append(args, "-e", "production", "-p", "3000", "-x", "/dev/stdout")

								return
							}(),
							Ports: pod.ContainerPorts(
								pod.ContainerPortTCP("http", 3000),
								pod.ContainerPortTCP("metrics", 9394),
							),
							Env:             gen.Options.BuildEnvironment(),
							Resources:       corev1.ResourceRequirements(*gen.ListenerSpec.Resources),
							ImagePullPolicy: *gen.Image.PullPolicy,
							LivenessProbe:   pod.TCPProbe(intstr.FromString("http"), *gen.ListenerSpec.LivenessProbe),
							ReadinessProbe:  pod.HTTPProbe("/status", intstr.FromString("http"), corev1.URISchemeHTTP, *gen.ListenerSpec.ReadinessProbe),
						},
					},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					Affinity:                      pod.Affinity(gen.GetSelector(), gen.ListenerSpec.NodeAffinity),
					Tolerations:                   gen.ListenerSpec.Tolerations,
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
