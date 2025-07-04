package service

import (
	"testing"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestAddMarin3rSidecar(t *testing.T) {
	type args struct {
		dep  *appsv1.Deployment
		spec saasv1alpha1.Marin3rSidecarSpec
	}

	tests := []struct {
		name string
		args args
		want *appsv1.Deployment
	}{
		{
			name: "Adds marin3r labels and annotations to a Deployment",
			args: args{
				dep: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{},
						},
					},
				},
				spec: saasv1alpha1.Marin3rSidecarSpec{
					EnvoyAPIVersion: ptr.To("xx"),
					EnvoyImage:      ptr.To("image"),
					NodeID:          ptr.To("node-id"),
					ShutdownManagerPort: func() *uint32 {
						var v uint32 = 5000
						return &v
					}(),
					ShutdownManagerExtraLifecycleHooks: []string{"container1", "container2"},
					Ports: []saasv1alpha1.SidecarPort{
						{
							Name: "test",
							Port: 9999,
						},
					},
					Resources: &saasv1alpha1.ResourceRequirementsSpec{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("200Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
					ExtraPodAnnotations: map[string]string{
						"extra-key": "extra-value",
					},
				},
			},
			want: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"marin3r.3scale.net/node-id":                                "node-id",
								"marin3r.3scale.net/envoy-image":                            "image",
								"marin3r.3scale.net/envoy-api-version":                      "xx",
								"marin3r.3scale.net/shutdown-manager.port":                  "5000",
								"marin3r.3scale.net/shutdown-manager.extra-lifecycle-hooks": "container1,container2",
								"marin3r.3scale.net/resources.limits.cpu":                   "200m",
								"marin3r.3scale.net/resources.limits.memory":                "200Mi",
								"marin3r.3scale.net/resources.requests.cpu":                 "100m",
								"marin3r.3scale.net/resources.requests.memory":              "100Mi",
								"marin3r.3scale.net/ports":                                  "test:9999",
								"marin3r.3scale.net/shutdown-manager.enabled":               "true",
								"extra-key": "extra-value",
							},
							Labels: map[string]string{
								"marin3r.3scale.net/status": "enabled",
							},
						},
					},
				},
			},
		},
		{
			name: "All empty should not fail",
			args: args{
				dep: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "dep"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{},
						},
					},
				},
				spec: saasv1alpha1.Marin3rSidecarSpec{},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep"},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"marin3r.3scale.net/node-id":                  "dep",
								"marin3r.3scale.net/shutdown-manager.enabled": "true",
							},
							Labels: map[string]string{
								"marin3r.3scale.net/status": "enabled",
							},
						},
					},
				},
			},
		},
		{
			name: "ExtraAnnotations takes precedence",
			args: args{
				dep: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "dep"},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{},
						},
					},
				},
				spec: saasv1alpha1.Marin3rSidecarSpec{
					EnvoyImage: ptr.To("image"),
					ExtraPodAnnotations: map[string]string{
						"marin3r.3scale.net/envoy-image": "override",
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep"},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"marin3r.3scale.net/shutdown-manager.enabled": "true",
								"marin3r.3scale.net/envoy-image":              "override",
								"marin3r.3scale.net/node-id":                  "dep",
							},
							Labels: map[string]string{
								"marin3r.3scale.net/status": "enabled",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddMarin3rSidecar(tt.args.dep, tt.args.spec)

			if diff := cmp.Diff(tt.args.dep, tt.want); len(diff) > 0 {
				t.Errorf("AddMarin3rSidecar() = got diff %v", diff)
			}
		})
	}
}
