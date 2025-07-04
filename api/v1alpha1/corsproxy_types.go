/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/3scale-sre/basereconciler/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

var (
	corsproxyDefaultReplicas int32            = 2
	corsproxyDefaultImage    defaultImageSpec = defaultImageSpec{
		Name:       ptr.To("quay.io/3scale/cors-proxy"),
		Tag:        ptr.To("latest"),
		PullPolicy: (*corev1.PullPolicy)(ptr.To(string(corev1.PullIfNotPresent))),
	}
	corsproxyDefaultResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("75m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("150m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}
	corsproxyDefaultHPA defaultHorizontalPodAutoscalerSpec = defaultHorizontalPodAutoscalerSpec{
		MinReplicas:         ptr.To[int32](2),
		MaxReplicas:         ptr.To[int32](4),
		ResourceUtilization: ptr.To[int32](90),
		ResourceName:        ptr.To("cpu"),
	}
	corsproxyDefaultProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](3),
		TimeoutSeconds:      ptr.To[int32](1),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	corsproxyDefaultPDB defaultPodDisruptionBudgetSpec = defaultPodDisruptionBudgetSpec{
		MaxUnavailable: ptr.To(intstr.FromInt(1)),
	}

	corsproxyDefaultGrafanaDashboard defaultGrafanaDashboardSpec = defaultGrafanaDashboardSpec{
		SelectorKey:   ptr.To("monitoring-key"),
		SelectorValue: ptr.To("middleware"),
	}
)

// CORSProxySpec defines the desired state of CORSProxy
type CORSProxySpec struct {
	// Image specification for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// Pod Disruption Budget for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PDB *PodDisruptionBudgetSpec `json:"pdb,omitempty"`
	// Horizontal Pod Autoscaler for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	HPA *HorizontalPodAutoscalerSpec `json:"hpa,omitempty"`
	// Number of replicas (ignored if hpa is enabled) for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Resource requirements for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Resources *ResourceRequirementsSpec `json:"resources,omitempty"`
	// Liveness probe for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	LivenessProbe *ProbeSpec `json:"livenessProbe,omitempty"`
	// Readiness probe for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ReadinessProbe *ProbeSpec `json:"readinessProbe,omitempty"`
	// Configures the Grafana Dashboard for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	GrafanaDashboard *GrafanaDashboardSpec `json:"grafanaDashboard,omitempty"`
	// Application specific configuration options for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Config CORSProxyConfig `json:"config"`
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Describes how the services provided by this workload are exposed to its consumers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PublishingStrategies *PublishingStrategies `json:"publishingStrategies,omitempty"`
}

// Default implements defaulting for CORSProxySpec
func (spec *CORSProxySpec) Default() {
	spec.Image = InitializeImageSpec(spec.Image, corsproxyDefaultImage)
	spec.HPA = InitializeHorizontalPodAutoscalerSpec(spec.HPA, corsproxyDefaultHPA)
	spec.Replicas = intOrDefault(spec.Replicas, &corsproxyDefaultReplicas)
	spec.PDB = InitializePodDisruptionBudgetSpec(spec.PDB, corsproxyDefaultPDB)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, corsproxyDefaultResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, corsproxyDefaultProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, corsproxyDefaultProbe)
	spec.GrafanaDashboard = InitializeGrafanaDashboardSpec(spec.GrafanaDashboard, corsproxyDefaultGrafanaDashboard)
	spec.Config.Default()
}

// CORSProxyConfig defines configuration options for the component
type CORSProxyConfig struct {
	// External Secret common configuration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ExternalSecret ExternalSecret `json:"externalSecret,omitempty"`
	// System database connection string
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SystemDatabaseDSN SecretReference `json:"systemDatabaseDSN"`
}

// Default sets default values for any value not specifically set in the CORSProxyConfig struct
func (cfg *CORSProxyConfig) Default() {
	cfg.ExternalSecret.SecretStoreRef = InitializeExternalSecretSecretStoreReferenceSpec(cfg.ExternalSecret.SecretStoreRef, defaultExternalSecretSecretStoreReference)
	cfg.ExternalSecret.RefreshInterval = durationOrDefault(cfg.ExternalSecret.RefreshInterval, &defaultExternalSecretRefreshInterval)
}

// CORSProxyStatus defines the observed state of CORSProxy
type CORSProxyStatus struct {
	AggregatedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CORSProxy is the Schema for the corsproxies API
type CORSProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CORSProxySpec   `json:"spec,omitempty"`
	Status CORSProxyStatus `json:"status,omitempty"`
}

// Default implements defaulting for the CORSProxy resource
func (c *CORSProxy) Default() {
	c.Spec.Default()
}

var _ reconciler.ObjectWithAppStatus = &CORSProxy{}

func (d *CORSProxy) GetStatus() any {
	return &d.Status
}

// +kubebuilder:object:root=true

// CORSProxyList contains a list of CORSProxy
type CORSProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CORSProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CORSProxy{}, &CORSProxyList{})
}
