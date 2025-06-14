package system

import (
	"fmt"
	"strings"

	"github.com/3scale-sre/basereconciler/util"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pod"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func (gen *SearchdGenerator) statefulset() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.GetComponent(),
			Namespace: gen.Namespace,
			Labels:    gen.GetLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To[int32](1),
			Selector: &metav1.LabelSelector{MatchLabels: gen.GetSelector()},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: util.MergeMaps(map[string]string{}, gen.GetLabels(), gen.GetSelector()),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: func() []corev1.LocalObjectReference {
						if gen.Image.PullSecretName != nil {
							return []corev1.LocalObjectReference{{Name: *gen.Image.PullSecretName}}
						}

						return nil
					}(),
					Containers: []corev1.Container{
						{
							Name:  strings.Join([]string{component, searchd}, "-"),
							Image: fmt.Sprintf("%s:%s", *gen.Image.Name, *gen.Image.Tag),
							Args:  []string{},
							Ports: pod.ContainerPorts(
								pod.ContainerPortTCP("searchd", gen.DatabasePort),
							),
							Resources:       corev1.ResourceRequirements(*gen.Spec.Resources),
							LivenessProbe:   pod.TCPProbe(intstr.FromString("searchd"), *gen.Spec.LivenessProbe),
							ReadinessProbe:  pod.TCPProbe(intstr.FromString("searchd"), *gen.Spec.ReadinessProbe),
							ImagePullPolicy: *gen.Image.PullPolicy,
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "system-searchd-database",
								MountPath: gen.DatabasePath,
							}},
						},
					},
					Affinity:                      pod.Affinity(gen.GetSelector(), gen.Spec.NodeAffinity),
					Tolerations:                   gen.Spec.Tolerations,
					TerminationGracePeriodSeconds: gen.Spec.TerminationGracePeriodSeconds,
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "system-searchd-database",
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: corev1.ClaimPending,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources:        corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: gen.DatabaseStorageSize}},
					StorageClassName: gen.DatabaseStorageClass,
					VolumeMode:       (*corev1.PersistentVolumeMode)(ptr.To(string(corev1.PersistentVolumeFilesystem))),
					DataSource:       &corev1.TypedLocalObjectReference{},
				},
			}},
		},
	}
}
