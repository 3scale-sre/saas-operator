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

package controllers

import (
	"context"
	"crypto/rand"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	"github.com/3scale-sre/saas-operator/internal/pkg/reconcilers/threads"
	redis "github.com/3scale-sre/saas-operator/internal/pkg/redis/server"
	"github.com/goombaio/namegenerator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient     client.Client
	testEnv       *envtest.Environment
	nameGenerator namegenerator.Generator
	timeout       time.Duration = 60 * time.Second
	poll          time.Duration = 10 * time.Second
	ctx           context.Context
	cancel        context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../..", "config", "crd", "bases"),
			filepath.Join("../..", "test", "assets", "external-apis"),
		},
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	Expect(err).NotTo(HaveOccurred())
	nameGenerator = namegenerator.NewNameGenerator(nBig.Int64())

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = saasv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = grafanav1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = externalsecretsv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = marin3rv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = pipelinev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		// Disable the metrics port to allow running the
		// test suite in parallel
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient = mgr.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	// nolint: fatcontext
	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	redisPool := redis.NewServerPool()

	// Add controllers for testing
	err = (&AutoSSLReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("AutoSSL")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&ApicastReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("Apicast")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&EchoAPIReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("EchoAPI")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&MappingServiceReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("MappingService")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&CORSProxyReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("CORSProxy")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&BackendReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("Backend")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&RedisShardReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("RedisShard")),
		Pool: redisPool,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&SentinelReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("Sentinel")),
		SentinelEvents: threads.NewManager(),
		Metrics:        threads.NewManager(),
		Pool:           redisPool,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&SystemReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("System")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&ZyncReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("Zync")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
})
