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
	"encoding/json"

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/basereconciler"
	"github.com/3scale/saas-operator/pkg/generators/sentinel"
	"github.com/3scale/saas-operator/pkg/redis"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SentinelReconciler reconciles a Sentinel object
type SentinelReconciler struct {
	basereconciler.Reconciler
	Log logr.Logger
}

// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels/finalizers,verbs=update
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",namespace=placeholder,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="monitoring.coreos.com",namespace=placeholder,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="policy",namespace=placeholder,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="integreatly.org",namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SentinelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)

	instance := &saasv1alpha1.Sentinel{}
	key := types.NamespacedName{Name: req.Name, Namespace: req.Namespace}
	result, err := r.GetInstance(ctx, key, instance, saasv1alpha1.Finalizer, log)
	if result != nil || err != nil {
		return *result, err
	}

	// Apply defaults for reconcile but do not store them in the API
	instance.Default()
	json, _ := json.Marshal(instance.Spec)
	log.V(1).Info("Apply defaults before resolving templates", "JSON", string(json))

	gen := sentinel.NewGenerator(
		instance.GetName(),
		instance.GetNamespace(),
		instance.Spec,
	)

	err = r.ReconcileOwnedResources(ctx, instance, basereconciler.ControlledResources{
		StatefulSets: []basereconciler.StatefulSet{{
			Template:        gen.StatefulSet(),
			RolloutTriggers: nil,
			Enabled:         true,
		}},
		ConfigMaps: []basereconciler.ConfigMaps{{
			Template: gen.ConfigMap(),
			Enabled:  true,
		}},
		Services: func() []basereconciler.Service {
			fns := append(gen.PodServices(int(*instance.Spec.Replicas)), gen.StatefulSetService())
			var svcs = []basereconciler.Service{}
			for _, fn := range fns {
				svcs = append(svcs, basereconciler.Service{Template: fn, Enabled: true})
			}
			return svcs
		}(),
		// PodDisruptionBudgets: []basereconciler.PodDisruptionBudget{{
		// 	Template: gen.PDB(),
		// 	Enabled:  !instance.Spec.PDB.IsDeactivated(),
		// }},
	})

	if err != nil {
		log.Error(err, "unable to update owned resources")
		return r.ManageError(ctx, instance, err)
	}

	sentinelPool, err := redis.NewSentinelPool(ctx, r.GetClient(),
		types.NamespacedName{Name: gen.GetComponent(), Namespace: gen.GetNamespace()}, int(*instance.Spec.Replicas))
	if err != nil {
		return r.ManageError(ctx, instance, err)
	}

	shardedCluster, err := redis.NewShardedCluster(ctx, instance.Spec.Config.MonitoredShards, log)
	if err != nil {
		return r.ManageError(ctx, instance, err)
	}

	if err := sentinelPool.Monitor(ctx, shardedCluster); err != nil {
		return r.ManageError(ctx, instance, err)
	}

	return r.ManageSuccess(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SentinelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&saasv1alpha1.Sentinel{}).
		Watches(&source.Channel{Source: r.GetStatusChangeChannel()}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
