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
	"errors"
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	"github.com/3scale-sre/basereconciler/util"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/generators/sentinel"
	"github.com/3scale-sre/saas-operator/internal/pkg/reconcilers/threads"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/events"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/metrics"
	redis "github.com/3scale-sre/saas-operator/internal/pkg/redis/server"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/sharded"
	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SentinelReconciler reconciles a Sentinel object
type SentinelReconciler struct {
	*reconciler.Reconciler
	SentinelEvents threads.Manager
	Metrics        threads.Manager
	Pool           *redis.ServerPool
}

// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=sentinels/finalizers,verbs=update
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",namespace=placeholder,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="monitoring.coreos.com",namespace=placeholder,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="policy",namespace=placeholder,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="grafana.integreatly.org",namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SentinelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger := r.Logger(ctx, "name", req.Name, "namespace", req.Namespace)
	instance := &saasv1alpha1.Sentinel{}

	result := r.ManageResourceLifecycle(ctx, req, instance,
		reconciler.WithInMemoryInitializationFunc(util.ResourceDefaulter(instance)),
		reconciler.WithFinalizer(saasv1alpha1.Finalizer),
		reconciler.WithFinalizationFunc(r.SentinelEvents.CleanupThreads(instance)),
		reconciler.WithFinalizationFunc(r.Metrics.CleanupThreads(instance)),
	)
	if result.ShouldReturn() {
		return result.Values()
	}

	gen := sentinel.NewGenerator(instance.GetName(), instance.GetNamespace(), instance.Spec)

	result = r.ReconcileOwnedResources(ctx, instance, gen.Resources())
	if result.ShouldReturn() {
		return result.Values()
	}

	clustermap, err := gen.ClusterTopology(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	shardedCluster, err := sharded.NewShardedClusterFromTopology(ctx, clustermap, r.Pool)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ensure all shards are being monitored
	for _, sentinel := range shardedCluster.Sentinels {
		allMonitored, err := sentinel.IsMonitoringShards(ctx, shardedCluster.GetShardNames())
		if err != nil {
			return ctrl.Result{}, err
		}

		if !allMonitored {
			if err := shardedCluster.Discover(ctx); err != nil {
				return ctrl.Result{}, err
			}

			if _, err := sentinel.Monitor(ctx, shardedCluster, saasv1alpha1.SentinelDefaultQuorum); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Reconcile sentinel the event watchers and metrics gatherers
	eventWatchers := make([]threads.RunnableThread, 0, len(gen.SentinelURIs()))
	metricsGatherers := make([]threads.RunnableThread, 0, len(gen.SentinelURIs()))

	for _, uri := range gen.SentinelURIs() {
		watcher, err := events.NewSentinelEventWatcher(uri, instance, shardedCluster, true, r.Pool)
		if err != nil {
			return ctrl.Result{}, err
		}

		gatherer, err := metrics.NewSentinelMetricsGatherer(uri, *gen.Spec.Config.MetricsRefreshInterval, r.Pool)
		if err != nil {
			return ctrl.Result{}, err
		}

		eventWatchers = append(eventWatchers, watcher)
		metricsGatherers = append(metricsGatherers, gatherer)
	}

	if err := r.SentinelEvents.ReconcileThreads(ctx, instance, eventWatchers, logger.WithName("event-watcher")); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Metrics.ReconcileThreads(ctx, instance, metricsGatherers, logger.WithName("metrics-gatherer")); err != nil {
		return ctrl.Result{}, err
	}

	// reconcile the status
	result = r.ReconcileStatus(ctx, instance,
		nil, []types.NamespacedName{gen.GetKey()},
		func() (bool, error) { return sentinelStatusReconciler(ctx, instance, shardedCluster, logger) },
	)
	if result.ShouldReturn() {
		return result.Values()
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func sentinelStatusReconciler(ctx context.Context,
	instance *saasv1alpha1.Sentinel, cluster *sharded.Cluster, log logr.Logger) (bool, error) {
	// sentinels info to the status
	sentinels := make([]string, len(cluster.Sentinels))
	for idx, srv := range cluster.Sentinels {
		sentinels[idx] = srv.ID()
	}

	// redis shards info to the status
	merr := cluster.SentinelDiscover(ctx, sharded.SlaveReadOnlyDiscoveryOpt,
		sharded.SaveConfigDiscoveryOpt, sharded.SlavePriorityDiscoveryOpt, sharded.ReplicationInfoDiscoveryOpt)
	// if the failure occurred calling sentinel discard the result and return error
	// otherwise keep going on and use the information that was returned, even if there were some
	// other errors
	sentinelError := &sharded.DiscoveryError_Sentinel_Failure{}
	if errors.As(merr, sentinelError) {
		return false, merr
	}
	// We don't want the controller to keep failing while things reconfigure as
	// this makes controller throttling to kick in. Instead, just log the errors
	// and rely on reconciles triggered by sentinel events to correct the situation.
	masterError := &sharded.DiscoveryError_Master_SingleServerFailure{}
	slaveError := &sharded.DiscoveryError_Slave_SingleServerFailure{}
	failoverInProgressError := &sharded.DiscoveryError_Slave_FailoverInProgress{}

	if errors.As(merr, masterError) || errors.As(merr, slaveError) || errors.As(merr, failoverInProgressError) {
		log.Error(merr, "DiscoveryError")
	}

	// publish metrics based on the discovered cluster status
	if err := metrics.FromShardedCluster(ctx, cluster, false, instance.GetName()); err != nil {
		log.Error(err, "unable to publish redis cluster status metrics")
	}

	shards := make(saasv1alpha1.MonitoredShards, len(cluster.Shards))
	for idx, shard := range cluster.Shards {
		shards[idx] = saasv1alpha1.MonitoredShard{
			Name:    shard.Name,
			Servers: make(map[string]saasv1alpha1.RedisServerDetails, len(shard.Servers)),
		}
		for _, srv := range shard.Servers {
			shards[idx].Servers[srv.GetAlias()] = saasv1alpha1.RedisServerDetails{
				Role:    srv.Role,
				Address: srv.ID(),
				Config:  srv.Config,
				Info:    srv.Info,
			}
		}
	}

	if !equality.Semantic.DeepEqual(sentinels, instance.Status.Sentinels) ||
		!equality.Semantic.DeepEqual(shards, instance.Status.MonitoredShards) {
		// update required
		instance.Status.Sentinels = sentinels
		instance.Status.MonitoredShards = shards

		return true, nil
	}

	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SentinelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return reconciler.SetupWithDynamicTypeWatches(r,
		ctrl.NewControllerManagedBy(mgr).
			For(&saasv1alpha1.Sentinel{}).
			WatchesRawSource(source.Channel(r.SentinelEvents.GetChannel(), &handler.EnqueueRequestForObject{})).
			WithOptions(controller.Options{RateLimiter: PermissiveRateLimiter()}),
	)
}

func PermissiveRateLimiter() workqueue.TypedRateLimiter[reconcile.Request] {
	// return workqueue.DefaultControllerRateLimiter()
	return workqueue.NewTypedMaxOfRateLimiter(
		// First retries are more spaced that default
		// Max retry time is limited to 10 seconds
		workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](5*time.Millisecond, 10*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.TypedBucketRateLimiter[reconcile.Request]{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
}
