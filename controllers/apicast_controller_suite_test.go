package controllers

import (
	"context"

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	grafanav1alpha1 "github.com/3scale/saas-operator/pkg/apis/grafana/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Apicast controller", func() {
	var namespace string
	var apicast *saasv1alpha1.Apicast

	BeforeEach(func() {
		// Create a namespace for each block
		namespace = "test-ns-" + nameGenerator.Generate()

		// Add any setup steps that needs to be executed before each test
		testNamespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		err := k8sClient.Create(context.Background(), testNamespace)
		Expect(err).ToNot(HaveOccurred())

		n := &corev1.Namespace{}
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace}, n)
		}, timeout, poll).ShouldNot(HaveOccurred())

	})

	When("deploying a defaulted Apicast instance", func() {

		BeforeEach(func() {
			By("creating an Apicast simple resource", func() {
				apicast = &saasv1alpha1.Apicast{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: namespace,
					},
					Spec: saasv1alpha1.ApicastSpec{
						Staging: saasv1alpha1.ApicastEnvironmentSpec{
							Config: saasv1alpha1.ApicastConfig{
								ConfigurationCache:       30,
								ThreescalePortalEndpoint: "http://example/config",
							},
							Endpoint: saasv1alpha1.Endpoint{
								DNS: []string{"apicast-staging.example.com"},
							},
						},
						Production: saasv1alpha1.ApicastEnvironmentSpec{
							Config: saasv1alpha1.ApicastConfig{
								ConfigurationCache:       300,
								ThreescalePortalEndpoint: "http://example/config",
							},
							Endpoint: saasv1alpha1.Endpoint{
								DNS: []string{"apicast-production.example.com"},
							},
						},
					},
				}
				err := k8sClient.Create(context.Background(), apicast)
				Expect(err).ToNot(HaveOccurred())
				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, apicast)
				}, timeout, poll).ShouldNot(HaveOccurred())
			})
		})

		It("creates the required Apicast resources", func() {

			dep := &appsv1.Deployment{}
			By("deploying an apicast-production workload",
				checkWorkloadResources(dep,
					expectedWorkload{
						Name:          "apicast-production",
						Namespace:     namespace,
						Replicas:      2,
						ContainerName: "apicast",
						PDB:           true,
						HPA:           true,
						PodMonitor:    true,
					},
				),
			)
			for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
				switch env.Name {
				case "THREESCALE_DEPLOYMENT_ENV":
					Expect(env.Value).To(Equal("production"))
				case "APICAST_CONFIGURATION_LOADER":
					Expect(env.Value).To(Equal("lazy"))
				case "APICAST_LOG_LEVEL":
					Expect(env.Value).To(Equal("warn"))
				case "APICAST_RESPONSE_CODES":
					Expect(env.Value).To(Equal("true"))
				}
			}

			By("deploying an apicast-staging workload",
				checkWorkloadResources(dep,
					expectedWorkload{
						Name:          "apicast-staging",
						Namespace:     namespace,
						Replicas:      2,
						ContainerName: "apicast",
						PDB:           true,
						HPA:           true,
						PodMonitor:    true,
					},
				),
			)
			for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
				switch env.Name {
				case "APICAST_CONFIGURATION_LOADER":
					Expect(env.Value).To(Equal("lazy"))
				case "APICAST_LOG_LEVEL":
					Expect(env.Value).To(Equal("warn"))
				case "THREESCALE_DEPLOYMENT_ENV":
					Expect(env.Value).To(Equal("staging"))
				case "APICAST_RESPONSE_CODES":
					Expect(env.Value).To(Equal("true"))
				}
			}

			svc := &corev1.Service{}
			By("deploying the apicast-production service",
				checkResource(svc, expectedResource{
					Name: "apicast-production", Namespace: namespace,
				}),
			)
			Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-production"))
			Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-production"))
			Expect(
				svc.Annotations["external-dns.alpha.kubernetes.io/hostname"],
			).To(Equal("apicast-production.example.com"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"],
			).To(Equal("*"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"],
			).To(Equal("true"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"],
			).To(Equal("true"))

			By("deploying the apicast-production-management service",
				checkResource(svc, expectedResource{
					Name: "apicast-production-management", Namespace: namespace,
				}),
			)
			Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-production"))
			Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-production"))
			Expect(svc.Spec.Ports[0].Name).To(Equal("management"))
			Expect(svc.Annotations).To(HaveLen(0))

			By("deploying the apicast-staging service",
				checkResource(svc, expectedResource{
					Name: "apicast-staging", Namespace: namespace,
				}),
			)
			Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-staging"))
			Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-staging"))
			Expect(
				svc.Annotations["external-dns.alpha.kubernetes.io/hostname"],
			).To(Equal("apicast-staging.example.com"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"],
			).To(Equal("*"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"],
			).To(Equal("true"))
			Expect(
				svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"],
			).To(Equal("true"))

			By("deploying the apicast-staging-management service",
				checkResource(svc, expectedResource{
					Name: "apicast-staging-management", Namespace: namespace,
				}),
			)
			Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-staging"))
			Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-staging"))
			Expect(svc.Spec.Ports[0].Name).To(Equal("management"))
			Expect(svc.Annotations).To(HaveLen(0))

			By("deploying an apicast grafana dashboard",
				checkResource(
					&grafanav1alpha1.GrafanaDashboard{},
					expectedResource{
						Name:      "apicast",
						Namespace: namespace,
					},
				),
			)

			By("deploying an apicast-services grafana dashboard",
				checkResource(
					&grafanav1alpha1.GrafanaDashboard{},
					expectedResource{
						Name:      "apicast-services",
						Namespace: namespace,
					},
				),
			)

		})

		It("doesn't creates the non-default resources", func() {

			dep := &appsv1.Deployment{}
			By("ensuring an apicast-production-canary workload is gone",
				checkResource(dep,
					expectedResource{
						Name:      "apicast-production-canary",
						Namespace: namespace,
						Missing:   true,
					},
				),
			)
			By("ensuring an apicast-staging-canary workload is gone",
				checkResource(dep,
					expectedResource{
						Name:      "apicast-staging-canary",
						Namespace: namespace,
						Missing:   true,
					},
				),
			)

		})

		When("updating an Apicast resource with customizations", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					apicast := &saasv1alpha1.Apicast{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						apicast,
					); err != nil {
						return err
					}

					rvs["deployment/apicast-production"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-production", namespace,
					)
					rvs["deployment/apicast-staging"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-staging", namespace,
					)

					patch := client.MergeFrom(apicast.DeepCopy())
					apicast.Spec.Production = saasv1alpha1.ApicastEnvironmentSpec{
						Config: saasv1alpha1.ApicastConfig{
							ConfigurationCache:       42,
							ThreescalePortalEndpoint: "http://updated-example/config",
						},
						Endpoint: saasv1alpha1.Endpoint{
							DNS: []string{"updated-apicast-production.example.com"},
						},
						HPA: &saasv1alpha1.HorizontalPodAutoscalerSpec{
							MinReplicas: pointer.Int32(3),
						},
						LivenessProbe:  &saasv1alpha1.ProbeSpec{},
						ReadinessProbe: &saasv1alpha1.ProbeSpec{},
					}
					apicast.Spec.Staging = saasv1alpha1.ApicastEnvironmentSpec{
						Config: saasv1alpha1.ApicastConfig{
							ConfigurationCache:       42,
							ThreescalePortalEndpoint: "http://updated-example/config",
						},
						Endpoint: saasv1alpha1.Endpoint{
							DNS: []string{"updated-apicast-staging.example.com"},
						},
						HPA: &saasv1alpha1.HorizontalPodAutoscalerSpec{
							MinReplicas: pointer.Int32(3),
						},
						LivenessProbe:  &saasv1alpha1.ProbeSpec{},
						ReadinessProbe: &saasv1alpha1.ProbeSpec{},
					}
					apicast.Spec.GrafanaDashboard = &saasv1alpha1.GrafanaDashboardSpec{}

					return k8sClient.Patch(context.Background(), apicast, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("updates the Apicast resources", func() {

				By("ensuring the Apicast grafana dashboard is gone",
					checkResource(
						&grafanav1alpha1.GrafanaDashboard{},
						expectedResource{
							Name:      "apicast-production",
							Namespace: namespace,
							Missing:   true,
						},
					),
				)

				dep := &appsv1.Deployment{}
				By("updating the Apicast Production workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:          "apicast-production",
							Namespace:     namespace,
							Replicas:      3,
							ContainerName: "apicast",
							HPA:           true,
							PDB:           true,
							PodMonitor:    true,
							LastVersion:   rvs["deployment/apicast-production"],
						},
					),
				)
				for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
					switch env.Name {
					case "THREESCALE_PORTAL_ENDPOINT":
						Expect(env.Value).To(Equal("http://updated-example/config"))
					case "THREESCALE_DEPLOYMENT_ENV":
						Expect(env.Value).To(Equal("production"))
					case "APICAST_RESPONSE_CODES":
						Expect(env.Value).To(Equal("true"))
					case "APICAST_CONFIGURATION_CACHE":
						Expect(env.Value).To(Equal("42"))
					}
				}

				By("updating the Apicast Staging workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:          "apicast-staging",
							Namespace:     namespace,
							Replicas:      3,
							ContainerName: "apicast",
							HPA:           true,
							PDB:           true,
							PodMonitor:    true,
							LastVersion:   rvs["deployment/apicast-staging"],
						},
					),
				)
				for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
					switch env.Name {
					case "THREESCALE_PORTAL_ENDPOINT":
						Expect(env.Value).To(Equal("http://updated-example/config"))
					case "APICAST_LOG_LEVEL":
						Expect(env.Value).To(Equal("warn"))
					case "THREESCALE_DEPLOYMENT_ENV":
						Expect(env.Value).To(Equal("staging"))
					case "APICAST_RESPONSE_CODES":
						Expect(env.Value).To(Equal("true"))
					case "APICAST_CONFIGURATION_CACHE":
						Expect(env.Value).To(Equal("42"))
					}
				}

				svc := &corev1.Service{}
				By("updating the apicast-production service annotation",
					checkResource(svc, expectedResource{
						Name: "apicast-production", Namespace: namespace,
					}),
				)
				Expect(svc.Annotations["external-dns.alpha.kubernetes.io/hostname"]).To(
					Equal("updated-apicast-production.example.com"),
				)

				By("updating the apicast-staging service annotation",
					checkResource(svc, expectedResource{
						Name: "apicast-staging", Namespace: namespace,
					}),
				)
				Expect(svc.Annotations["external-dns.alpha.kubernetes.io/hostname"]).To(
					Equal("updated-apicast-staging.example.com"),
				)

			})

		})

		When("updating an Apicast resource with canary", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {
					apicast := &saasv1alpha1.Apicast{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						apicast,
					); err != nil {
						return err
					}

					rvs["svc/apicast-production"] = getResourceVersion(
						&corev1.Service{}, "apicast-production", namespace,
					)
					rvs["deployment/apicast-production"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-production", namespace,
					)
					rvs["svc/apicast-staging"] = getResourceVersion(
						&corev1.Service{}, "apicast-staging", namespace,
					)
					rvs["deployment/apicast-staging"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-staging", namespace,
					)

					patch := client.MergeFrom(apicast.DeepCopy())
					apicast.Spec.Production.Canary = &saasv1alpha1.Canary{
						ImageName: pointer.StringPtr("newImage"),
						ImageTag:  pointer.StringPtr("newTag"),
						Replicas:  pointer.Int32Ptr(1),
					}
					apicast.Spec.Staging.Canary = &saasv1alpha1.Canary{
						ImageName: pointer.StringPtr("newImage"),
						ImageTag:  pointer.StringPtr("newTag"),
						Replicas:  pointer.Int32Ptr(1),
					}

					return k8sClient.Patch(context.Background(), apicast, patch)
				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("creates the required canary resources", func() {

				dep := &appsv1.Deployment{}
				By("deploying an Apicast-production-canary workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:          "apicast-production-canary",
							Namespace:     namespace,
							Replicas:      1,
							ContainerName: "apicast",
							PodMonitor:    true,
						},
					),
				)
				for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
					switch env.Name {
					case "THREESCALE_DEPLOYMENT_ENV":
						Expect(env.Value).To(Equal("production"))
					case "APICAST_LOG_LEVEL":
						Expect(env.Value).To(Equal("warn"))
					case "APICAST_RESPONSE_CODES":
						Expect(env.Value).To(Equal("true"))
					}
				}

				By("deploying an Apicast-staging-canary workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:          "apicast-staging-canary",
							Namespace:     namespace,
							Replicas:      1,
							ContainerName: "apicast",
							PodMonitor:    true,
						},
					),
				)
				for _, env := range dep.Spec.Template.Spec.Containers[0].Env {
					switch env.Name {
					case "THREESCALE_DEPLOYMENT_ENV":
						Expect(env.Value).To(Equal("staging"))
					case "APICAST_LOG_LEVEL":
						Expect(env.Value).To(Equal("warn"))
					case "APICAST_RESPONSE_CODES":
						Expect(env.Value).To(Equal("true"))
					}
				}

				svc := &corev1.Service{}
				By("keeping the apicast-production service spec",
					checkResource(svc, expectedResource{
						Name: "apicast-production", Namespace: namespace,
					}),
				)
				Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-production"))
				Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-production"))
				Expect(
					svc.Annotations["external-dns.alpha.kubernetes.io/hostname"],
				).To(Equal("apicast-production.example.com"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"],
				).To(Equal("*"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"],
				).To(Equal("true"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"],
				).To(Equal("true"))

				By("keeping the apicast-production-management service spec",
					checkResource(svc, expectedResource{
						Name: "apicast-production-management", Namespace: namespace,
					}),
				)
				Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-production"))
				Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-production"))
				Expect(svc.Spec.Ports[0].Name).To(Equal("management"))
				Expect(svc.Annotations).To(HaveLen(0))

				By("keeping the apicast-staging service spec",
					checkResource(svc, expectedResource{
						Name: "apicast-staging", Namespace: namespace,
					}),
				)
				Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-staging"))
				Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-staging"))
				Expect(
					svc.Annotations["external-dns.alpha.kubernetes.io/hostname"],
				).To(Equal("apicast-staging.example.com"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"],
				).To(Equal("*"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"],
				).To(Equal("true"))
				Expect(
					svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"],
				).To(Equal("true"))

				By("keeping the apicast-staging-management service spec",
					checkResource(svc, expectedResource{
						Name: "apicast-staging-management", Namespace: namespace,
					}),
				)
				Expect(svc.Spec.Selector["deployment"]).To(Equal("apicast-staging"))
				Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-staging"))
				Expect(svc.Spec.Ports[0].Name).To(Equal("management"))
				Expect(svc.Annotations).To(HaveLen(0))

			})

			When("enabling canary traffic", func() {

				BeforeEach(func() {
					Eventually(func() error {
						apicast := &saasv1alpha1.Apicast{}
						if err := k8sClient.Get(
							context.Background(),
							types.NamespacedName{Name: "instance", Namespace: namespace},
							apicast,
						); err != nil {
							return err
						}
						rvs["svc/apicast-production"] = getResourceVersion(
							&corev1.Service{}, "apicast-production", namespace,
						)
						rvs["svc/apicast-staging"] = getResourceVersion(
							&corev1.Service{}, "apicast-staging", namespace,
						)
						patch := client.MergeFrom(apicast.DeepCopy())
						apicast.Spec.Production.Canary = &saasv1alpha1.Canary{
							SendTraffic: *pointer.Bool(true),
						}
						apicast.Spec.Staging.Canary = &saasv1alpha1.Canary{
							SendTraffic: *pointer.Bool(true),
						}
						return k8sClient.Patch(context.Background(), apicast, patch)
					}, timeout, poll).ShouldNot(HaveOccurred())
				})

				It("updates the apicast service", func() {

					svc := &corev1.Service{}
					By("removing the apicast-production service deployment label selector",
						checkResource(svc, expectedResource{
							Name: "apicast-production", Namespace: namespace,
							LastVersion: rvs["svc/apicast-production"],
						}),
					)
					Expect(svc.Spec.Selector).NotTo(HaveKey("deployment"))
					Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-production"))

					By("removing the apicast-staging service deployment label selector",
						checkResource(svc, expectedResource{
							Name: "apicast-staging", Namespace: namespace,
							LastVersion: rvs["svc/apicast-staging"],
						}),
					)
					Expect(svc.Spec.Selector).NotTo(HaveKey("deployment"))
					Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("apicast-staging"))

				})

			})

		})

		When("removing the PDB and HPA from an Apicast instance", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					apicast := &saasv1alpha1.Apicast{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						apicast,
					); err != nil {
						return err
					}

					rvs["deployment/apicast-production"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-production", namespace,
					)
					rvs["deployment/apicast-staging"] = getResourceVersion(
						&appsv1.Deployment{}, "apicast-staging", namespace,
					)

					patch := client.MergeFrom(apicast.DeepCopy())

					apicast.Spec.Production.Replicas = pointer.Int32(0)
					apicast.Spec.Production.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
					apicast.Spec.Production.PDB = &saasv1alpha1.PodDisruptionBudgetSpec{}

					apicast.Spec.Staging.Replicas = pointer.Int32(0)
					apicast.Spec.Staging.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
					apicast.Spec.Staging.PDB = &saasv1alpha1.PodDisruptionBudgetSpec{}

					return k8sClient.Patch(context.Background(), apicast, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("removes the Apicast disabled resources", func() {

				dep := &appsv1.Deployment{}
				By("updating the Apicast Production workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:        "apicast-production",
							Namespace:   namespace,
							Replicas:    0,
							HPA:         false,
							PDB:         false,
							PodMonitor:  true,
							LastVersion: rvs["deployment/apicast-production"],
						},
					),
				)
				By("updating the Apicast Staging workload",
					checkWorkloadResources(dep,
						expectedWorkload{
							Name:        "apicast-staging",
							Namespace:   namespace,
							Replicas:    0,
							HPA:         false,
							PDB:         false,
							PodMonitor:  true,
							LastVersion: rvs["deployment/apicast-staging"],
						},
					),
				)

			})

		})

	})

})
