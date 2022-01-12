package mappingservice

import (
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/service"
	basereconciler "github.com/3scale/saas-operator/pkg/reconcilers/basereconciler/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service returns a basereconciler.GeneratorFunction function that will return the
// management Service resource when called
func (gen *Generator) Service() basereconciler.GeneratorFunction {

	return func() client.Object {

		return &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      gen.GetComponent(),
				Namespace: gen.GetNamespace(),
				Labels:    gen.GetLabels(),
			},
			Spec: corev1.ServiceSpec{
				Type:            corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
				Ports: service.Ports(
					service.TCPPort("mapping", 80, intstr.FromString("mapping")),
					service.TCPPort("management", 8090, intstr.FromString("management")),
				),
				Selector: gen.Selector().MatchLabels,
			},
		}
	}
}
