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

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/generators/mappingservice"
	basereconciler "github.com/3scale/saas-operator/pkg/reconcilers/basereconciler/v2"
	"github.com/3scale/saas-operator/pkg/reconcilers/workloads"
	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// MappingServiceReconciler reconciles a MappingService object
type MappingServiceReconciler struct {
	workloads.WorkloadReconciler
	Log logr.Logger
}

// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=mappingservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=mappingservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=mappingservices/finalizers,verbs=update
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",namespace=placeholder,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="monitoring.coreos.com",namespace=placeholder,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="autoscaling",namespace=placeholder,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="policy",namespace=placeholder,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="integreatly.org",namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="external-secrets.io",namespace=placeholder,resources=externalsecrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MappingServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = log.IntoContext(ctx, logger)

	instance := &saasv1alpha1.MappingService{}
	key := types.NamespacedName{Name: req.Name, Namespace: req.Namespace}
	result, err := r.GetInstance(ctx, key, instance, saasv1alpha1.Finalizer, nil)
	if result != nil || err != nil {
		return *result, err
	}

	// Apply defaults for reconcile but do not store them in the API
	instance.Default()

	gen := mappingservice.NewGenerator(
		instance.GetName(),
		instance.GetNamespace(),
		instance.Spec,
	)

	resources := []basereconciler.Resource{
		gen.GrafanaDashboard(),
		gen.ExternalSecret(),
	}

	var workload []basereconciler.Resource
	workload, err = r.NewDeploymentWorkloadWithTraffic(ctx, instance, r.GetScheme(), &gen, &gen)
	if err != nil {
		return r.ManageError(ctx, instance, err)
	}
	resources = append(resources, workload...)

	err = r.ReconcileOwnedResources(ctx, instance, resources)
	if err != nil {
		logger.Error(err, "unable to reconcile owned resources")
		return r.ManageError(ctx, instance, err)
	}

	return r.ManageSuccess(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MappingServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&saasv1alpha1.MappingService{}, builder.WithPredicates(util.ResourceGenerationOrFinalizerChangedPredicate{})).
		Owns(&corev1.Service{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Watches(&source.Channel{Source: r.GetStatusChangeChannel()}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret"}}},
			r.SecretEventHandler(&saasv1alpha1.MappingServiceList{}, r.Log)).
		Complete(r)
}
