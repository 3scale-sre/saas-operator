package twemproxy

import (
	"path/filepath"

	"github.com/3scale-sre/basereconciler/util"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pod"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	twemproxy                  = "twemproxy"
	twemproxyPreStopScriptName = "pre-stop"
	healthCommand              = "health"
)

func TwemproxyContainer(twemproxySpec *saasv1alpha1.TwemproxySpec) corev1.Container {
	return corev1.Container{
		Env:   NewOptions(*twemproxySpec).BuildEnvironment(),
		Name:  twemproxy,
		Image: pod.Image(*twemproxySpec.Image),
		Ports: pod.ContainerPorts(
			pod.ContainerPortTCP(twemproxy, 22121),
			pod.ContainerPortTCP("twem-metrics", *twemproxySpec.Options.MetricsPort),
		),
		Resources:       corev1.ResourceRequirements(*twemproxySpec.Resources),
		ImagePullPolicy: *twemproxySpec.Image.PullPolicy,
		LivenessProbe:   pod.ExecProbe(healthCommand, *twemproxySpec.LivenessProbe),
		ReadinessProbe:  pod.ExecProbe(healthCommand, *twemproxySpec.ReadinessProbe),
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"pre-stop", TwemproxyConfigFile},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      twemproxy + "-config",
				MountPath: filepath.Dir(TwemproxyConfigFile),
			},
		},
	}
}

func TwemproxyContainerVolume(twemproxySpec *saasv1alpha1.TwemproxySpec) corev1.Volume {
	return corev1.Volume{
		Name: twemproxy + "-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: twemproxySpec.ConfigMapName(),
				},
				DefaultMode: ptr.To[int32](420),
			},
		},
	}
}

func AddTwemproxySidecar(podTemplateSpec corev1.PodTemplateSpec, twemproxySpec *saasv1alpha1.TwemproxySpec) corev1.PodTemplateSpec {
	// Labels to subscribe to the TwemproxyConfig sync events
	podTemplateSpec.ObjectMeta.Labels = util.MergeMaps(
		map[string]string{},
		podTemplateSpec.GetLabels(),
		map[string]string{saasv1alpha1.TwemproxyPodSyncLabelKey: twemproxySpec.TwemproxyConfigRef},
	)

	// Twemproxy container
	podTemplateSpec.Spec.Containers = append(
		podTemplateSpec.Spec.Containers,
		TwemproxyContainer(twemproxySpec),
	)

	if podTemplateSpec.Spec.Volumes == nil {
		podTemplateSpec.Spec.Volumes = []corev1.Volume{}
	}

	// Mount the TwemproxyConfig ConfigMap in the Pod
	podTemplateSpec.Spec.Volumes = append(
		podTemplateSpec.Spec.Volumes, TwemproxyContainerVolume(twemproxySpec),
	)

	return podTemplateSpec
}

func AddTwemproxyTaskSidecar(taskSpec pipelinev1.TaskSpec, twemproxySpec *saasv1alpha1.TwemproxySpec) pipelinev1.TaskSpec {
	twemproxySidecar := pipelinev1.Sidecar{}
	twemproxySidecar.SetContainerFields(TwemproxyContainer(twemproxySpec))

	// Twemproxy container
	taskSpec.Sidecars = append(taskSpec.Sidecars, twemproxySidecar)

	if taskSpec.Volumes == nil {
		taskSpec.Volumes = []corev1.Volume{}
	}

	// Mount the TwemproxyConfig ConfigMap in the Pod
	taskSpec.Volumes = append(
		taskSpec.Volumes, TwemproxyContainerVolume(twemproxySpec),
	)

	return taskSpec
}
