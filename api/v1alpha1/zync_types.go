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
	zyncDefaultImage defaultImageSpec = defaultImageSpec{
		Name:       ptr.To("quay.io/3scale/zync"),
		Tag:        ptr.To("nightly"),
		PullPolicy: (*corev1.PullPolicy)(ptr.To(string(corev1.PullIfNotPresent))),
	}
	zyncDefaultGrafanaDashboard defaultGrafanaDashboardSpec = defaultGrafanaDashboardSpec{
		SelectorKey:   ptr.To("monitoring-key"),
		SelectorValue: ptr.To("middleware"),
	}
	zyncDefaultRailsConsoleEnabled    bool                               = false
	zyncDefaultConfigRailsEnvironment string                             = "development"
	zyncDefaultConfigRailsLogLevel    string                             = "info"
	zyncDefaultConfigRailsMaxThreads  int32                              = 10
	zyncDefaultConfigBugsnagSpec      BugsnagSpec                        = BugsnagSpec{}
	zyncDefaultAPIHPA                 defaultHorizontalPodAutoscalerSpec = defaultHorizontalPodAutoscalerSpec{
		MinReplicas:         ptr.To[int32](2),
		MaxReplicas:         ptr.To[int32](4),
		ResourceUtilization: ptr.To[int32](90),
		ResourceName:        ptr.To("cpu"),
	}
	zyncDefaultAPIPDB defaultPodDisruptionBudgetSpec = defaultPodDisruptionBudgetSpec{
		MaxUnavailable: ptr.To(intstr.FromInt(1)),
	}
	zyncDefaultAPIReplicas  int32                           = 2
	zyncDefaultAPIResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("250Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("750m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
	zyncDefaultAPILivenessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](10),
		TimeoutSeconds:      ptr.To[int32](30),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	zyncDefaultAPIReadinessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](30),
		TimeoutSeconds:      ptr.To[int32](10),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	zyncDefaultQueHPA defaultHorizontalPodAutoscalerSpec = defaultHorizontalPodAutoscalerSpec{
		MinReplicas:         ptr.To[int32](2),
		MaxReplicas:         ptr.To[int32](4),
		ResourceUtilization: ptr.To[int32](90),
		ResourceName:        ptr.To("cpu"),
	}
	zyncDefaultQuePDB defaultPodDisruptionBudgetSpec = defaultPodDisruptionBudgetSpec{
		MaxUnavailable: ptr.To(intstr.FromInt(1)),
	}
	zyncDefaultQueReplicas  int32                           = 2
	zyncDefaultQueResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("250Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("750m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
	zyncDefaultQueLivenessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](10),
		TimeoutSeconds:      ptr.To[int32](30),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	zyncDefaultQueReadinessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](30),
		TimeoutSeconds:      ptr.To[int32](10),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	zyncDefaultRailsConsoleResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("400m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
)

// ZyncSpec defines the desired state of Zync
type ZyncSpec struct {
	// Image specification for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// Application specific configuration options for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Config ZyncConfig `json:"config"`
	// Configures the Grafana Dashboard for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	GrafanaDashboard *GrafanaDashboardSpec `json:"grafanaDashboard,omitempty"`
	// Configures the main zync api component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	API *APISpec `json:"api,omitempty"`
	// Configures the zync que component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Que *QueSpec `json:"que,omitempty"`
	// Console specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Console *ZyncRailsConsoleSpec `json:"console,omitempty"`
}

// Default implements defaulting for ZyncSpec
func (spec *ZyncSpec) Default() {
	spec.Image = InitializeImageSpec(spec.Image, zyncDefaultImage)
	spec.Config.Default()

	if spec.API == nil {
		spec.API = &APISpec{}
	}

	spec.API.Default()

	if spec.Que == nil {
		spec.Que = &QueSpec{}
	}

	spec.Que.Default()
	spec.GrafanaDashboard = InitializeGrafanaDashboardSpec(spec.GrafanaDashboard, zyncDefaultGrafanaDashboard)

	if spec.Console == nil {
		spec.Console = &ZyncRailsConsoleSpec{}
	}

	spec.Console.Default(spec.Image)
}

// APISpec is the configuration for main Zync api component
type APISpec struct {
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

// Default implements defaulting for the each main zync api component
func (spec *APISpec) Default() {
	spec.HPA = InitializeHorizontalPodAutoscalerSpec(spec.HPA, zyncDefaultAPIHPA)
	spec.Replicas = intOrDefault(spec.Replicas, &zyncDefaultAPIReplicas)
	spec.PDB = InitializePodDisruptionBudgetSpec(spec.PDB, zyncDefaultAPIPDB)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, zyncDefaultAPIResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, zyncDefaultAPILivenessProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, zyncDefaultAPIReadinessProbe)
}

// QueSpec is the configuration for Zync que
type QueSpec struct {
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
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
}

// Default implements defaulting for the each zync que
func (spec *QueSpec) Default() {
	spec.HPA = InitializeHorizontalPodAutoscalerSpec(spec.HPA, zyncDefaultQueHPA)
	spec.Replicas = intOrDefault(spec.Replicas, &zyncDefaultQueReplicas)
	spec.PDB = InitializePodDisruptionBudgetSpec(spec.PDB, zyncDefaultQuePDB)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, zyncDefaultQueResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, zyncDefaultQueLivenessProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, zyncDefaultQueReadinessProbe)
}

// ZyncConfig configures app behavior for Zync
type ZyncConfig struct {
	// Rails configuration options for zync components
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Rails *ZyncRailsSpec `json:"rails,omitempty"`
	// External Secret common configuration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ExternalSecret ExternalSecret `json:"externalSecret,omitempty"`
	// A reference to the secret holding the database DSN
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DatabaseDSN SecretReference `json:"databaseDSN"`
	// A reference to the secret holding the secret-key-base
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecretKeyBase SecretReference `json:"secretKeyBase"`
	// A reference to the secret holding the zync authentication token
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ZyncAuthToken SecretReference `json:"zyncAuthToken"`
	// Options for configuring Bugsnag integration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Bugsnag *BugsnagSpec `json:"bugsnag,omitempty"`
}

// Default sets default values for any value not specifically set in the ZyncConfig struct
func (cfg *ZyncConfig) Default() {
	if cfg.Rails == nil {
		cfg.Rails = &ZyncRailsSpec{}
	}

	if cfg.Bugsnag == nil {
		cfg.Bugsnag = &zyncDefaultConfigBugsnagSpec
	}

	cfg.ExternalSecret.SecretStoreRef = InitializeExternalSecretSecretStoreReferenceSpec(cfg.ExternalSecret.SecretStoreRef, defaultExternalSecretSecretStoreReference)
	cfg.ExternalSecret.RefreshInterval = durationOrDefault(cfg.ExternalSecret.RefreshInterval, &defaultExternalSecretRefreshInterval)
	cfg.Rails.Default()
}

// ZyncRailsSpec configures rails for system components
type ZyncRailsSpec struct {
	// Rails environment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Environment *string `json:"environment,omitempty"`
	// Rails log level (debug, info, warn, error, fatal or unknown)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=debug;info;warn;error;fatal;unknown
	// +optional
	LogLevel *string `json:"logLevel,omitempty"`
	// Rails max threads (only applies to api)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MaxThreads *int32 `json:"maxThreads,omitempty"`
}

// Default applies defaults for ZyncRailsSpec
func (zrs *ZyncRailsSpec) Default() {
	zrs.Environment = stringOrDefault(zrs.Environment, ptr.To(zyncDefaultConfigRailsEnvironment))
	zrs.LogLevel = stringOrDefault(zrs.LogLevel, ptr.To(zyncDefaultConfigRailsLogLevel))
	zrs.MaxThreads = intOrDefault(zrs.MaxThreads, ptr.To[int32](zyncDefaultConfigRailsMaxThreads))
}

// ZyncRailsConsoleSpec configures the Console component of Zync
type ZyncRailsConsoleSpec struct {
	// Enables or disables the Zync Console statefulset
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Enabled *bool `json:"enabled,omitempty"` // Image specification for the Console component.
	// Defaults to zync image if not defined.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// Resource requirements for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Resources *ResourceRequirementsSpec `json:"resources,omitempty"`
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
}

// Default implements defaulting for the Zync console component
func (spec *ZyncRailsConsoleSpec) Default(zyncDefaultImage *ImageSpec) {
	spec.Enabled = boolOrDefault(spec.Enabled, ptr.To(zyncDefaultRailsConsoleEnabled))
	spec.Image = InitializeImageSpec(spec.Image, defaultImageSpec(*zyncDefaultImage))
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, zyncDefaultRailsConsoleResources)
}

// ZyncStatus defines the observed state of Zync
type ZyncStatus struct {
	AggregatedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Zync is the Schema for the zyncs API
type Zync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZyncSpec   `json:"spec,omitempty"`
	Status ZyncStatus `json:"status,omitempty"`
}

// Default implements defaulting for the Zync resource
func (z *Zync) Default() {
	z.Spec.Default()
}

var _ reconciler.ObjectWithAppStatus = &Zync{}

func (d *Zync) GetStatus() any {
	return &d.Status
}

// +kubebuilder:object:root=true

// ZyncList contains a list of Zync
type ZyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zync{}, &ZyncList{})
}
