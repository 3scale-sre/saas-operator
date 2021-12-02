package redis

import (
	"context"
	"fmt"

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	redistypes "github.com/3scale/saas-operator/pkg/redis/types"
	"github.com/3scale/saas-operator/pkg/util"
	"github.com/go-redis/redis/v8"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	shardNotInitializedError = "ERR No such master with that name"
)

// SentinelServer represents a sentinel Pod
type SentinelServer string

// IsMonitoringShards checks whether or all the shards in the passed list are being monitored by the SentinelServer
func (ss *SentinelServer) IsMonitoringShards(ctx context.Context, shards []string) (bool, error) {

	monitoredShards, err := ss.Masters(ctx)
	if err != nil {
		return false, err
	}

	if len(monitoredShards) == 0 {
		return false, nil
	}

	for _, name := range shards {
		found := false
		for _, monitored := range monitoredShards {
			if monitored.Name == name {
				found = true
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// Monitor ensures that all the shards in the ShardedCluster object are monitored by the SentinelServer
func (ss *SentinelServer) Monitor(ctx context.Context, shards ShardedCluster) error {

	opt, err := redis.ParseURL(string(*ss))
	if err != nil {
		return err
	}

	sentinel := redis.NewSentinelClient(opt)

	// Initialize unmonitored shards
	for name, shard := range shards {

		result := &redistypes.SentinelMasterCmdResult{}
		err = sentinel.Master(ctx, name).Scan(result)
		if err != nil {
			if err.Error() == shardNotInitializedError {
				host, port, err := shard.GetMasterAddr()
				if err != nil {
					return err
				}
				_, err = sentinel.Monitor(ctx, name, host, port, "2").Result()
				if err != nil {
					return util.WrapError("[redis-sentinel/SentinelServer.Monitor]", err)
				}
				_, err = sentinel.Set(ctx, name, "down-after-milliseconds", "5000").Result()
				if err != nil {
					return util.WrapError("[redis-sentinel/SentinelServer.Monitor]", err)
				}

				// TODO: change the default failover timeout.
				// TODO: maybe add a generic mechanism to set/modify parameters

			} else {
				return err
			}
		}
	}

	return nil
}

// Masters executes the "sentinel masters" command against the given SentinelServer
func (ss *SentinelServer) Masters(ctx context.Context) ([]redistypes.SentinelMasterCmdResult, error) {
	opt, err := redis.ParseURL(string(*ss))
	if err != nil {
		return nil, err
	}

	sentinel := redis.NewSentinelClient(opt)

	values, err := sentinel.Masters(ctx).Result()
	if err != nil {
		return nil, err
	}

	result := make([]redistypes.SentinelMasterCmdResult, len(values))
	for i, val := range values {
		masterResult := &redistypes.SentinelMasterCmdResult{}
		sliceCmdToStruct(val, masterResult)
		if err != nil {
			return nil, err
		}
		result[i] = *masterResult
	}

	return result, nil
}

// Slaves executes the "sentinel slaves <shard>" against the given SentinelServer
func (ss *SentinelServer) Slaves(ctx context.Context, shard string) ([]redistypes.SentinelSlaveCmdResult, error) {
	opt, err := redis.ParseURL(string(*ss))
	if err != nil {
		return nil, err
	}

	sentinel := redis.NewSentinelClient(opt)

	values, err := sentinel.Slaves(ctx, shard).Result()
	if err != nil {
		return nil, err
	}

	result := make([]redistypes.SentinelSlaveCmdResult, len(values))
	for i, val := range values {
		slaveResult := &redistypes.SentinelSlaveCmdResult{}
		sliceCmdToStruct(val, slaveResult)
		result[i] = *slaveResult
	}

	return result, nil
}

// Subscribe watches for the given list of events generated by the SentinelServer
func (ss *SentinelServer) Subscribe(ctx context.Context, events ...string) (<-chan *redis.Message, func() error, error) {
	opt, err := redis.ParseURL(string(*ss))
	if err != nil {
		return nil, nil, err
	}

	sentinel := redis.NewSentinelClient(opt)
	pubsub := sentinel.PSubscribe(ctx, events...)
	return pubsub.Channel(), pubsub.Close, nil
}

// SentinelPool represents a pool of SentinelServers that monitor the same
// group of redis shards
type SentinelPool []SentinelServer

// NewSentinelPool creates a new SentinelPool object given a key and a number of replicas by calling the k8s API
// to discover sentinel Pods. The kye es the Name/Namespace of the StatefulSet that owns the sentinel Pods.
func NewSentinelPool(ctx context.Context, cl client.Client, key types.NamespacedName, replicas int) (SentinelPool, error) {

	spool := SentinelPool{}
	for i := 0; i < replicas; i++ {
		pod := &corev1.Pod{}
		key := types.NamespacedName{Name: fmt.Sprintf("%s-%d", key.Name, i), Namespace: key.Namespace}
		err := cl.Get(ctx, key, pod)
		if err != nil {
			return nil, err
		}
		spool = append(spool, SentinelServer(fmt.Sprintf("redis://%s:%d", pod.Status.PodIP, saasv1alpha1.SentinelPort)))
	}
	return spool, nil
}

// IsMonitoringShards checks whether or all the shards in the passed list are being monitored by all
// sentinel servers in the SentinelPool
func (sp SentinelPool) IsMonitoringShards(ctx context.Context, shards []string) (bool, error) {

	for _, connString := range sp {
		ss := SentinelServer(connString)
		ok, err := ss.IsMonitoringShards(ctx, shards)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// Monitor ensures that all the shards in the ShardedCluster object are monitored by
// all sentinel servers in the SentinelPool
func (sp SentinelPool) Monitor(ctx context.Context, shards ShardedCluster) error {

	for _, connString := range sp {
		ss := SentinelServer(connString)
		err := ss.Monitor(ctx, shards)
		if err != nil {
			return err
		}
	}
	return nil
}

// This is a horrible function to parse the horrible structs that the redis-go
// client returns for administrative commands. I swear it's not my fault ...
func sliceCmdToStruct(in interface{}, out interface{}) (interface{}, error) {
	m := map[string]string{}
	for i := range in.([]interface{}) {
		if i%2 != 0 {
			continue
		}
		m[in.([]interface{})[i].(string)] = in.([]interface{})[i+1].(string)
	}

	err := redis.NewStringStringMapResult(m, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
