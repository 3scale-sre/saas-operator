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
	"fmt"
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	"github.com/3scale-sre/basereconciler/util"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/pkg/generators/redisshard"
	"github.com/3scale-sre/saas-operator/pkg/redis/client"
	redis "github.com/3scale-sre/saas-operator/pkg/redis/server"
	"github.com/3scale-sre/saas-operator/pkg/redis/sharded"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RedisShardReconciler reconciles a RedisShard object
type RedisShardReconciler struct {
	*reconciler.Reconciler
	Pool *redis.ServerPool
}

// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=redisshards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=redisshards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=saas.3scale.net,namespace=placeholder,resources=redisshards/finalizers,verbs=update
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",namespace=placeholder,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",namespace=placeholder,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RedisShardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx, logger := r.Logger(ctx, "name", req.Name, "namespace", req.Namespace)
	instance := &saasv1alpha1.RedisShard{}
	result := r.ManageResourceLifecycle(ctx, req, instance,
		reconciler.WithInMemoryInitializationFunc(util.ResourceDefaulter(instance)))
	if result.ShouldReturn() {
		return result.Values()
	}

	gen := redisshard.NewGenerator(instance.GetName(), instance.GetNamespace(), instance.Spec)

	result = r.ReconcileOwnedResources(ctx, instance, gen.Resources())
	if result.ShouldReturn() {
		return result.Values()
	}

	shard, result := r.setRedisRoles(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		*instance.Spec.MasterIndex, *instance.Spec.SlaveCount+1, gen.ServiceName(), logger)
	if result.ShouldReturn() {
		return result.Values()
	}

	if err := r.updateStatus(ctx, shard, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisShardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return reconciler.SetupWithDynamicTypeWatches(r,
		ctrl.NewControllerManagedBy(mgr).
			For(&saasv1alpha1.RedisShard{}),
	)
}

func (r *RedisShardReconciler) setRedisRoles(ctx context.Context, key types.NamespacedName,
	masterIndex, replicas int32, serviceName string, log logr.Logger) (*sharded.Shard, reconciler.Result) {

	var masterHostPort string
	redisURLs := make(map[string]string, replicas)
	for i := 0; i < int(replicas); i++ {
		pod := &corev1.Pod{}
		key := types.NamespacedName{Name: fmt.Sprintf("%s-%d", serviceName, i), Namespace: key.Namespace}
		err := r.Client.Get(ctx, key, pod)
		if err != nil {
			return &sharded.Shard{Name: key.Name}, reconciler.Result{Error: err}
		}
		if pod.Status.PodIP == "" {
			log.Info("waiting for pod IP to be allocated")
			// return &sharded.Shard{Name: key.Name}, &ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			return &sharded.Shard{Name: key.Name}, reconciler.Result{Action: reconciler.ReturnAndRequeueAction, RequeueAfter: 5 * time.Second}
		}

		redisURLs[fmt.Sprintf("%s-%d", serviceName, i)] = fmt.Sprintf("redis://%s:%d", pod.Status.PodIP, 6379)
		if int(masterIndex) == i {
			masterHostPort = fmt.Sprintf("%s:%d", pod.Status.PodIP, 6379)
		}
	}

	shard, err := sharded.NewShardFromTopology(key.Name, redisURLs, r.Pool)
	if err != nil {
		return shard, reconciler.Result{Error: err}
	}

	_, err = shard.Init(ctx, masterHostPort)
	if err != nil {
		log.Info("waiting for redis shard init")
		return shard, reconciler.Result{Action: reconciler.ReturnAndRequeueAction, RequeueAfter: 10 * time.Second}
	}

	return shard, reconciler.Result{}
}

func (r *RedisShardReconciler) updateStatus(ctx context.Context, shard *sharded.Shard, instance *saasv1alpha1.RedisShard, log logr.Logger) error {

	status := saasv1alpha1.RedisShardStatus{
		ShardNodes: &saasv1alpha1.RedisShardNodes{Master: map[string]string{}, Slaves: map[string]string{}},
	}

	for _, server := range shard.Servers {
		if server.Role == client.Master {
			status.ShardNodes.Master[server.GetAlias()] = server.ID()
		} else if server.Role == client.Slave {
			status.ShardNodes.Slaves[server.GetAlias()] = server.ID()
		}
	}
	if !equality.Semantic.DeepEqual(status, instance.Status) {
		instance.Status = status
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return err
		}
	}

	return nil
}
