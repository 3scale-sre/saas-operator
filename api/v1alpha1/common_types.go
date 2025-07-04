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
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const (
	// Finalizer is the finalizer string for resoures in the saas group
	Finalizer string = "saas.3scale.net"
	// AnnotationsDomain is a common prefix for all "rollout triggering"
	// annotation keys
	AnnotationsDomain string = "saas.3scale.net"
	// The name for the default seed Secret
	DefaultSeedSecret string = "saas-seed"
)

var (
	defaultExternalSecretRefreshInterval      metav1.Duration                               = metav1.Duration{Duration: 60 * time.Second}
	defaultExternalSecretSecretStoreReference defaultExternalSecretSecretStoreReferenceSpec = defaultExternalSecretSecretStoreReferenceSpec{
		Name: ptr.To("vault-mgmt"),
		Kind: ptr.To("ClusterSecretStore"),
	}
)

// ImageSpec defines the image for the component
type ImageSpec struct {
	// Docker repository of the image
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Name *string `json:"name,omitempty"`
	// Image tag
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Tag *string `json:"tag,omitempty"`
	// Name of the Secret that holds quay.io credentials to access
	// the image repository
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PullSecretName *string `json:"pullSecretName,omitempty"`
	// Pull policy for the image
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PullPolicy *corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type defaultImageSpec struct {
	Name, Tag, PullSecretName *string
	PullPolicy                *corev1.PullPolicy
}

// Default sets default values for any value not specifically set in the ImageSpec struct
func (spec *ImageSpec) Default(def defaultImageSpec) {
	spec.Name = stringOrDefault(spec.Name, def.Name)
	spec.Tag = stringOrDefault(spec.Tag, def.Tag)
	spec.PullSecretName = stringOrDefault(spec.PullSecretName, def.PullSecretName)
	spec.PullPolicy = func() *corev1.PullPolicy {
		if spec.PullPolicy == nil {
			return def.PullPolicy
		}

		return spec.PullPolicy
	}()
}

// IsDeactivated true if the field is set with the deactivated value (empty struct)
func (spec *ImageSpec) IsDeactivated() bool { return false }

// InitializeImageSpec initializes a ImageSpec struct
func InitializeImageSpec(spec *ImageSpec, def defaultImageSpec) *ImageSpec {
	if spec == nil {
		newImage := &ImageSpec{}
		newImage.Default(def)

		return newImage
	}

	dcopy := spec.DeepCopy()
	dcopy.Default(def)

	return dcopy
}

// ProbeSpec specifies configuration for a probe
type ProbeSpec struct {
	// Number of seconds after the container has started before liveness probes are initiated
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which the probe times out
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
	// How often (in seconds) to perform the probe
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SuccessThreshold *int32 `json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	FailureThreshold *int32 `json:"failureThreshold,omitempty"`
}

type defaultProbeSpec struct {
	InitialDelaySeconds, TimeoutSeconds, PeriodSeconds,
	SuccessThreshold, FailureThreshold *int32
}

// Default sets default values for any value not specifically set in the ProbeSpec struct
func (spec *ProbeSpec) Default(def defaultProbeSpec) {
	spec.InitialDelaySeconds = intOrDefault(spec.InitialDelaySeconds, def.InitialDelaySeconds)
	spec.TimeoutSeconds = intOrDefault(spec.TimeoutSeconds, def.TimeoutSeconds)
	spec.PeriodSeconds = intOrDefault(spec.PeriodSeconds, def.PeriodSeconds)
	spec.SuccessThreshold = intOrDefault(spec.SuccessThreshold, def.SuccessThreshold)
	spec.FailureThreshold = intOrDefault(spec.FailureThreshold, def.FailureThreshold)
}

func (spec *ProbeSpec) IsDeactivated() bool {
	return reflect.DeepEqual(spec, &ProbeSpec{})
}

// InitializeProbeSpec initializes a ProbeSpec struct
func InitializeProbeSpec(spec *ProbeSpec, def defaultProbeSpec) *ProbeSpec {
	if spec == nil {
		nreProbe := &ProbeSpec{}
		nreProbe.Default(def)

		return nreProbe
	}

	if !spec.IsDeactivated() {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}

	return spec
}

// ElasticLoadBalancerSpec configures the AWS load balancer for the component
type ElasticLoadBalancerSpec struct {
	// Enables/disbles use of proxy protocol in the load balancer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ProxyProtocol *bool `json:"proxyProtocol,omitempty"`
	// Enables/disables cross zone load balancing
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	CrossZoneLoadBalancingEnabled *bool `json:"crossZoneLoadBalancingEnabled,omitempty"`
	// Enables/disables connection draining
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ConnectionDrainingEnabled *bool `json:"connectionDrainingEnabled,omitempty"`
	// Sets the timeout for connection draining
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ConnectionDrainingTimeout *int32 `json:"connectionDrainingTimeout,omitempty"`
	// Sets the healthy threshold for the load balancer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	HealthcheckHealthyThreshold *int32 `json:"healthcheckHealthyThreshold,omitempty"`
	// Sets the unhealthy threshold for the load balancer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	HealthcheckUnhealthyThreshold *int32 `json:"healthcheckUnhealthyThreshold,omitempty"`
	// Sets the interval between health checks
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	HealthcheckInterval *int32 `json:"healthcheckInterval,omitempty"`
	// Sets the timeout for the health check
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	HealthcheckTimeout *int32 `json:"healthcheckTimeout,omitempty"`
}

var DefaultElasticLoadBalancerSpec ElasticLoadBalancerSpec = ElasticLoadBalancerSpec{
	ProxyProtocol:                 ptr.To(true),
	CrossZoneLoadBalancingEnabled: ptr.To(true),
	ConnectionDrainingEnabled:     ptr.To(true),
	ConnectionDrainingTimeout:     ptr.To[int32](60),
	HealthcheckHealthyThreshold:   ptr.To[int32](2),
	HealthcheckUnhealthyThreshold: ptr.To[int32](2),
	HealthcheckInterval:           ptr.To[int32](5),
	HealthcheckTimeout:            ptr.To[int32](3),
}

// Default sets default values for any value not specifically set in the LoadBalancerSpec struct
func (spec *ElasticLoadBalancerSpec) Default(def ElasticLoadBalancerSpec) {
	spec.ProxyProtocol = boolOrDefault(spec.ProxyProtocol, def.ProxyProtocol)
	spec.CrossZoneLoadBalancingEnabled = boolOrDefault(spec.CrossZoneLoadBalancingEnabled, def.CrossZoneLoadBalancingEnabled)
	spec.ConnectionDrainingEnabled = boolOrDefault(spec.ConnectionDrainingEnabled, def.ConnectionDrainingEnabled)
	spec.ConnectionDrainingTimeout = intOrDefault(spec.ConnectionDrainingTimeout, def.ConnectionDrainingTimeout)
	spec.HealthcheckHealthyThreshold = intOrDefault(spec.HealthcheckHealthyThreshold, def.HealthcheckHealthyThreshold)
	spec.HealthcheckUnhealthyThreshold = intOrDefault(spec.HealthcheckUnhealthyThreshold, def.HealthcheckUnhealthyThreshold)
	spec.HealthcheckInterval = intOrDefault(spec.HealthcheckInterval, def.HealthcheckInterval)
	spec.HealthcheckTimeout = intOrDefault(spec.HealthcheckTimeout, def.HealthcheckTimeout)
}

// InitializeElasticLoadBalancerSpec initializes a LoadBalancerSpec struct
func InitializeElasticLoadBalancerSpec(spec *ElasticLoadBalancerSpec, def ElasticLoadBalancerSpec) *ElasticLoadBalancerSpec {
	if spec == nil {
		newELB := &ElasticLoadBalancerSpec{}
		newELB.Default(def)

		return newELB
	} else {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}
}

// NetworkLoadBalancerSpec configures the AWS NLB load balancer for the component
type NetworkLoadBalancerSpec struct {
	// Enables/disbles use of proxy protocol in the load balancer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ProxyProtocol *bool `json:"proxyProtocol,omitempty"`
	// Enables/disables cross zone load balancing
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	CrossZoneLoadBalancingEnabled *bool `json:"crossZoneLoadBalancingEnabled,omitempty"`
	// The list of optional Elastic IPs allocations
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	EIPAllocations []string `json:"eipAllocations,omitempty"`
	// Optionally specify the load balancer name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	LoadBalancerName *string `json:"loadBalancerName,omitempty"`
	// Deletion protection setting
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	DeletionProtection *bool `json:"deletionProtection,omitempty"`
}

var DefaultNetworkLoadBalancerSpec NetworkLoadBalancerSpec = NetworkLoadBalancerSpec{
	ProxyProtocol:                 ptr.To(true),
	CrossZoneLoadBalancingEnabled: ptr.To(true),
	DeletionProtection:            ptr.To(false),
}

// Default sets default values for any value not specifically set in the NLBLoadBalancerSpec struct
func (spec *NetworkLoadBalancerSpec) Default(def NetworkLoadBalancerSpec) {
	spec.ProxyProtocol = boolOrDefault(spec.ProxyProtocol, def.ProxyProtocol)
	spec.CrossZoneLoadBalancingEnabled = boolOrDefault(spec.CrossZoneLoadBalancingEnabled, def.CrossZoneLoadBalancingEnabled)
	spec.DeletionProtection = boolOrDefault(spec.DeletionProtection, def.DeletionProtection)
}

// InitializeNetworkLoadBalancerSpec initializes a NLBLoadBalancerSpec struct
func InitializeNetworkLoadBalancerSpec(spec *NetworkLoadBalancerSpec, def NetworkLoadBalancerSpec) *NetworkLoadBalancerSpec {
	if spec == nil {
		newNLB := &NetworkLoadBalancerSpec{}
		newNLB.Default(def)

		return newNLB
	} else {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}
}

// GrafanaDashboardSpec configures the Grafana Dashboard for the component
type GrafanaDashboardSpec struct {
	// Label key used by grafana-operator for dashboard discovery
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SelectorKey *string `json:"selectorKey,omitempty"`
	// Label value used by grafana-operator for dashboard discovery
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SelectorValue *string `json:"selectorValue,omitempty"`
}

type defaultGrafanaDashboardSpec struct {
	SelectorKey, SelectorValue *string
}

// Default sets default values for any value not specifically set in the GrafanaDashboardSpec struct
func (spec *GrafanaDashboardSpec) Default(def defaultGrafanaDashboardSpec) {
	spec.SelectorKey = stringOrDefault(spec.SelectorKey, def.SelectorKey)
	spec.SelectorValue = stringOrDefault(spec.SelectorValue, def.SelectorValue)
}

// IsDeactivated true if the field is set with the deactivated value (empty struct)
func (spec *GrafanaDashboardSpec) IsDeactivated() bool {
	return reflect.DeepEqual(spec, &GrafanaDashboardSpec{})
}

// InitializeGrafanaDashboardSpec initializes a GrafanaDashboardSpec struct
func InitializeGrafanaDashboardSpec(spec *GrafanaDashboardSpec, def defaultGrafanaDashboardSpec) *GrafanaDashboardSpec {
	if spec == nil {
		newDashboard := &GrafanaDashboardSpec{}
		newDashboard.Default(def)

		return newDashboard
	}

	if !spec.IsDeactivated() {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}

	return spec
}

// Endpoint sets the external endpoint for the component
type Endpoint struct {
	// The list of dns records that will point to the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DNS []string `json:"dns"`
}

// PodDisruptionBudgetSpec defines the PDB for the component
type PodDisruptionBudgetSpec struct {
	// An eviction is allowed if at least "minAvailable" pods selected by
	// "selector" will still be available after the eviction, i.e. even in the
	// absence of the evicted pod.  So for example you can prevent all voluntary
	// evictions by specifying "100%".
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MinAvailable *intstr.IntOrString `json:"minAvailable,omitempty"`
	// An eviction is allowed if at most "maxUnavailable" pods selected by
	// "selector" are unavailable after the eviction, i.e. even in absence of
	// the evicted pod. For example, one can prevent all voluntary evictions
	// by specifying 0. This is a mutually exclusive setting with "minAvailable".
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type defaultPodDisruptionBudgetSpec struct {
	MinAvailable, MaxUnavailable *intstr.IntOrString
}

// Default sets default values for any value not specifically set in the PodDisruptionBudgetSpec struct
func (spec *PodDisruptionBudgetSpec) Default(def defaultPodDisruptionBudgetSpec) {
	if spec.MinAvailable == nil && spec.MaxUnavailable == nil {
		if def.MinAvailable != nil {
			spec.MinAvailable = def.MinAvailable
			spec.MaxUnavailable = nil
		} else if def.MaxUnavailable != nil {
			spec.MinAvailable = nil
			spec.MaxUnavailable = def.MaxUnavailable
		}
	}
}

// IsDeactivated true if the field is set with the deactivated value (empty struct)
func (spec *PodDisruptionBudgetSpec) IsDeactivated() bool {
	return reflect.DeepEqual(spec, &PodDisruptionBudgetSpec{})
}

// InitializePodDisruptionBudgetSpec initializes a PodDisruptionBudgetSpec struct
func InitializePodDisruptionBudgetSpec(spec *PodDisruptionBudgetSpec, def defaultPodDisruptionBudgetSpec) *PodDisruptionBudgetSpec {
	if spec == nil {
		newPDB := &PodDisruptionBudgetSpec{}
		newPDB.Default(def)

		return newPDB
	}

	if !spec.IsDeactivated() {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}

	return spec
}

// HorizontalPodAutoscalerSpec defines the HPA for the component
type HorizontalPodAutoscalerSpec struct {
	// Lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// Upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// Target resource used to autoscale (cpu/memory)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=cpu;memory
	// +optional
	ResourceName *string `json:"resourceName,omitempty"`
	// A percentage indicating the target resource consumption used to autoscale
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ResourceUtilization *int32 `json:"resourceUtilization,omitempty"`
	// Behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

type defaultHorizontalPodAutoscalerSpec struct {
	MinReplicas, MaxReplicas, ResourceUtilization *int32
	ResourceName                                  *string
}

// Default sets default values for any value not specifically set in the PodDisruptionBudgetSpec struct
func (spec *HorizontalPodAutoscalerSpec) Default(def defaultHorizontalPodAutoscalerSpec) {
	spec.MinReplicas = intOrDefault(spec.MinReplicas, def.MinReplicas)
	spec.MaxReplicas = intOrDefault(spec.MaxReplicas, def.MaxReplicas)
	spec.ResourceName = stringOrDefault(spec.ResourceName, def.ResourceName)
	spec.ResourceUtilization = intOrDefault(spec.ResourceUtilization, def.ResourceUtilization)
}

// IsDeactivated true if the field is set with the deactivated value (empty struct)
func (spec *HorizontalPodAutoscalerSpec) IsDeactivated() bool {
	return reflect.DeepEqual(spec, &HorizontalPodAutoscalerSpec{})
}

// InitializeHorizontalPodAutoscalerSpec initializes a HorizontalPodAutoscalerSpec struct
func InitializeHorizontalPodAutoscalerSpec(spec *HorizontalPodAutoscalerSpec, def defaultHorizontalPodAutoscalerSpec) *HorizontalPodAutoscalerSpec {
	if spec == nil {
		newHPA := &HorizontalPodAutoscalerSpec{}
		newHPA.Default(def)

		return newHPA
	}

	if !spec.IsDeactivated() {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}

	return spec
}

type DeploymentStrategySpec struct {
	// Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type appsv1.DeploymentStrategyType `json:"type,omitempty"`
	// Rolling update config params. Present only if DeploymentStrategyType =
	// RollingUpdate.
	// +optional
	RollingUpdate *appsv1.RollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

type defaultDeploymentRollingStrategySpec struct {
	MaxUnavailable, MaxSurge *intstr.IntOrString
}

// InitializeDeploymentStrategySpec initializes a DeploymentStrategySpec struct
func InitializeDeploymentStrategySpec(spec *DeploymentStrategySpec, def defaultDeploymentRollingStrategySpec) *DeploymentStrategySpec {
	if spec == nil {
		newDeploymentStrategy := &DeploymentStrategySpec{}
		newDeploymentStrategy.Default(def)

		return newDeploymentStrategy
	}

	return spec
}

// Default sets default values for any value not specifically set in the DeploymentStrategySpec struct
func (spec *DeploymentStrategySpec) Default(def defaultDeploymentRollingStrategySpec) {
	spec.Type = appsv1.RollingUpdateDeploymentStrategyType
	spec.RollingUpdate = &appsv1.RollingUpdateDeployment{
		MaxSurge:       def.MaxSurge,
		MaxUnavailable: def.MaxUnavailable,
	}
}

// ResourceRequirementsSpec defines the resource requirements for the component
type ResourceRequirementsSpec struct {
	// Limits describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Limits corev1.ResourceList `json:"limits,omitempty"`
	// Requests describes the minimum amount of compute resources required.
	// If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
	// Claims lists the names of resources, defined in spec.resourceClaims,
	// that are used by this container.
	//
	// This is an alpha field and requires enabling the
	// DynamicResourceAllocation feature gate.
	//
	// This field is immutable.
	//
	// +listType=map
	// +listMapKey=name
	// +featureGate=DynamicResourceAllocation
	// +optional
	Claims []corev1.ResourceClaim `json:"claims,omitempty" protobuf:"bytes,3,opt,name=claims"`
}

type defaultResourceRequirementsSpec struct {
	Limits, Requests corev1.ResourceList
}

// Default sets default values for any value not specifically set in the ResourceRequirementsSpec struct
func (spec *ResourceRequirementsSpec) Default(def defaultResourceRequirementsSpec) {
	if spec.Requests == nil {
		spec.Requests = def.Requests
	}

	if spec.Limits == nil {
		spec.Limits = def.Limits
	}
}

// IsDeactivated true if the field is set with the deactivated value (empty struct)
func (spec *ResourceRequirementsSpec) IsDeactivated() bool {
	return reflect.DeepEqual(spec, &ResourceRequirementsSpec{})
}

// InitializeResourceRequirementsSpec initializes a ResourceRequirementsSpec struct
func InitializeResourceRequirementsSpec(spec *ResourceRequirementsSpec, def defaultResourceRequirementsSpec) *ResourceRequirementsSpec {
	if spec == nil {
		newResources := &ResourceRequirementsSpec{}
		newResources.Default(def)

		return newResources
	}

	if !spec.IsDeactivated() {
		dcopy := spec.DeepCopy()
		dcopy.Default(def)

		return dcopy
	}

	return spec
}

// ExternalSecret is a reference to the ExternalSecret common configuration
type ExternalSecret struct {
	// SecretStoreRef defines which SecretStore to use when fetching the secret data
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SecretStoreRef *ExternalSecretSecretStoreReferenceSpec `json:"secretStoreRef,omitempty"`
	// RefreshInterval is the amount of time before the values reading again from the SecretStore provider (duration)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	RefreshInterval *metav1.Duration `json:"refreshInterval,omitempty"`
}

func (spec *ExternalSecret) Default() {
	spec.SecretStoreRef = InitializeExternalSecretSecretStoreReferenceSpec(spec.SecretStoreRef, defaultExternalSecretSecretStoreReference)
	spec.RefreshInterval = durationOrDefault(spec.RefreshInterval, &defaultExternalSecretRefreshInterval)
}

// SecretReference is a reference to a secret stored in some secrets engine
type SecretReference struct {
	// FromVault is a reference to a secret key/value stored in a Hashicorp Vault
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	FromVault *VaultSecretReference `json:"fromVault,omitempty"`
	// Override allows to directly specify a string value.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Override *string `json:"override,omitempty"`
	// FromSeed will try to retrieve the secret value from
	// the default seed Secret.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	FromSeed *SeedSecretReference `json:"fromSeed,omitempty"`
}

// VaultSecretReference is a reference to a secret stored in
// a Hashicorp Vault
type VaultSecretReference struct {
	// The Vault path where the secret is located
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Path string `json:"path"`
	// The Vault key of the secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Key string `json:"key"`
}

func (spec *VaultSecretReference) Default() {}

// SeedSecretReference represents options to
// retrieve the secret value from the default seed Secret.
// There are no configurable options at this point.
type SeedSecretReference struct{}

// ExternalSecretSecretStoreReferenceSpec is a reference to a secret store
type ExternalSecretSecretStoreReferenceSpec struct {
	// The Vault secret store reference name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Name *string `json:"name,omitempty"`
	// The Vault secret store reference kind
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Kind *string `json:"kind,omitempty"`
}

type defaultExternalSecretSecretStoreReferenceSpec struct {
	Name, Kind *string
}

// Default sets default values for any value not specifically set in the ExternalSecretSecretStoreReferenceSpec struct
func (spec *ExternalSecretSecretStoreReferenceSpec) Default(def defaultExternalSecretSecretStoreReferenceSpec) {
	spec.Name = stringOrDefault(spec.Name, def.Name)
	spec.Kind = stringOrDefault(spec.Kind, def.Kind)
}

// InitializeExternalSecretSecretStoreReferenceSpec initializes a ExternalSecretSecretStoreReferenceSpec struct
func InitializeExternalSecretSecretStoreReferenceSpec(spec *ExternalSecretSecretStoreReferenceSpec, def defaultExternalSecretSecretStoreReferenceSpec) *ExternalSecretSecretStoreReferenceSpec {
	if spec == nil {
		newSecretStore := &ExternalSecretSecretStoreReferenceSpec{}
		newSecretStore.Default(def)

		return newSecretStore
	}

	dcopy := spec.DeepCopy()
	dcopy.Default(def)

	return dcopy
}

// BugsnagSpec has configuration for Bugsnag integration
type BugsnagSpec struct {
	// Release Stage to identify environment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ReleaseStage *string `json:"releaseStage,omitempty"`
	// API key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	APIKey SecretReference `json:"apiKey"`
}

// Enabled returns a boolean indication whether the
// Bugsnag integration is enabled or not
func (bs *BugsnagSpec) Enabled() bool {
	return !reflect.DeepEqual(bs, &BugsnagSpec{})
}

// AddressSpec allows the definition of an address
type AddressSpec struct {
	// Defines the address host
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Host *string `json:"host,omitempty"`
	// Defines the address port
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Port *int32 `json:"port,omitempty"`
}

// Canary allows the definition of a canary Deployment
type Canary struct {
	// SendTraffic controls if traffic is sent to the canary
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SendTraffic bool `json:"sendTraffic"`
	// ImageName to use for the canary Deployment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ImageName *string `json:"imageName,omitempty"`
	// ImageTag to use for the canary Deployment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ImageTag *string `json:"imageTag,omitempty"`
	// Number of replicas for the canary Deployment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Patches to apply for the canary Deployment. Patches are expected
	// to be JSON documents as an RFC 6902 patches.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Patches []string `json:"patches,omitempty"`
}

// PatchSpec returns a modified spec given the canary configuration
func (c *Canary) PatchSpec(spec, canarySpec any) error {
	doc, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("unable to marshal spec: '%s'", err.Error())
	}

	for _, p := range c.Patches {
		patch, err := jsonpatch.DecodePatch([]byte(p))
		if err != nil {
			return fmt.Errorf("unable to decode canary patch: '%s'", err.Error())
		}

		doc, err = patch.Apply(doc)
		if err != nil {
			return fmt.Errorf("unable to apply canary patch: '%s'", err.Error())
		}
	}

	if err := json.Unmarshal(doc, canarySpec); err != nil {
		return fmt.Errorf("unable to unmarshal spec: '%s'", err.Error())
	}

	return nil
}

type WorkloadStatus struct {
	// HealthStatus holds the status of the individual workload
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	HealthStatus string `json:"healthStatus,omitempty"`
	// HealthMessage holds the message describing the health status
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	HealthMessage string `json:"healthMessage,omitempty"`
	// DeploymentStatus is a copy of the status of the owned Deployment
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	DeploymentStatus *appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`
	// StatefulSetStatus is a copy of the status of the owned Deployment
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	StatefulSetStatus *appsv1.StatefulSetStatus `json:"statefulsetStatus,omitempty"`
}

var _ reconciler.AppStatusWithAggregatedHealth = &AggregatedStatus{}

type AggregatedStatus struct {
	// Health is the overall health of the custom resource
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	Health string `json:"health,omitempty"`
	// OwnedWorkloads is a map with the health statuses of individual owned workloads
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	OwnedWorkloads map[string]*WorkloadStatus `json:"ownedWorkloads,omitempty"`
}

func (status *AggregatedStatus) Init(key types.NamespacedName) {
	if status.OwnedWorkloads == nil {
		status.OwnedWorkloads = map[string]*WorkloadStatus{}
	}

	if _, ok := status.OwnedWorkloads[key.Name]; !ok {
		status.OwnedWorkloads[key.Name] = &WorkloadStatus{
			HealthStatus:  "Unknown",
			HealthMessage: "Unable to determine health",
		}
	}
}

func (status *AggregatedStatus) GetDeploymentStatus(key types.NamespacedName) *appsv1.DeploymentStatus {
	if w, ok := status.OwnedWorkloads[key.Name]; !ok {
		return nil
	} else {
		return w.DeploymentStatus
	}
}

func (status *AggregatedStatus) SetDeploymentStatus(key types.NamespacedName, s *appsv1.DeploymentStatus) {
	status.Init(key)
	status.OwnedWorkloads[key.Name].DeploymentStatus = s
}

func (status *AggregatedStatus) GetStatefulSetStatus(key types.NamespacedName) *appsv1.StatefulSetStatus {
	if w, ok := status.OwnedWorkloads[key.Name]; !ok {
		return nil
	} else {
		return w.StatefulSetStatus
	}
}
func (status *AggregatedStatus) SetStatefulSetStatus(key types.NamespacedName, s *appsv1.StatefulSetStatus) {
	status.Init(key)
	status.OwnedWorkloads[key.Name].StatefulSetStatus = s
}

func (status *AggregatedStatus) GetHealthStatus(key types.NamespacedName) string {
	if w, ok := status.OwnedWorkloads[key.Name]; !ok {
		return "Unknown"
	} else {
		return w.HealthStatus
	}
}

func (status *AggregatedStatus) SetHealthStatus(key types.NamespacedName, s string) {
	status.Init(key)
	status.OwnedWorkloads[key.Name].HealthStatus = s
}

func (status *AggregatedStatus) GetHealthMessage(key types.NamespacedName) string {
	if w, ok := status.OwnedWorkloads[key.Name]; !ok {
		return "Unable to determine health"
	} else {
		return w.HealthMessage
	}
}

func (status *AggregatedStatus) SetHealthMessage(key types.NamespacedName, msg string) {
	status.Init(key)
	status.OwnedWorkloads[key.Name].HealthMessage = msg
}

func (status *AggregatedStatus) GetAggregatedHealthStatus() string {
	if status.Health == "" {
		return "Unknown"
	} else {
		return status.Health
	}
}

func (status *AggregatedStatus) SetAggregatedHealthStatus(s string) {
	status.Health = s
}

func stringOrDefault(value *string, defValue *string) *string {
	if value == nil {
		return defValue
	}

	return value
}

func stringSliceOrDefault(value []string, defValue []string) []string {
	if len(value) == 0 {
		return defValue
	}

	return value
}

func intOrDefault(value *int32, defValue *int32) *int32 {
	if value == nil {
		return defValue
	}

	return value
}

func int64OrDefault(value *int64, defValue *int64) *int64 {
	if value == nil {
		return defValue
	}

	return value
}

func boolOrDefault(value *bool, defValue *bool) *bool {
	if value == nil {
		return defValue
	}

	return value
}

func durationOrDefault(value *metav1.Duration, defValue *metav1.Duration) *metav1.Duration {
	if value == nil {
		return defValue
	}

	return value
}
