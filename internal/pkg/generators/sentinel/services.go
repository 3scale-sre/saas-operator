package sentinel

import (
	"fmt"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	statefulsetPodSelectorLabelKey string = "statefulset.kubernetes.io/pod-name"
)

func (gen *Generator) statefulSetService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.GetComponent() + "-headless",
			Namespace: gen.GetNamespace(),
			Labels:    gen.GetLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports:     []corev1.ServicePort{},
			Selector:  gen.GetSelector(),
		},
	}
}

func (gen *Generator) podServices(index int) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.PodServiceName(index),
			Namespace: gen.GetNamespace(),
			Labels:    gen.GetLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Name:       gen.GetComponent(),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(saasv1alpha1.SentinelPort),
				TargetPort: intstr.FromString(gen.GetComponent()),
			}},
			Selector: map[string]string{
				statefulsetPodSelectorLabelKey: fmt.Sprintf("%s-%d", gen.GetComponent(), index),
			},
		},
	}
}

// PodServiceName generates the name of the pod specific Service
func (gen *Generator) PodServiceName(index int) string {
	return fmt.Sprintf("%s-%d", gen.GetComponent(), index)
}

// SentinelEndpoints returns the list of redis URLs of all the sentinels
// These URLs point to the Pod specific Service of each sentinel Pod
func (gen *Generator) SentinelURIs() []string {
	urls := make([]string, 0, *gen.Spec.Replicas)
	for idx := range int(*gen.Spec.Replicas) {
		urls = append(urls,
			fmt.Sprintf("redis://%s.%s.svc.cluster.local:%d", gen.PodServiceName(idx), gen.GetNamespace(), saasv1alpha1.SentinelPort))
	}

	return urls
}
