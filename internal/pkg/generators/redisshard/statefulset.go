package redisshard

import (
	"fmt"
	"strings"

	"github.com/3scale-sre/basereconciler/util"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pod"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func (gen *Generator) statefulSet() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.GetKey().Name,
			Namespace: gen.GetKey().Namespace,
			Labels:    gen.GetLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			Replicas:             ptr.To[int32](gen.Replicas),
			RevisionHistoryLimit: ptr.To[int32](1),
			Selector:             &metav1.LabelSelector{MatchLabels: gen.GetSelector()},
			ServiceName:          gen.ServiceName(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: util.MergeMaps(gen.GetLabels(), gen.GetSelector()),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: func() []corev1.LocalObjectReference {
						if gen.Image.PullSecretName != nil {
							return []corev1.LocalObjectReference{{Name: *gen.Image.PullSecretName}}
						}

						return nil
					}(),
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Command: strings.Split(gen.Command, " "),
							Image:   fmt.Sprintf("%s:%s", *gen.Image.Name, *gen.Image.Tag),
							Name:    "redis-server",
							Ports: pod.ContainerPorts(
								pod.ContainerPortTCP("redis-server", 6379),
							),
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{
									Command: strings.Split("/bin/sh /redis-readiness/ready.sh", " "),
								}},
								FailureThreshold:    3,
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      5,
							},
							ImagePullPolicy: *gen.Image.PullPolicy,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "redis-config", MountPath: "/redis"},
								{Name: "redis-readiness-script", MountPath: "/redis-readiness"},
								{Name: "redis-data", MountPath: "/data"},
							},
						},
					},
					TerminationGracePeriodSeconds: ptr.To[int64](0),
					Volumes: []corev1.Volume{
						{
							Name: "redis-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									DefaultMode:          ptr.To[int32](420),
									LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config-" + gen.GetInstanceName()}},
							}},
						{
							Name: "redis-readiness-script",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									DefaultMode:          ptr.To[int32](484),
									LocalObjectReference: corev1.LocalObjectReference{Name: "redis-readiness-script-" + gen.GetInstanceName()}},
							}},
						{
							Name: "redis-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							}},
					},
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: ptr.To[int32](0),
				},
			},
		},
	}
}
