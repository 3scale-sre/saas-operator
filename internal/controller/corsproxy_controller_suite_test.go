package controllers

import (
	"context"
	"time"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	testutil "github.com/3scale-sre/saas-operator/test/util"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CORSProxy controller", func() {
	var corsproxy *saasv1alpha1.CORSProxy
	namespace := *(new(string))

	BeforeEach(func() {
		namespace = testutil.CreateNamespace(nameGenerator, k8sClient, timeout, poll)
	})

	When("deploying a defaulted CORSProxy instance", func() {

		BeforeEach(func() {

			By("creating a CORSProxy resource", func() {

				corsproxy = &saasv1alpha1.CORSProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: namespace,
					},
					Spec: saasv1alpha1.CORSProxySpec{
						Config: saasv1alpha1.CORSProxyConfig{
							SystemDatabaseDSN: saasv1alpha1.SecretReference{
								FromVault: &saasv1alpha1.VaultSecretReference{
									Path: "some-path",
									Key:  "some-key",
								},
							},
						},
					},
				}
				err := k8sClient.Create(context.Background(), corsproxy)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: "instance", Namespace: namespace}, corsproxy)
				}, timeout, poll).ShouldNot(HaveOccurred())

			})
		})

		It("creates the required CORSProxy resources", func() {

			dep := &appsv1.Deployment{}
			By("deploying a CORSProxy workload",
				(&testutil.ExpectedWorkload{

					Name:          "cors-proxy",
					Namespace:     namespace,
					Replicas:      2,
					ContainerName: "cors-proxy",
					Health:        "Progressing",
					PDB:           true,
					HPA:           true,
					PodMonitor:    true,
				}).Assert(k8sClient, corsproxy, dep, timeout, poll))

			Expect(dep.Spec.Template.Spec.Volumes).To(BeEmpty())
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("DATABASE_URL"))
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.SecretKeyRef.Key).To(Equal("DATABASE_URL"))
			Expect(dep.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.SecretKeyRef.LocalObjectReference.Name).To(Equal("cors-proxy-system-database"))

			svc := &corev1.Service{}
			By("deploying a CORSProxy service",
				(&testutil.ExpectedResource{
					Name:      "cors-proxy-http-svc",
					Namespace: namespace,
				}).Assert(k8sClient, svc, timeout, poll))

			Expect(svc.Spec.Selector["deployment"]).To(Equal("cors-proxy"))
			Expect(svc.Spec.Selector["saas.3scale.net/traffic"]).To(Equal("cors-proxy"))

			es := &externalsecretsv1beta1.ExternalSecret{}
			By("deploying the CORSProxy System Database external secret",
				(&testutil.ExpectedResource{
					Name:      "cors-proxy-system-database",
					Namespace: namespace,
				}).Assert(k8sClient, es, timeout, poll))

			Expect(es.Spec.RefreshInterval.ToUnstructured()).To(Equal("1m0s"))
			Expect(es.Spec.SecretStoreRef.Name).To(Equal("vault-mgmt"))
			Expect(es.Spec.SecretStoreRef.Kind).To(Equal("ClusterSecretStore"))

			for _, data := range es.Spec.Data {
				switch data.SecretKey {
				case "DATABASE_URL":
					Expect(data.RemoteRef.Property).To(Equal("some-key"))
					Expect(data.RemoteRef.Key).To(Equal("some-path"))
				}
			}

			By("deploying the CORSProxy grafana dashboard",
				(&testutil.ExpectedResource{
					Name:      "cors-proxy",
					Namespace: namespace,
				}).Assert(k8sClient, &grafanav1beta1.GrafanaDashboard{}, timeout, poll))

		})

		When("updating a CORSProxy resource with customizations", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					corsproxy := &saasv1alpha1.CORSProxy{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						corsproxy,
					); err != nil {
						return err
					}

					rvs["cors-proxy"] = testutil.GetResourceVersion(
						k8sClient, corsproxy, "instance", namespace, timeout, poll)

					rvs["deployment/corsproxy"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "cors-proxy", namespace, timeout, poll)

					rvs["externalsecret/cors-proxy-system-database"] = testutil.GetResourceVersion(
						k8sClient, &externalsecretsv1beta1.ExternalSecret{}, "cors-proxy-system-database", namespace, timeout, poll)

					patch := client.MergeFrom(corsproxy.DeepCopy())
					corsproxy.Spec.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{
						MinReplicas: ptr.To[int32](3),
					}
					corsproxy.Spec.LivenessProbe = &saasv1alpha1.ProbeSpec{}
					corsproxy.Spec.ReadinessProbe = &saasv1alpha1.ProbeSpec{}
					corsproxy.Spec.Config.ExternalSecret.RefreshInterval = &metav1.Duration{Duration: 1 * time.Second}
					corsproxy.Spec.Config.ExternalSecret.SecretStoreRef = &saasv1alpha1.ExternalSecretSecretStoreReferenceSpec{
						Name: ptr.To("other-store"),
						Kind: ptr.To("SecretStore"),
					}
					corsproxy.Spec.Config.SystemDatabaseDSN.FromVault.Path = "secret/data/updated-path"
					corsproxy.Spec.GrafanaDashboard = &saasv1alpha1.GrafanaDashboardSpec{}

					return k8sClient.Patch(context.Background(), corsproxy, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("updates CORSProxy resources", func() {

				dep := &appsv1.Deployment{}
				By("updating the CORSProxy workload",
					(&testutil.ExpectedWorkload{

						Name:          "cors-proxy",
						Namespace:     namespace,
						Replicas:      3,
						ContainerName: "cors-proxy",
						PDB:           true,
						HPA:           true,
						PodMonitor:    true,
						LastVersion:   rvs["deployment/corsproxy"],
					}).Assert(k8sClient, corsproxy, dep, timeout, poll))

				Expect(dep.Spec.Template.Spec.Containers[0].LivenessProbe).To(BeNil())
				Expect(dep.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())

				es := &externalsecretsv1beta1.ExternalSecret{}
				By("updating the CORSProxy System Database external secret",
					(&testutil.ExpectedResource{
						Name:        "cors-proxy-system-database",
						Namespace:   namespace,
						LastVersion: rvs["externalsecret/cors-proxy-system-database"],
					}).Assert(k8sClient, es, timeout, poll))

				Expect(es.Spec.RefreshInterval.ToUnstructured()).To(Equal("1s"))
				Expect(es.Spec.SecretStoreRef.Name).To(Equal("other-store"))
				Expect(es.Spec.SecretStoreRef.Kind).To(Equal("SecretStore"))

				for _, data := range es.Spec.Data {
					switch data.SecretKey {
					case "DATABASE_URL":
						Expect(data.RemoteRef.Key).To(Equal("updated-path"))
					}
				}

				By("ensuring the CORSProxy grafana dashboard is gone",
					(&testutil.ExpectedResource{
						Name:      "cors-proxy",
						Namespace: namespace,
						Missing:   true,
					}).Assert(k8sClient, &grafanav1beta1.GrafanaDashboard{}, timeout, poll))

			})

		})

		When("removing the PDB and HPA from a CORSProxy instance", func() {

			// Resource Versions
			rvs := make(map[string]string)

			BeforeEach(func() {
				Eventually(func() error {

					corsproxy := &saasv1alpha1.CORSProxy{}
					if err := k8sClient.Get(
						context.Background(),
						types.NamespacedName{Name: "instance", Namespace: namespace},
						corsproxy,
					); err != nil {
						return err
					}

					rvs["deployment/corsproxy"] = testutil.GetResourceVersion(
						k8sClient, &appsv1.Deployment{}, "cors-proxy", namespace, timeout, poll)

					patch := client.MergeFrom(corsproxy.DeepCopy())
					corsproxy.Spec.Replicas = ptr.To[int32](0)
					corsproxy.Spec.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
					corsproxy.Spec.PDB = &saasv1alpha1.PodDisruptionBudgetSpec{}

					return k8sClient.Patch(context.Background(), corsproxy, patch)

				}, timeout, poll).ShouldNot(HaveOccurred())
			})

			It("removes the CORSProxy disabled resources", func() {

				dep := &appsv1.Deployment{}
				By("updating the CORSProxy workload",
					(&testutil.ExpectedWorkload{

						Name:        "cors-proxy",
						Namespace:   namespace,
						Replicas:    0,
						HPA:         false,
						PDB:         false,
						PodMonitor:  true,
						LastVersion: rvs["deployment/corsproxy"],
					}).Assert(k8sClient, corsproxy, dep, timeout, poll))

			})

		})

	})

})
