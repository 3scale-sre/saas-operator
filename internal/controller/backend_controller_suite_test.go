package controllers

import (
	"context"
	"errors"
	"time"

	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	testutil "github.com/3scale-sre/saas-operator/test/util"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Backend controller", func() {
	var backend *saasv1alpha1.Backend
	namespace := *(new(string))

	BeforeEach(func() {
		namespace = testutil.CreateNamespace(nameGenerator, k8sClient, timeout, poll)
	})

	When("deploying a defaulted Backend instance", func() {

		BeforeEach(func() {
			By("creating an Backend minimal resource")
			backend = &saasv1alpha1.Backend{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: namespace,
				},
				Spec: saasv1alpha1.BackendSpec{
					Config: saasv1alpha1.BackendConfig{
						RedisStorageDSN:     "storageDSN",
						RedisQueuesDSN:      "queuesDSN",
						SystemEventsHookURL: saasv1alpha1.SecretReference{Override: ptr.To("system-app")},
						SystemEventsHookPassword: saasv1alpha1.SecretReference{
							FromVault: &saasv1alpha1.VaultSecretReference{
								Path: "some-path-hook-password",
								Key:  "some-key-hook-password",
							},
						},
						InternalAPIUser: saasv1alpha1.SecretReference{
							FromVault: &saasv1alpha1.VaultSecretReference{
								Path: "some-path-api-user",
								Key:  "some-key-api-user",
							},
						},
						InternalAPIPassword: saasv1alpha1.SecretReference{
							FromVault: &saasv1alpha1.VaultSecretReference{
								Path: "some-path-api-password",
								Key:  "some-key-api-password",
							},
						},
						ErrorMonitoringKey: &saasv1alpha1.SecretReference{
							FromVault: &saasv1alpha1.VaultSecretReference{
								Path: "some-path-error-key",
								Key:  "some-key-error-key",
							},
						},
						ErrorMonitoringService: &saasv1alpha1.SecretReference{
							FromVault: &saasv1alpha1.VaultSecretReference{
								Path: "some-path-error-service",
								Key:  "some-key-error-service",
							},
						},
					},
					Worker: &saasv1alpha1.WorkerSpec{
						HPA: &saasv1alpha1.HorizontalPodAutoscalerSpec{
							Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
								ScaleUp: &autoscalingv2.HPAScalingRules{
									Policies: []autoscalingv2.HPAScalingPolicy{
										{
											Type:          autoscalingv2.PercentScalingPolicy,
											Value:         10,
											PeriodSeconds: 60,
										},
									},
								},
							},
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), backend)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, backend)
			}, timeout, poll).ShouldNot(HaveOccurred())

		})

		It("creates the required backend workload resources", func() {

			dep := &appsv1.Deployment{}
			By("deploying the backend-listener workload",
				(&testutil.ExpectedWorkload{
					Name:          "backend-listener",
					Namespace:     namespace,
					Replicas:      2,
					ContainerName: "backend-listener",
					ContainterArgs: []string{
						"bin/3scale_backend", "start",
						"-e", "production",
						"-p", "3000",
						"-x", "/dev/stdout",
					},
					Health:     "Progressing",
					PDB:        true,
					HPA:        true,
					PodMonitor: true,
				}).Assert(k8sClient, backend, dep, timeout, poll))

			svc := &corev1.Service{}
			By("deploying the backend-listener-http-svc service",
				(&testutil.ExpectedResource{
					Name: "backend-listener-http-svc", Namespace: namespace,
				}).Assert(k8sClient, svc, timeout, poll))

			Expect(svc.Spec.Selector["deployment"]).To(Equal("backend-listener"))

			By("deploying the backend-worker workload",
				(&testutil.ExpectedWorkload{
					Name:          "backend-worker",
					Namespace:     namespace,
					Replicas:      2,
					ContainerName: "backend-worker",
					ContainterArgs: []string{
						"bin/3scale_backend_worker", "run",
					},
					PDB:        true,
					HPA:        true,
					PodMonitor: true,
				}).Assert(k8sClient, backend, dep, timeout, poll))

			Expect(dep.Spec.Template.Spec.Containers[0].Env[12].Name).To(Equal("CONFIG_EVENTS_HOOK"))
			Expect(dep.Spec.Template.Spec.Containers[0].Env[12].Value).To(Equal("system-app"))

			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			By("updates backend-worker hpa behaviour",
				(&testutil.ExpectedResource{Name: "backend-worker", Namespace: namespace}).
					Assert(k8sClient, hpa, timeout, poll))
			Expect(hpa.Spec.Behavior.ScaleUp.Policies).To(Not(BeEmpty()))
			Expect(hpa.Spec.Behavior.ScaleUp.Policies[0].Type).To(Equal(autoscalingv2.PercentScalingPolicy))
			Expect(hpa.Spec.Behavior.ScaleUp.Policies[0].Value).To(Equal(int32(10)))

			By("deploying the backend-cron workload",
				(&testutil.ExpectedWorkload{
					Name:           "backend-cron",
					Namespace:      namespace,
					Replicas:       1,
					ContainerName:  "backend-cron",
					ContainterArgs: []string{"backend-cron"},
				}).Assert(k8sClient, backend, dep, timeout, poll))

		})

		It("creates the required backend shared resources", func() {

			gd := &grafanav1beta1.GrafanaDashboard{}
			By("deploying the backend grafana dashboard",
				(&testutil.ExpectedResource{Name: "backend", Namespace: namespace}).
					Assert(k8sClient, gd, timeout, poll))

			esApi := &externalsecretsv1beta1.ExternalSecret{}
			By("deploying the backend-internal-api external secret",
				(&testutil.ExpectedResource{Name: "backend-internal-api", Namespace: namespace}).
					Assert(k8sClient, esApi, timeout, poll))

			Expect(esApi.Spec.RefreshInterval.ToUnstructured()).To(Equal("1m0s"))
			Expect(esApi.Spec.SecretStoreRef.Name).To(Equal("vault-mgmt"))
			Expect(esApi.Spec.SecretStoreRef.Kind).To(Equal("ClusterSecretStore"))

			for _, data := range esApi.Spec.Data {
				switch data.SecretKey {
				case "CONFIG_INTERNAL_API_USER":
					Expect(data.RemoteRef.Property).To(Equal("some-key-api-user"))
					Expect(data.RemoteRef.Key).To(Equal("some-path-api-user"))
				case "CONFIG_INTERNAL_API_PASSWORD":
					Expect(data.RemoteRef.Property).To(Equal("some-key-api-password"))
					Expect(data.RemoteRef.Key).To(Equal("some-path-api-password"))
				}
			}

			esHook := &externalsecretsv1beta1.ExternalSecret{}
			By("deploying the backend-system-events-hook external secret",
				(&testutil.ExpectedResource{Name: "backend-system-events-hook", Namespace: namespace}).
					Assert(k8sClient, esHook, timeout, poll))

			Expect(esHook.Spec.RefreshInterval.ToUnstructured()).To(Equal("1m0s"))
			Expect(esHook.Spec.SecretStoreRef.Name).To(Equal("vault-mgmt"))
			Expect(esHook.Spec.SecretStoreRef.Kind).To(Equal("ClusterSecretStore"))

			for _, data := range esHook.Spec.Data {
				switch data.SecretKey {
				case "CONFIG_EVENTS_HOOK_SHARED_SECRET":
					Expect(data.RemoteRef.Property).To(Equal("some-key-hook-password"))
					Expect(data.RemoteRef.Key).To(Equal("some-path-hook-password"))
				}

				esError := &externalsecretsv1beta1.ExternalSecret{}
				By("deploying the backend-error-monitoring external secret",
					(&testutil.ExpectedResource{Name: "backend-error-monitoring", Namespace: namespace}).
						Assert(k8sClient, esError, timeout, poll))

				Expect(esError.Spec.RefreshInterval.ToUnstructured()).To(Equal("1m0s"))
				Expect(esError.Spec.SecretStoreRef.Name).To(Equal("vault-mgmt"))
				Expect(esError.Spec.SecretStoreRef.Kind).To(Equal("ClusterSecretStore"))

				for _, data := range esError.Spec.Data {
					switch data.SecretKey {
					case "CONFIG_HOPTOAD_API_KEY":
						Expect(data.RemoteRef.Property).To(Equal("some-key-error-key"))
						Expect(data.RemoteRef.Key).To(Equal("some-path-error-key"))
					case "CONFIG_HOPTOAD_SERVICE":
						Expect(data.RemoteRef.Property).To(Equal("some-key-error-service"))
						Expect(data.RemoteRef.Key).To(Equal("some-path-error-service"))
					}
				}
			}
		})

		It("doesn't create the non-default resources", func() {

			dep := &appsv1.Deployment{}
			By("ensuring an backend-listener-canary workload is not created",
				(&testutil.ExpectedResource{Name: "backend-listener-canary", Namespace: namespace, Missing: true}).
					Assert(k8sClient, dep, timeout, poll))
			By("ensuring an backend-worker-canary workload is not created",
				(&testutil.ExpectedResource{Name: "backend-worker-canary", Namespace: namespace, Missing: true}).
					Assert(k8sClient, dep, timeout, poll))

			ec := &marin3rv1alpha1.EnvoyConfig{}
			By("ensuring an backend-listener envoyconfig is not created",
				(&testutil.ExpectedResource{Name: "backend-listener", Namespace: namespace, Missing: true}).
					Assert(k8sClient, ec, timeout, poll))

		})

		When("updating a backend resource with some customizations", func() {

			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					backend := &saasv1alpha1.Backend{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						backend,
					); err != nil {
						return err
					}

					rvs["svc/backend-listener-http-svc"] = testutil.GetResourceVersion(
						k8sClient, &corev1.Service{}, "backend-listener-http-svc", namespace, timeout, poll)
					rvs["deployment/backend-listener"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-listener", namespace, timeout, poll)
					rvs["hpa/backend-worker"] = testutil.GetResourceVersion(
						k8sClient, &autoscalingv2.HorizontalPodAutoscaler{}, "backend-worker", namespace, timeout, poll)
					rvs["deployment/backend-cron"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-cron", namespace, timeout, poll)
					rvs["externalsecret/backend-internal-api"] = testutil.GetResourceVersion(
						k8sClient, &externalsecretsv1beta1.ExternalSecret{}, "backend-internal-api", namespace, timeout, poll)
					rvs["externalsecret/backend-system-events-hook"] = testutil.GetResourceVersion(
						k8sClient, &externalsecretsv1beta1.ExternalSecret{}, "backend-system-events-hook", namespace, timeout, poll)
					rvs["externalsecret/backend-error-monitoring"] = testutil.GetResourceVersion(
						k8sClient, &externalsecretsv1beta1.ExternalSecret{}, "backend-error-monitoring", namespace, timeout, poll)

					patch := client.MergeFrom(backend.DeepCopy())
					backend.Spec.Image = &saasv1alpha1.ImageSpec{
						Name: ptr.To("newImage"),
						Tag:  ptr.To("newTag"),
					}
					backend.Spec.Listener.Replicas = ptr.To[int32](3)
					backend.Spec.Listener.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
					backend.Spec.Listener.Config = &saasv1alpha1.ListenerConfig{
						RedisAsync: ptr.To(true),
					}
					backend.Spec.Worker = &saasv1alpha1.WorkerSpec{
						HPA: &saasv1alpha1.HorizontalPodAutoscalerSpec{
							MinReplicas: ptr.To[int32](3),
						},
					}
					backend.Spec.Cron = &saasv1alpha1.CronSpec{
						Replicas: ptr.To[int32](3),
					}

					backend.Spec.Config.ExternalSecret.RefreshInterval = &metav1.Duration{Duration: 1 * time.Second}
					backend.Spec.Config.ExternalSecret.SecretStoreRef = &saasv1alpha1.ExternalSecretSecretStoreReferenceSpec{
						Name: ptr.To("other-store"),
						Kind: ptr.To("SecretStore"),
					}
					backend.Spec.Config.InternalAPIUser.FromVault.Path = "secret/data/updated-path-api"
					backend.Spec.Config.SystemEventsHookPassword.FromVault.Path = "secret/data/updated-path-hook"
					backend.Spec.Config.ErrorMonitoringKey.FromVault.Path = "secret/data/updated-path-error"

					backend.Spec.Listener.PublishingStrategies = &saasv1alpha1.PublishingStrategies{
						Endpoints: []saasv1alpha1.PublishingStrategy{{
							Strategy:     "Marin3rSidecar",
							EndpointName: "HTTP",
							Marin3rSidecar: &saasv1alpha1.Marin3rSidecarSpec{
								Simple: &saasv1alpha1.Simple{
									ServiceType: ptr.To(saasv1alpha1.ServiceTypeNLB),
									NetworkLoadBalancerConfig: &saasv1alpha1.NetworkLoadBalancerSpec{
										CrossZoneLoadBalancingEnabled: ptr.To(false),
									},
								},
								EnvoyDynamicConfig: saasv1alpha1.MapOfEnvoyDynamicConfig{
									"http": {
										GeneratorVersion: ptr.To("v1"),
										ListenerHttp: &saasv1alpha1.ListenerHttp{
											Port:            8080,
											RouteConfigName: "route",
										},
									}}},
						}},
					}

					return k8sClient.Patch(context.Background(), backend, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("updates the backend-listener resources", func() {

				dep := &appsv1.Deployment{}
				By("updating the backend-listener workload",
					(&testutil.ExpectedWorkload{

						Name:           "backend-listener",
						Namespace:      namespace,
						Replicas:       3,
						ContainerName:  "backend-listener",
						ContainerImage: "newImage:newTag",
						ContainterArgs: []string{
							"bin/3scale_backend", "-s", "falcon", "start",
							"-e", "production",
							"-p", "3000",
							"-x", "/dev/stdout",
						},
						PDB:         true,
						PodMonitor:  true,
						EnvoyConfig: true,
						LastVersion: rvs["deployment/backend-listener"],
					}).Assert(k8sClient, backend, dep, timeout, poll))

				svc := &corev1.Service{}

				By("creating backend-listener-http-svc service",
					(&testutil.ExpectedResource{
						Name: "backend-listener-http-marin3r-nlb", Namespace: namespace,
					}).Assert(k8sClient, svc, timeout, poll))
				Expect(svc.Spec.Selector["deployment"]).To(Equal("backend-listener"))
				Expect(svc.GetAnnotations()["service.beta.kubernetes.io/aws-load-balancer-attributes"]).
					To(Equal("load_balancing.cross_zone.enabled=false,deletion_protection.enabled=false"))

				By("deleting backend-listener-http-svc service",
					(&testutil.ExpectedResource{
						Name: "backend-listener-http-svc", Namespace: namespace, Missing: true,
					}).Assert(k8sClient, svc, timeout, poll))

				hpa := &autoscalingv2.HorizontalPodAutoscaler{}
				By("updating the backend-worker workload",
					(&testutil.ExpectedResource{
						Name:        "backend-worker",
						Namespace:   namespace,
						LastVersion: rvs["hpa/backend-worker"],
					}).Assert(k8sClient, hpa, timeout, poll))

				Expect(hpa.Spec.MinReplicas).To(Equal(ptr.To[int32](3)))

				By("updating the backend-cron workload",
					(&testutil.ExpectedWorkload{
						Name:           "backend-cron",
						Namespace:      namespace,
						Replicas:       3,
						ContainerName:  "backend-cron",
						ContainerImage: "newImage:newTag",
						LastVersion:    rvs["deployment/backend-cron"],
					}).Assert(k8sClient, backend, dep, timeout, poll))

				esApi := &externalsecretsv1beta1.ExternalSecret{}
				By("updating the backend-internal-api external secret",
					(&testutil.ExpectedResource{
						Name:        "backend-internal-api",
						Namespace:   namespace,
						LastVersion: rvs["externalsecret/backend-internal-api"],
					}).Assert(k8sClient, esApi, timeout, poll))

				Expect(esApi.Spec.RefreshInterval.ToUnstructured()).To(Equal("1s"))
				Expect(esApi.Spec.SecretStoreRef.Name).To(Equal("other-store"))
				Expect(esApi.Spec.SecretStoreRef.Kind).To(Equal("SecretStore"))

				for _, data := range esApi.Spec.Data {
					switch data.SecretKey {
					case "CONFIG_INTERNAL_API_USER":
						Expect(data.RemoteRef.Key).To(Equal("updated-path-api"))
					}
				}

				esHook := &externalsecretsv1beta1.ExternalSecret{}
				By("updating the backend-system-events-hook external secret",
					(&testutil.ExpectedResource{
						Name:        "backend-system-events-hook",
						Namespace:   namespace,
						LastVersion: rvs["externalsecret/backend-system-events-hook"],
					}).Assert(k8sClient, esHook, timeout, poll))

				Expect(esHook.Spec.RefreshInterval.ToUnstructured()).To(Equal("1s"))
				Expect(esHook.Spec.SecretStoreRef.Name).To(Equal("other-store"))
				Expect(esHook.Spec.SecretStoreRef.Kind).To(Equal("SecretStore"))

				for _, data := range esHook.Spec.Data {
					switch data.SecretKey {
					case "CONFIG_EVENTS_HOOK_SHARED_SECRET":
						Expect(data.RemoteRef.Key).To(Equal("updated-path-hook"))
					}
				}

				esError := &externalsecretsv1beta1.ExternalSecret{}
				By("updating the backend-error-monitoring external secret",
					(&testutil.ExpectedResource{
						Name:        "backend-error-monitoring",
						Namespace:   namespace,
						LastVersion: rvs["backend-error-monitoring"],
					}).Assert(k8sClient, esError, timeout, poll))

				Expect(esError.Spec.RefreshInterval.ToUnstructured()).To(Equal("1s"))
				Expect(esError.Spec.SecretStoreRef.Name).To(Equal("other-store"))
				Expect(esError.Spec.SecretStoreRef.Kind).To(Equal("SecretStore"))

				for _, data := range esError.Spec.Data {
					switch data.SecretKey {
					case "CONFIG_HOPTOAD_API_KEY":
						Expect(data.RemoteRef.Key).To(Equal("updated-path-error"))
					}
				}
			})
		})

		When("updating a backend resource with canary", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					backend := &saasv1alpha1.Backend{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						backend,
					); err != nil {
						return err
					}

					rvs["svc/backend-listener"] = testutil.GetResourceVersion(
						k8sClient, &corev1.Service{}, "backend-listener-http-svc", namespace, timeout, poll)
					rvs["deployment/backend-listener"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-listener", namespace, timeout, poll)
					rvs["deployment/backend-worker"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-worker", namespace, timeout, poll)

					patch := client.MergeFrom(backend.DeepCopy())
					backend.Spec.Listener.Canary = &saasv1alpha1.Canary{
						ImageName: ptr.To("newImage"),
						ImageTag:  ptr.To("newTag"),
					}
					backend.Spec.Worker = &saasv1alpha1.WorkerSpec{
						Canary: &saasv1alpha1.Canary{
							ImageName: ptr.To("newImage"),
							ImageTag:  ptr.To("newTag"),
							Patches: []string{
								`[{"op": "add", "path": "/config/rackEnv", "value": "test"}]`,
								`[{"op": "replace", "path": "/config/redisStorageDSN", "value": "testDSN"}]`,
							},
						},
					}
					if err := k8sClient.Patch(context.Background(), backend, patch); err != nil {
						return err
					}

					if testutil.GetResourceVersion(k8sClient, &appsv1.Deployment{}, "backend-listener-canary", namespace, timeout, poll) == "" {
						return errors.New("not ready")
					}
					if testutil.GetResourceVersion(k8sClient, &appsv1.Deployment{}, "backend-worker-canary", namespace, timeout, poll) == "" {
						return errors.New("not ready")
					}

					return nil

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("creates the required cannary resources", func() {

				dep := &appsv1.Deployment{}
				By("deploying the backend-listener-canary workload",
					(&testutil.ExpectedWorkload{

						Name:           "backend-listener-canary",
						Namespace:      namespace,
						Replicas:       2,
						ContainerName:  "backend-listener",
						ContainerImage: "newImage:newTag",
						ContainterArgs: []string{
							"bin/3scale_backend", "start",
							"-e", "production",
							"-p", "3000",
							"-x", "/dev/stdout",
						},
						PodMonitor:  true,
						LastVersion: rvs["deployment/backend-listener"],
					}).Assert(k8sClient, backend, dep, timeout, poll))

				svc := &corev1.Service{}
				By("keeps the backend-listener-http-svc service deployment label selector",
					(&testutil.ExpectedResource{
						Name: "backend-listener-http-svc", Namespace: namespace,
					}).Assert(k8sClient, svc, timeout, poll))

				Expect(svc.Spec.Selector["deployment"]).To(Equal("backend-listener"))
				Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("backend-listener"))

				By("deploying the backend-worker-canary workload",
					(&testutil.ExpectedWorkload{

						Name:           "backend-worker-canary",
						Namespace:      namespace,
						Replicas:       2,
						ContainerName:  "backend-worker",
						ContainerImage: "newImage:newTag",
						ContainterArgs: []string{
							"bin/3scale_backend_worker", "run",
						},
						PodMonitor: true,
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("RACK_ENV"))
				Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Value).To(Equal("test"))
				Expect(dep.Spec.Template.Spec.Containers[0].Env[1].Name).To(Equal("CONFIG_REDIS_PROXY"))
				Expect(dep.Spec.Template.Spec.Containers[0].Env[1].Value).To(Equal("testDSN"))

			})

			Context("and enabling canary traffic", func() {

				BeforeEach(func() {
					Eventually(func() error {

						backend := &saasv1alpha1.Backend{}
						if err := k8sClient.Get(
							context.Background(),
							types.NamespacedName{Name: "instance", Namespace: namespace},
							backend,
						); err != nil {
							return err
						}

						rvs["deployment/backend-listener-canary"] = testutil.GetResourceVersion(
							k8sClient, &appsv1.Deployment{}, "backend-listener-canary", namespace, timeout, poll)
						rvs["svc/backend-listener-http-svc"] = testutil.GetResourceVersion(
							k8sClient, &corev1.Service{}, "backend-listener-http-svc", namespace, timeout, poll)
						rvs["deployment/backend-worker-canary"] = testutil.GetResourceVersion(
							k8sClient, &appsv1.Deployment{}, "backend-worker-canary", namespace, timeout, poll)

						patch := client.MergeFrom(backend.DeepCopy())
						backend.Spec.Listener.Replicas = ptr.To[int32](3)
						backend.Spec.Listener.Canary = &saasv1alpha1.Canary{
							SendTraffic: *ptr.To(true),
							Replicas:    ptr.To[int32](3),
						}
						backend.Spec.Worker = &saasv1alpha1.WorkerSpec{
							Replicas: ptr.To[int32](3),
						}
						backend.Spec.Worker.Canary = &saasv1alpha1.Canary{
							Replicas: ptr.To[int32](3),
						}

						return k8sClient.Patch(context.Background(), backend, patch)

					}, timeout, poll).ShouldNot(HaveOccurred())
				})

				It("updates the backend resources", func() {

					dep := &appsv1.Deployment{}
					By("scaling up the backend-listener-canary workload",
						(&testutil.ExpectedWorkload{

							Name:        "backend-listener-canary",
							Namespace:   namespace,
							Replicas:    3,
							PodMonitor:  true,
							LastVersion: rvs["deployment/backend-listener-canary"],
						}).Assert(k8sClient, backend, dep, timeout, poll))

					Expect(dep.Spec.Replicas).To(Equal(ptr.To[int32](3)))

					svc := &corev1.Service{}
					By("removing the backend-listener service deployment label selector",
						(&testutil.ExpectedResource{
							Name: "backend-listener-http-svc", Namespace: namespace,
							LastVersion: rvs["svc/backend-listener-http-svc"],
						}).Assert(k8sClient, svc, timeout, poll))

					Expect(svc.Spec.Selector).ToNot(HaveKey("deployment"))
					Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("backend-listener"))

					By("scaling up the backend-worker-canary workload",
						(&testutil.ExpectedWorkload{

							Name:        "backend-worker-canary",
							Namespace:   namespace,
							Replicas:    3,
							PodMonitor:  true,
							LastVersion: rvs["deployment/backend-worker-canary"],
						}).Assert(k8sClient, backend, dep, timeout, poll))

				})

			})

			When("disabling the canary", func() {

				BeforeEach(func() {

					Eventually(func() error {
						backend := &saasv1alpha1.Backend{}
						if err := k8sClient.Get(
							context.Background(),
							types.NamespacedName{Name: "instance", Namespace: namespace},
							backend,
						); err != nil {
							return err
						}
						patch := client.MergeFrom(backend.DeepCopy())
						backend.Spec.Listener.Canary = nil
						backend.Spec.Worker.Canary = nil

						return k8sClient.Patch(context.Background(), backend, patch)
					}, timeout, poll).ShouldNot(HaveOccurred())
				})

				It("deletes the canary resources", func() {

					dep := &appsv1.Deployment{}
					By("removing the backend-listener-canary Deployment",
						(&testutil.ExpectedResource{
							Name: "backend-listener-canary", Namespace: namespace, Missing: true}).Assert(k8sClient, dep, timeout, poll))
					By("removing the backend-worker-canary Deployment",
						(&testutil.ExpectedResource{
							Name: "backend-worker-canary", Namespace: namespace, Missing: true}).Assert(k8sClient, dep, timeout, poll))

					pm := &monitoringv1.PodMonitor{}
					By("removing the backend-listener-canary PodMonitor",
						(&testutil.ExpectedResource{
							Name: "backend-listener-canary", Namespace: namespace, Missing: true}).Assert(k8sClient, pm, timeout, poll))
					By("removing the backend-worker-canary PodMonitor",
						(&testutil.ExpectedResource{
							Name: "backend-worker-canary", Namespace: namespace, Missing: true}).Assert(k8sClient, pm, timeout, poll))
				})
			})
		})

		When("updating a backend resource with twemproxyconfig", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					backend := &saasv1alpha1.Backend{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						backend,
					); err != nil {
						return err
					}

					rvs["deployment/backend-listener"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-listener", namespace, timeout, poll)
					rvs["deployment/backend-worker"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-worker", namespace, timeout, poll)
					rvs["deployment/backend-cron"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "backend-cron", namespace, timeout, poll)

					patch := client.MergeFrom(backend.DeepCopy())
					backend.Spec.Listener.Replicas = ptr.To[int32](2)
					backend.Spec.Listener.Canary = &saasv1alpha1.Canary{
						Replicas: ptr.To[int32](2),
						Patches: []string{
							`[{"op":"add","path":"/twemproxy","value":{"twemproxyConfigRef":"backend-canary-twemproxyconfig","options":{"logLevel":3}}}]`,
						},
					}
					backend.Spec.Worker = &saasv1alpha1.WorkerSpec{
						Replicas: ptr.To[int32](2),
					}
					backend.Spec.Worker.Canary = &saasv1alpha1.Canary{
						Replicas: ptr.To[int32](2),
						Patches: []string{
							`[{"op":"add","path":"/twemproxy/options","value":{"logLevel":4}}]`,
						},
					}
					backend.Spec.Twemproxy = &saasv1alpha1.TwemproxySpec{
						TwemproxyConfigRef: "backend-twemproxyconfig",
						Options: &saasv1alpha1.TwemproxyOptions{
							LogLevel: ptr.To[int32](2),
						},
					}

					return k8sClient.Patch(context.Background(), backend, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("updates the backend-listener resources", func() {

				dep := &appsv1.Deployment{}
				By("adding a twemproxy sidecar to the backend-listener workload",
					(&testutil.ExpectedWorkload{

						Name:        "backend-listener",
						Namespace:   namespace,
						Replicas:    2,
						PDB:         true,
						HPA:         true,
						PodMonitor:  true,
						LastVersion: rvs["deployment/backend-listener"],
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Volumes[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("backend-twemproxyconfig"))
				Expect(dep.Spec.Template.Spec.Containers[1].Name).To(Equal("twemproxy"))
				Expect(dep.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(Equal("twemproxy-config"))

				By("adding a twemproxy sidecar to the backend-listener-canary workload",
					(&testutil.ExpectedWorkload{

						Name:       "backend-listener-canary",
						Replicas:   2,
						Namespace:  namespace,
						PodMonitor: true,
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Volumes[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("backend-canary-twemproxyconfig"))
				Expect(dep.Spec.Template.Spec.Containers[1].Name).To(Equal("twemproxy"))
				Expect(dep.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Containers[1].Env[3].Name).To(Equal("TWEMPROXY_LOG_LEVEL"))
				Expect(dep.Spec.Template.Spec.Containers[1].Env[3].Value).To(Equal("3"))

				By("adding a twemproxy sidecar to the backend-worker workload",
					(&testutil.ExpectedWorkload{

						Name:        "backend-worker",
						Namespace:   namespace,
						Replicas:    2,
						PDB:         true,
						HPA:         true,
						PodMonitor:  true,
						LastVersion: rvs["deployment/backend-worker"],
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Volumes[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("backend-twemproxyconfig"))
				Expect(dep.Spec.Template.Spec.Containers[1].Name).To(Equal("twemproxy"))
				Expect(dep.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(Equal("twemproxy-config"))

				By("adding a twemproxy sidecar to the backend-worker-canary workload",
					(&testutil.ExpectedWorkload{

						Name:       "backend-worker-canary",
						Replicas:   2,
						Namespace:  namespace,
						PodMonitor: true,
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Volumes[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("backend-twemproxyconfig"))
				Expect(dep.Spec.Template.Spec.Containers[1].Name).To(Equal("twemproxy"))
				Expect(dep.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(Equal("twemproxy-config"))
				Expect(dep.Spec.Template.Spec.Containers[1].Env[3].Name).To(Equal("TWEMPROXY_LOG_LEVEL"))
				Expect(dep.Spec.Template.Spec.Containers[1].Env[3].Value).To(Equal("4"))

				By("not updating the backend-cron workload",
					(&testutil.ExpectedWorkload{

						Name:      "backend-cron",
						Namespace: namespace,
						Replicas:  1,
					}).Assert(k8sClient, backend, dep, timeout, poll))

				Expect(dep.GetResourceVersion()).To(Equal(rvs["deployment/backend-cron"]))
				Expect(dep.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(dep.Spec.Template.Spec.Containers[0].Name).To(Equal("backend-cron"))

			})
		})

	})
})
