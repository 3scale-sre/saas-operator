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

	"github.com/3scale-ops/basereconciler/reconciler"
	"github.com/3scale-ops/basereconciler/util"
	saasv1alpha1 "github.com/3scale-ops/saas-operator/api/v1alpha1"
	"github.com/3scale-ops/saas-operator/pkg/generators/mappingservice"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// MappingServiceReconciler reconciles a MappingService object
type MappingServiceReconciler struct {
	*reconciler.Reconciler
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

	ctx, _ = r.Logger(ctx, "name", req.Name, "namespace", req.Namespace)
	instance := &saasv1alpha1.MappingService{}
	result := r.ManageResourceLifecycle(ctx, req, instance,
		reconciler.WithInMemoryInitializationFunc(util.ResourceDefaulter(instance)))
	if result.ShouldReturn() {
		return result.Values()
	}

	gen := mappingservice.NewGenerator(instance.GetName(), instance.GetNamespace(), instance.Spec)
	resources, err := gen.Resources()
	if err != nil {
		return ctrl.Result{}, err
	}

	result = r.ReconcileOwnedResources(ctx, instance, resources)
	if result.ShouldReturn() {
		return result.Values()
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MappingServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return reconciler.SetupWithDynamicTypeWatches(r,
		ctrl.NewControllerManagedBy(mgr).
			For(&saasv1alpha1.MappingService{}).
			Watches(&source.Kind{Type: &corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret"}}},
				r.FilteredEventHandler(&saasv1alpha1.MappingServiceList{}, nil, r.Log)),
	)
}
