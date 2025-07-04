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
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type systemSidekiqType string

const (
	Default systemSidekiqType = "default"
	Billing systemSidekiqType = "billing"
	Low     systemSidekiqType = "low"
)

var (

	// Common
	systemDefaultSandboxProxyOpensslVerifyMode string           = "VERIFY_NONE"
	systemDefaultForceSSL                      bool             = true
	systemDefaultSSLCertsDir                   string           = "/etc/pki/tls/certs"
	systemDefaultThreescaleProviderPlan        string           = "enterprise"
	systemDefaultThreescaleSuperdomain         string           = "localhost"
	systemDefaultRailsConsole                  bool             = false
	systemDefaultRailsEnvironment              string           = "preview"
	systemDefaultRailsLogLevel                 string           = "info"
	systemDefaultConfigFilesSecret             string           = "system-config"
	systemDefaultBugsnagSpec                   BugsnagSpec      = BugsnagSpec{}
	systemDefaultImage                         defaultImageSpec = defaultImageSpec{
		Name:       ptr.To("quay.io/3scale/porta"),
		Tag:        ptr.To("nightly"),
		PullPolicy: (*corev1.PullPolicy)(ptr.To(string(corev1.PullIfNotPresent))),
	}
	systemDefaultGrafanaDashboard defaultGrafanaDashboardSpec = defaultGrafanaDashboardSpec{
		SelectorKey:   ptr.To("monitoring-key"),
		SelectorValue: ptr.To("middleware"),
	}
	systemDefaultTerminationGracePeriodSeconds *int64           = ptr.To[int64](60)
	systemDefaultSearchServer                  SearchServerSpec = SearchServerSpec{
		AddressSpec: AddressSpec{
			Host: ptr.To("system-searchd"),
			Port: ptr.To[int32](9306),
		},
		BatchSize: ptr.To[int32](100),
	}

	// App
	systemDefaultAppReplicas  int32                           = 2
	systemDefaultAppResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("400m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
	systemDefaultAppDeploymentStrategy defaultDeploymentRollingStrategySpec = defaultDeploymentRollingStrategySpec{
		MaxUnavailable: ptr.To(intstr.FromInt(0)),
		MaxSurge:       ptr.To(intstr.FromString("10%")),
	}
	systemDefaultAppHPA defaultHorizontalPodAutoscalerSpec = defaultHorizontalPodAutoscalerSpec{
		MinReplicas:         ptr.To[int32](2),
		MaxReplicas:         ptr.To[int32](4),
		ResourceUtilization: ptr.To[int32](90),
		ResourceName:        ptr.To("cpu"),
	}
	systemDefaultAppLivenessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](30),
		TimeoutSeconds:      ptr.To[int32](1),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	systemDefaultAppReadinessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](30),
		TimeoutSeconds:      ptr.To[int32](5),
		PeriodSeconds:       ptr.To[int32](10),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](3),
	}
	systemDefaultAppPDB defaultPodDisruptionBudgetSpec = defaultPodDisruptionBudgetSpec{
		MaxUnavailable: ptr.To(intstr.FromInt(1)),
	}

	// Sidekiq
	systemDefaultSidekiqReplicas  int32                           = 2
	systemDefaultSidekiqResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
	systemDefaultSidekiqDeploymentStrategy defaultDeploymentRollingStrategySpec = defaultDeploymentRollingStrategySpec{
		MaxUnavailable: ptr.To(intstr.FromInt(0)),
		MaxSurge:       ptr.To(intstr.FromInt(1)),
	}
	systemDefaultSidekiqHPA defaultHorizontalPodAutoscalerSpec = defaultHorizontalPodAutoscalerSpec{
		MinReplicas:         ptr.To[int32](2),
		MaxReplicas:         ptr.To[int32](4),
		ResourceUtilization: ptr.To[int32](90),
		ResourceName:        ptr.To("cpu"),
	}
	systemDefaultSidekiqLivenessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](10),
		TimeoutSeconds:      ptr.To[int32](3),
		PeriodSeconds:       ptr.To[int32](15),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](5),
	}
	systemDefaultSidekiqReadinessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](10),
		TimeoutSeconds:      ptr.To[int32](5),
		PeriodSeconds:       ptr.To[int32](30),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](5),
	}
	systemDefaultSidekiqPDB defaultPodDisruptionBudgetSpec = defaultPodDisruptionBudgetSpec{
		MaxUnavailable: ptr.To(intstr.FromInt(1)),
	}

	systemDefaultSidekiqConfigDefault defaultSidekiqConfig = defaultSidekiqConfig{
		Queues: []string{
			"critical", "backend_sync", "events", "zync,40",
			"priority,25", "default,15", "web_hooks,10", "deletion,5",
		},
		MaxThreads: ptr.To[int32](15),
	}
	systemDefaultSidekiqConfigBilling defaultSidekiqConfig = defaultSidekiqConfig{
		Queues:     []string{"billing"},
		MaxThreads: ptr.To[int32](15),
	}
	systemDefaultSidekiqConfigLow defaultSidekiqConfig = defaultSidekiqConfig{
		Queues: []string{
			"mailers", "low", "bulk_indexing",
		},
		MaxThreads: ptr.To[int32](15),
	}

	// Searchd
	systemDefaultSearchdEnabled bool             = true
	systemDefaultSearchdImage   defaultImageSpec = defaultImageSpec{
		Name:       ptr.To("quay.io/3scale/searchd"),
		Tag:        ptr.To("latest"),
		PullPolicy: (*corev1.PullPolicy)(ptr.To(string(corev1.PullIfNotPresent))),
	}
	systemDefaultSearchdServiceName         string                          = "system-searchd"
	systemDefaultSearchdPort                int32                           = 9306
	systemDefaultSearchdDBPath              string                          = "/var/lib/searchd"
	systemDefaultSearchdDatabaseStorageSize string                          = "30Gi"
	systemDefaultSearchdResources           defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("750m"),
			corev1.ResourceMemory: resource.MustParse("5Gi"),
		},
	}
	systemDefaultSearchdLivenessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](60),
		TimeoutSeconds:      ptr.To[int32](3),
		PeriodSeconds:       ptr.To[int32](15),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](5),
	}
	systemDefaultSearchdReadinessProbe defaultProbeSpec = defaultProbeSpec{
		InitialDelaySeconds: ptr.To[int32](60),
		TimeoutSeconds:      ptr.To[int32](5),
		PeriodSeconds:       ptr.To[int32](30),
		SuccessThreshold:    ptr.To[int32](1),
		FailureThreshold:    ptr.To[int32](5),
	}
	systemDefaultRailsConsoleResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("400m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
	systemDefaultSystemTektonTasks []SystemTektonTaskSpec = []SystemTektonTaskSpec{
		{
			Name:        ptr.To("system-backend-sync"),
			Description: ptr.To("Runs the Backend Synchronization task"),
			Config: &SystemTektonTaskConfig{
				Command: []string{"container-entrypoint"},
				Args: []string{
					"bundle",
					"exec",
					"rails",
					"backend:storage:enqueue_rewrite",
				},
			},
		},
		{
			Name:        ptr.To("system-db-migrate"),
			Description: ptr.To("Runs the Database Migration task"),
			Config: &SystemTektonTaskConfig{
				Command: []string{"container-entrypoint"},
				Args: []string{
					"bundle",
					"exec",
					"rails",
					"db:migrate",
				},
			},
		},
		{
			Name:        ptr.To("system-searchd-reindex"),
			Description: ptr.To("Runs the Searchd Rendexation task"),
			Config: &SystemTektonTaskConfig{
				Command: []string{"container-entrypoint"},
				Args: []string{
					"bundle",
					"exec",
					"rake",
					"searchd:optimal_index",
				},
				ExtraEnv: []corev1.EnvVar{
					{
						Name: "THINKING_SPHINX_BATCH_SIZE", Value: "20",
					},
				},
			},
		},
	}
	systemDefaultSystemTektonTaskResources defaultResourceRequirementsSpec = defaultResourceRequirementsSpec{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("400m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
	systemDefaultSystemTektonTasksTimeout metav1.Duration = metav1.Duration{Duration: 3 * time.Hour}
)

// SystemSpec defines the desired state of System
type SystemSpec struct {
	// Application specific configuration options for System components
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Config SystemConfig `json:"config"`
	// Image specification for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// Application specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	App *SystemAppSpec `json:"app,omitempty"`
	// Sidekiq Default specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SidekiqDefault *SystemSidekiqSpec `json:"sidekiqDefault,omitempty"`
	// Sidekiq Billing specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SidekiqBilling *SystemSidekiqSpec `json:"sidekiqBilling,omitempty"`
	// Sidekiq Low specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SidekiqLow *SystemSidekiqSpec `json:"sidekiqLow,omitempty"`
	// Searchd specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Searchd *SystemSearchdSpec `json:"searchd,omitempty"`
	// Console specific configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Console *SystemRailsConsoleSpec `json:"console,omitempty"`
	// Configures the Tekton Tasks for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Tasks []SystemTektonTaskSpec `json:"tasks,omitempty"`
	// Configures the Grafana Dashboard for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	GrafanaDashboard *GrafanaDashboardSpec `json:"grafanaDashboard,omitempty"`
	// Configures twemproxy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Twemproxy *TwemproxySpec `json:"twemproxy,omitempty"`
}

// Default implements defaulting for SystemSpec
func (spec *SystemSpec) Default() {
	spec.Config.Default()
	spec.Image = InitializeImageSpec(spec.Image, systemDefaultImage)
	spec.GrafanaDashboard = InitializeGrafanaDashboardSpec(spec.GrafanaDashboard, systemDefaultGrafanaDashboard)

	if spec.App == nil {
		spec.App = &SystemAppSpec{}
	}

	spec.App.Default()

	if spec.SidekiqDefault == nil {
		spec.SidekiqDefault = &SystemSidekiqSpec{}
	}

	spec.SidekiqDefault.Default(Default)

	if spec.SidekiqBilling == nil {
		spec.SidekiqBilling = &SystemSidekiqSpec{}
	}

	spec.SidekiqBilling.Default(Billing)

	if spec.SidekiqLow == nil {
		spec.SidekiqLow = &SystemSidekiqSpec{}
	}

	spec.SidekiqLow.Default(Low)

	if spec.Searchd == nil {
		spec.Searchd = &SystemSearchdSpec{}
	}

	spec.Searchd.Default()

	if spec.Console == nil {
		spec.Console = &SystemRailsConsoleSpec{}
	}

	spec.Console.Default(spec.Image)

	if spec.Twemproxy != nil {
		spec.Twemproxy.Default()
	}

	for _, defaultTask := range systemDefaultSystemTektonTasks {
		defaultTaskFound := false

		// If a default task is defined, default missing information
		for t, resourceTask := range spec.Tasks {
			if *resourceTask.Name == *defaultTask.Name {
				spec.Tasks[t].Description = stringOrDefault(resourceTask.Description, defaultTask.Description)
				spec.Tasks[t].Enabled = boolOrDefault(resourceTask.Enabled, defaultTask.Enabled)
				spec.Tasks[t].Config.Merge(*defaultTask.Config)

				defaultTaskFound = true
			}
		}

		// Add the default task if missing
		if !defaultTaskFound {
			spec.Tasks = append(spec.Tasks, defaultTask)
		}
	}

	for i := range spec.Tasks {
		spec.Tasks[i].Default(spec.Image)
	}
}

type SearchServerSpec struct {
	AddressSpec `json:",inline"`
	// Defines the batch size
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	BatchSize *int32 `json:"batchSize,omitempty"`
}

// SystemConfig holds configuration for SystemApp component
type SystemConfig struct {
	// Rails configuration options for system components
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Rails *SystemRailsSpec `json:"rails,omitempty"`
	// OpenSSL verification mode for sandbox proxy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SandboxProxyOpensslVerifyMode *string `json:"sandboxProxyOpensslVerifyMode,omitempty"`
	// Enable (true) or disable (false) enforcing SSL
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ForceSSL *bool `json:"forceSSL,omitempty"`
	// SSL certificates path
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SSLCertsDir *string `json:"sslCertsDir,omitempty"`
	// Search service options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SearchServer SearchServerSpec `json:"searchServer,omitempty"`
	// 3scale provider plan
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ThreescaleProviderPlan *string `json:"threescaleProviderPlan,omitempty"`
	// 3scale superdomain
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ThreescaleSuperdomain *string `json:"threescaleSuperdomain,omitempty"`
	// Secret containging system configuration files to be mounted in the pods
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ConfigFilesSecret *string `json:"configFilesSecret,omitempty"`
	// External Secret common configuration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ExternalSecret ExternalSecret `json:"externalSecret,omitempty"`
	// DSN of system's main database
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DatabaseDSN SecretReference `json:"databaseDSN"`
	// EventsSharedSecret is a password that protects System's event
	// hooks endpoint.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	EventsSharedSecret SecretReference `json:"eventsSharedSecret"`
	// Holds recaptcha configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Recaptcha SystemRecaptchaSpec `json:"recaptcha"`
	// SecretKeyBase: https://api.rubyonrails.org/classes/Rails/Application.html#method-i-secret_key_base
	// You can generate one random key using 'bundle exec rake secret'
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecretKeyBase SecretReference `json:"secretKeyBase"`
	// AccessCode to protect admin urls
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AccessCode *SecretReference `json:"accessCode,omitempty"`
	// Options for Segment integration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Segment SegmentSpec `json:"segment"`
	// Options for Github integration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Github GithubSpec `json:"github"`
	// Options for configuring RH Customer Portal integration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	RedHatCustomerPortal RedHatCustomerPortalSpec `json:"redhatCustomerPortal"`
	// Options for configuring Bugsnag integration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Bugsnag *BugsnagSpec `json:"bugsnag,omitempty"`
	// DatabaseSecret is a site key stored off-database for improved more secure password hashing
	// See https://github.com/3scale/porta/blob/ae498814cef3d856613f60d29330882fa870271d/config/initializers/site_keys.rb#L2-L19
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DatabaseSecret SecretReference `json:"databaseSecret"`
	// Memcached servers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	MemcachedServers string `json:"memcachedServers"`
	// Redis configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Redis RedisSpec `json:"redis"`
	// SMTP configuration options
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SMTP SMTPSpec `json:"smtp"`
	// Mapping Service access token
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	MappingServiceAccessToken SecretReference `json:"mappingServiceAccessToken"`
	// Zync has configuration options for system to contact zync
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Zync SystemZyncSpec `json:"zync,omitempty"`
	// Backend has configuration options for system to contact backend
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Backend SystemBackendSpec `json:"backend"`
	// Assets has configuration to access assets in AWS s3
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Assets AssetsSpec `json:"assets"`
	// Apicast can be used to pass down apicast endpoints configuration
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Apicast *SystemApicastEndpointsSpec `json:"apicast,omitempty"`
}

// Default applies default values to a SystemConfig struct
func (sc *SystemConfig) Default() {
	if sc.Rails == nil {
		sc.Rails = &SystemRailsSpec{}
	}

	sc.Rails.Default()

	sc.ConfigFilesSecret = stringOrDefault(sc.ConfigFilesSecret, ptr.To(systemDefaultConfigFilesSecret))

	if sc.Bugsnag == nil {
		sc.Bugsnag = &systemDefaultBugsnagSpec
	}

	sc.SandboxProxyOpensslVerifyMode = stringOrDefault(sc.SandboxProxyOpensslVerifyMode, ptr.To(systemDefaultSandboxProxyOpensslVerifyMode))
	sc.ForceSSL = boolOrDefault(sc.ForceSSL, ptr.To(systemDefaultForceSSL))
	sc.SSLCertsDir = stringOrDefault(sc.SSLCertsDir, ptr.To(systemDefaultSSLCertsDir))
	sc.ThreescaleProviderPlan = stringOrDefault(sc.ThreescaleProviderPlan, ptr.To(systemDefaultThreescaleProviderPlan))
	sc.ThreescaleSuperdomain = stringOrDefault(sc.ThreescaleSuperdomain, ptr.To(systemDefaultThreescaleSuperdomain))
	sc.ExternalSecret.SecretStoreRef = InitializeExternalSecretSecretStoreReferenceSpec(sc.ExternalSecret.SecretStoreRef, defaultExternalSecretSecretStoreReference)
	sc.ExternalSecret.RefreshInterval = durationOrDefault(sc.ExternalSecret.RefreshInterval, &defaultExternalSecretRefreshInterval)

	sc.SearchServer.Host = stringOrDefault(sc.SearchServer.Host, systemDefaultSearchServer.Host)
	sc.SearchServer.Port = intOrDefault(sc.SearchServer.Port, systemDefaultSearchServer.Port)
	sc.SearchServer.BatchSize = intOrDefault(sc.SearchServer.BatchSize, systemDefaultSearchServer.BatchSize)
}

// ResolveCanarySpec modifies the SystemSpec given the provided canary configuration
func (spec *SystemSpec) ResolveCanarySpec(canary *Canary) (*SystemSpec, error) {
	canarySpec := &SystemSpec{}
	if err := canary.PatchSpec(spec, canarySpec); err != nil {
		return nil, err
	}

	if canary.ImageName != nil {
		canarySpec.Image.Name = canary.ImageName
	}

	if canary.ImageTag != nil {
		canarySpec.Image.Tag = canary.ImageTag
	}

	canarySpec.App.Replicas = canary.Replicas
	canarySpec.SidekiqDefault.Replicas = canary.Replicas
	canarySpec.SidekiqLow.Replicas = canary.Replicas
	canarySpec.SidekiqBilling.Replicas = canary.Replicas

	// Call Default() on the resolved canary spec to apply
	// defaulting to potentially added fields
	canarySpec.Default()

	return canarySpec, nil
}

// SystemRecaptchaSpec holds recaptcha configurations
type SystemRecaptchaSpec struct {
	// Public key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	PublicKey SecretReference `json:"publicKey"`
	// Private key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	PrivateKey SecretReference `json:"privateKey"`
}

// SegmentSpec has configuration for Segment integration
type SegmentSpec struct {
	// Deletion workspace
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DeletionWorkspace string `json:"deletionWorkspace"`
	// Deletion token
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	DeletionToken SecretReference `json:"deletionToken"`
	// Write key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	WriteKey SecretReference `json:"writeKey"`
}

// GithubSpec has configuration for Github integration
type GithubSpec struct {
	// Client ID
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClientID SecretReference `json:"clientID"`
	// Client secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClientSecret SecretReference `json:"clientSecret"`
}

// RedHatCustomerPortalSpec has configuration for integration with
// Red Hat Customer Portal
type RedHatCustomerPortalSpec struct {
	// Client ID
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClientID SecretReference `json:"clientID"`
	// Client secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClientSecret SecretReference `json:"clientSecret"`
	// Realm
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Realm *string `json:"realm,omitempty"`
}

// RedisSpec holds redis configuration
type RedisSpec struct {
	// Data source name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	QueuesDSN string `json:"queuesDSN"`
}

// SMTPSpec has options to configure system's SMTP
type SMTPSpec struct {
	// Address
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Address string `json:"address"`
	// User
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	User SecretReference `json:"user"`
	// Password
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Password SecretReference `json:"password"`
	// Port
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Port int32 `json:"port"`
	// Authentication protocol
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	AuthProtocol string `json:"authProtocol"`
	// OpenSSL verify mode
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	OpenSSLVerifyMode string `json:"opensslVerifyMode"`
	// Enable/disable STARTTLS
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	STARTTLS *bool `json:"starttls,omitempty"`
	// Enable/disable auto STARTTLS
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	STARTTLSAuto *bool `json:"starttlsAuto,omitempty"`
}

// SystemZyncSpec has configuration options for zync
type SystemZyncSpec struct {
	// Zync authentication token
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	AuthToken SecretReference `json:"authToken"`
	// Zync endpoint
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Endpoint string `json:"endpoint"`
}

// SystemBackendSpec has configuration options for backend
type SystemBackendSpec struct {
	// External endpoint
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ExternalEndpoint string `json:"externalEndpoint"`
	// Internal endpoint
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	InternalEndpoint string `json:"internalEndpoint"`
	// Internal API user
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	InternalAPIUser SecretReference `json:"internalAPIUser"`
	// Internal API password
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	InternalAPIPassword SecretReference `json:"internalAPIPassword"`
	// Redis data source name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	RedisDSN string `json:"redisDSN"`
}

// AssetsSpec has configuration to access assets in AWS s3
type AssetsSpec struct {
	// AWS S3 bucket name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Bucket string `json:"bucket"`
	// AWS S3 region
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Region string `json:"region"`
	// AWS access key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	AccessKey SecretReference `json:"accessKey"`
	// AWS secret access key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecretKey SecretReference `json:"secretKey"`
	// Assets host (CDN)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Host *string `json:"host,omitempty"`
	// Assets custom S3 endpoint
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	S3Endpoint *string `json:"s3Endpoint,omitempty"`
}

// SystemRailsSpec configures rails for system components
type SystemRailsSpec struct {
	// Rails Console
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Console *bool `json:"console,omitempty"`
	// Rails environment
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Environment *string `json:"environment,omitempty"`
	// Rails log level (debug, info, warn, error, fatal or unknown)
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=debug;info;warn;error;fatal;unknown
	// +optional
	LogLevel *string `json:"logLevel,omitempty"`
}

// ApicastSpec holds properties to configure Apicast endpoints
type SystemApicastEndpointsSpec struct {
	// Apicast Staging endpoint
	StagingDomain string `json:"stagingDomain"`
	// Apicast Production endpoint
	ProductionDomain string `json:"productionDomain"`
	// Policies registry URL for Apicast Cloud Hosteed
	CloudHostedRegistryURL string `json:"cloudHostedRegistryURL"`
	// Policies registry URL for Apicast Self Managed (on-prem)
	SelfManagedRegistryURL string `json:"selfManagedRegistryURL"`
}

// Default applies defaults for SystemRailsSpec
func (srs *SystemRailsSpec) Default() {
	srs.Console = boolOrDefault(srs.Console, ptr.To(systemDefaultRailsConsole))
	srs.Environment = stringOrDefault(srs.Environment, ptr.To(systemDefaultRailsEnvironment))
	srs.LogLevel = stringOrDefault(srs.LogLevel, ptr.To(systemDefaultRailsLogLevel))
}

// SystemAppSpec configures the App component of System
type SystemAppSpec struct {
	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	DeploymentStrategy *DeploymentStrategySpec `json:"deploymentStrategy,omitempty"`
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
	// Configures the TerminationGracePeriodSeconds
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// Canary defines spec changes for the canary Deployment. If
	// left unset the canary Deployment wil not be created.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Canary *Canary `json:"canary,omitempty"`
	// Describes how the services provided by this workload are exposed to its consumers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PublishingStrategies *PublishingStrategies `json:"publishingStrategies,omitempty"`
}

// Default implements defaulting for the system App component
func (spec *SystemAppSpec) Default() {
	spec.DeploymentStrategy = InitializeDeploymentStrategySpec(spec.DeploymentStrategy, systemDefaultAppDeploymentStrategy)
	spec.HPA = InitializeHorizontalPodAutoscalerSpec(spec.HPA, systemDefaultAppHPA)
	spec.Replicas = intOrDefault(spec.Replicas, &systemDefaultAppReplicas)
	spec.PDB = InitializePodDisruptionBudgetSpec(spec.PDB, systemDefaultAppPDB)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, systemDefaultAppResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, systemDefaultAppLivenessProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, systemDefaultAppReadinessProbe)
	spec.TerminationGracePeriodSeconds = int64OrDefault(
		spec.TerminationGracePeriodSeconds, systemDefaultTerminationGracePeriodSeconds,
	)
	spec.PublishingStrategies = InitializePublishingStrategies(spec.PublishingStrategies)
}

// SystemSidekiqSpec configures the Sidekiq component of System
type SystemSidekiqSpec struct {
	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	DeploymentStrategy *DeploymentStrategySpec `json:"deploymentStrategy,omitempty"`
	// Sidekiq specific configuration options for the component element
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Config *SidekiqConfig `json:"config,omitempty"`
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
	// Canary defines spec changes for the canary Deployment. If
	// left unset the canary Deployment wil not be created.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Canary *Canary `json:"canary,omitempty"`
	// Configures the TerminationGracePeriodSeconds
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// SidekiqConfig configures app behavior for System Sidekiq
type SidekiqConfig struct {
	// List of queues to be consumed by sidekiq. Format: queue[,Priority]
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Queues []string `json:"queues,omitempty"`
	// Number of rails max threads per sidekiq pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MaxThreads *int32 `json:"maxThreads,omitempty"`
}

type defaultSidekiqConfig struct {
	Queues     []string
	MaxThreads *int32
}

// Default sets default values for any value not specifically set in the SidekiqConfig struct
func (cfg *SidekiqConfig) Default(def defaultSidekiqConfig) {
	if cfg.Queues == nil {
		cfg.Queues = def.Queues
	}

	cfg.MaxThreads = intOrDefault(cfg.MaxThreads, ptr.To[int32](*def.MaxThreads))
}

// Default implements defaulting for the system Sidekiq component
func (spec *SystemSidekiqSpec) Default(sidekiqType systemSidekiqType) {
	spec.DeploymentStrategy = InitializeDeploymentStrategySpec(spec.DeploymentStrategy, systemDefaultSidekiqDeploymentStrategy)
	spec.HPA = InitializeHorizontalPodAutoscalerSpec(spec.HPA, systemDefaultSidekiqHPA)
	spec.Replicas = intOrDefault(spec.Replicas, &systemDefaultSidekiqReplicas)
	spec.PDB = InitializePodDisruptionBudgetSpec(spec.PDB, systemDefaultSidekiqPDB)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, systemDefaultSidekiqResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, systemDefaultSidekiqLivenessProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, systemDefaultSidekiqReadinessProbe)
	spec.TerminationGracePeriodSeconds = int64OrDefault(
		spec.TerminationGracePeriodSeconds, systemDefaultTerminationGracePeriodSeconds,
	)

	if spec.Config == nil {
		spec.Config = &SidekiqConfig{}
	}

	if sidekiqType == Billing {
		spec.Config.Default(systemDefaultSidekiqConfigBilling)
	} else if sidekiqType == Low {
		spec.Config.Default(systemDefaultSidekiqConfigLow)
	} else {
		spec.Config.Default(systemDefaultSidekiqConfigDefault)
	}
}

// SystemSearchdSpec configures the App component of System
type SystemSearchdSpec struct {
	// Deploy searchd instance
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// Image specification for the Searchd component.
	// Defaults to system image if not defined.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// Configuration options for System's Searchd
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Config *SearchdConfig `json:"config,omitempty"`
	// Resource requirements for the Searchd component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Resources *ResourceRequirementsSpec `json:"resources,omitempty"`
	// Liveness probe for the Searchd component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	LivenessProbe *ProbeSpec `json:"livenessProbe,omitempty"`
	// Readiness probe for the Searchd component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ReadinessProbe *ProbeSpec `json:"readinessProbe,omitempty"`
	// Describes node affinity scheduling rules for the Searchd pod
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// If specified, the Searchd pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Configures the TerminationGracePeriodSeconds for Searchd
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// Default implements defaulting for the system searchd component
func (spec *SystemSearchdSpec) Default() {
	spec.Enabled = boolOrDefault(spec.Enabled, ptr.To(systemDefaultSearchdEnabled))
	spec.Image = InitializeImageSpec(spec.Image, systemDefaultSearchdImage)
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, systemDefaultSearchdResources)
	spec.LivenessProbe = InitializeProbeSpec(spec.LivenessProbe, systemDefaultSearchdLivenessProbe)
	spec.ReadinessProbe = InitializeProbeSpec(spec.ReadinessProbe, systemDefaultSearchdReadinessProbe)

	if spec.Config == nil {
		spec.Config = &SearchdConfig{}
	}

	spec.Config.Default()
	spec.TerminationGracePeriodSeconds = int64OrDefault(
		spec.TerminationGracePeriodSeconds, systemDefaultTerminationGracePeriodSeconds,
	)
}

// SearchdConfig has configuration options for System's searchd
type SearchdConfig struct {
	// Allows setting the service name for Searchd
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ServiceName *string `json:"serviceName,omitempty"`
	// The TCP port Searchd will run its daemon on
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Port *int32 `json:"port,omitempty"`
	// Searchd database path
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	DatabasePath *string `json:"databasePath,omitempty"`
	// Searchd database storage size
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	DatabaseStorageSize *resource.Quantity `json:"databaseStorageSize,omitempty"`
	// Searchd database storage type
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	DatabaseStorageClass *string `json:"databaseStorageClass,omitempty"`
}

// Default implements defaulting for SearchdConfig
func (sc *SearchdConfig) Default() {
	sc.ServiceName = stringOrDefault(sc.ServiceName, ptr.To(systemDefaultSearchdServiceName))
	sc.Port = intOrDefault(sc.Port, ptr.To[int32](systemDefaultSearchdPort))
	sc.DatabasePath = stringOrDefault(sc.DatabasePath, ptr.To(systemDefaultSearchdDBPath))

	if sc.DatabaseStorageSize == nil {
		size := resource.MustParse(systemDefaultSearchdDatabaseStorageSize)
		sc.DatabaseStorageSize = &size
	}
}

// SystemRailsConsoleSpec configures the App component of System
type SystemRailsConsoleSpec struct {
	// Image specification for the Console component.
	// Defaults to system image if not defined.
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

// Default implements defaulting for the system App component
func (spec *SystemRailsConsoleSpec) Default(systemDefaultImage *ImageSpec) {
	spec.Image = InitializeImageSpec(spec.Image, defaultImageSpec(*systemDefaultImage))
	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, systemDefaultRailsConsoleResources)
}

// SystemStatus defines the observed state of System
type SystemStatus struct {
	AggregatedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// System is the Schema for the systems API
type System struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SystemSpec   `json:"spec,omitempty"`
	Status SystemStatus `json:"status,omitempty"`
}

// Default implements defaulting for the System resource
func (s *System) Default() {
	s.Spec.Default()
}

var _ reconciler.ObjectWithAppStatus = &System{}

func (d *System) GetStatus() any {
	return &d.Status
}

// +kubebuilder:object:root=true

// SystemList contains a list of System
type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []System `json:"items"`
}

// GetItem returns a client.Objectfrom a SystemList
func (sl *SystemList) GetItem(idx int) client.Object {
	return &sl.Items[idx]
}

// CountItems returns the item count in SystemList.Items
func (sl *SystemList) CountItems() int {
	return len(sl.Items)
}

func init() {
	SchemeBuilder.Register(&System{}, &SystemList{})
}

// SystemTektonTaskSpec configures the Sidekiq component of System
type SystemTektonTaskSpec struct {
	// Deploy task instance
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// Name for the Tekton task and pipeline
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Name *string `json:"name,omitempty"`
	// Description for the Tekton task and pipeline
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Description *string `json:"description,omitempty"`
	// System Tekton Task specific configuration options for the component element
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Config *SystemTektonTaskConfig `json:"config,omitempty"`
	// Pod Disruption Budget for the component
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Resources *ResourceRequirementsSpec `json:"resources,omitempty"`
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Configures the TerminationGracePeriodSeconds
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// SystemTektonTaskConfig configures app behavior for System SystemTektonTask
type SystemTektonTaskConfig struct {
	// Image specification for the Console component.
	// Defaults to system image if not defined.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Image *ImageSpec `json:"image,omitempty"`
	// List of commands to be consumed by the task.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Command []string `json:"command,omitempty"`
	// List of args to be consumed by the task.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Args []string `json:"args,omitempty"`
	// List of extra evironment variables to be consumed by the task.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
	// Timeout for the Tekton task
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// Default sets default values for any value not specifically set in the SystemTektonTaskConfig struct
func (cfg *SystemTektonTaskConfig) Default(systemDefaultImage *ImageSpec) {
	cfg.Image = InitializeImageSpec(cfg.Image, defaultImageSpec(*systemDefaultImage))
	cfg.Command = stringSliceOrDefault(cfg.Command, []string{"echo"})
	cfg.Args = stringSliceOrDefault(cfg.Args, []string{"Step command not set."})
	cfg.Timeout = durationOrDefault(cfg.Timeout, &systemDefaultSystemTektonTasksTimeout)

	if len(cfg.ExtraEnv) == 0 {
		cfg.ExtraEnv = []corev1.EnvVar{}
	}
}

// Merges default preloaded task values for any value not specifically set in the SystemTektonTaskConfig struct
func (cfg *SystemTektonTaskConfig) Merge(def SystemTektonTaskConfig) {
	if cfg == nil {
		cfg = &SystemTektonTaskConfig{}
	}

	cfg.Command = stringSliceOrDefault(cfg.Command, def.Command)
	cfg.Args = stringSliceOrDefault(cfg.Args, def.Args)

	if len(cfg.ExtraEnv) == 0 {
		cfg.ExtraEnv = def.ExtraEnv
	}

	// If a DefaultExtraEnvVar is missing from the resource definition
	// is appended to the cfg.ExtraEnv slide.
	for _, DefaultExtraEnvVar := range def.ExtraEnv {
		found := false

		for _, ExtraEnvVar := range cfg.ExtraEnv {
			if DefaultExtraEnvVar.Name == ExtraEnvVar.Name {
				found = true
			}
		}

		if !found {
			cfg.ExtraEnv = append(cfg.ExtraEnv, DefaultExtraEnvVar)
		}
	}
}

// Default implements defaulting for the system SystemTektonTask component
func (spec *SystemTektonTaskSpec) Default(systemDefaultImage *ImageSpec) {
	if spec.Config == nil {
		spec.Config = &SystemTektonTaskConfig{}
	}

	spec.Description = stringOrDefault(spec.Description, spec.Name)
	spec.Enabled = boolOrDefault(spec.Enabled, ptr.To(true))
	spec.Config.Default(systemDefaultImage)

	spec.Resources = InitializeResourceRequirementsSpec(spec.Resources, systemDefaultSystemTektonTaskResources)
	spec.TerminationGracePeriodSeconds = int64OrDefault(
		spec.TerminationGracePeriodSeconds, systemDefaultTerminationGracePeriodSeconds,
	)
}
