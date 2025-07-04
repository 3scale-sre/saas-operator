package sharded

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	redis "github.com/3scale-sre/saas-operator/internal/pkg/redis/server"
	operatorutils "github.com/3scale-sre/saas-operator/internal/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Cluster represents a sharded redis cluster, composed by several Shards
type Cluster struct {
	Shards    []*Shard
	Sentinels []*SentinelServer
	pool      *redis.ServerPool
}

func NewShardedCluster(ctx context.Context, pool *redis.ServerPool, sentinels map[string]string, shards ...*Shard) (*Cluster, error) {
	cluster := &Cluster{pool: pool}
	cluster.Shards = make([]*Shard, 0, len(shards))

	// populate sentinel servers
	has, err := NewHighAvailableSentinel(sentinels, pool)
	if err != nil {
		return nil, err
	}

	cluster.Sentinels = has

	// populate shards
	cluster.Shards = append(cluster.Shards, shards...)

	return cluster, nil
}

// NewShardedClusterFromTopology returns a new ShardedCluster given the shard structure passed as a map[string][]string
func NewShardedClusterFromTopology(ctx context.Context, serverList map[string]map[string]string, pool *redis.ServerPool) (*Cluster, error) {
	logger := log.FromContext(ctx, "function", "NewShardedClusterFromTopology")
	cluster := Cluster{pool: pool}
	cluster.Shards = make([]*Shard, 0, len(serverList))

	for shardName, shardServers := range serverList {
		switch shardName {
		case "sentinel":
			sentinels, err := NewHighAvailableSentinel(serverList["sentinel"], pool)
			if err != nil {
				return nil, err
			}

			cluster.Sentinels = sentinels

		default:
			shard, err := NewShardFromTopology(shardName, shardServers, pool)
			if err != nil {
				logger.Error(err, "unable to create sharded cluster")

				return nil, err
			}

			cluster.Shards = append(cluster.Shards, shard)
		}
	}

	// sort the slices to obtain consistent results
	sort.Slice(cluster.Shards, func(i, j int) bool {
		return cluster.Shards[i].Name < cluster.Shards[j].Name
	})
	sort.Slice(cluster.Sentinels, func(i, j int) bool {
		return cluster.Sentinels[i].ID() < cluster.Sentinels[j].ID()
	})

	return &cluster, nil
}

func (cluster *Cluster) GetShardNames() []string {
	shards := make([]string, len(cluster.Shards))
	for i, shard := range cluster.Shards {
		shards[i] = shard.Name
	}

	sort.Strings(shards)

	return shards
}

func (cluster *Cluster) LookupShardByName(name string) *Shard {
	for _, shard := range cluster.Shards {
		if shard.Name == name {
			return shard
		}
	}

	return nil
}

func (cluster *Cluster) LookupServerByID(hostport string) *RedisServer {
	for _, shard := range cluster.Shards {
		for _, srv := range shard.Servers {
			if hostport == srv.ID() {
				return srv
			}
		}
	}

	return nil
}

func (cluster *Cluster) GetPool() *redis.ServerPool {
	return cluster.pool
}

func (cluster *Cluster) Discover(ctx context.Context, options ...DiscoveryOption) error {
	var merr operatorutils.MultiError

	for _, shard := range cluster.Shards {
		if err := shard.Discover(ctx, nil, options...); err != nil {
			merr = append(merr, err)

			continue
		}
	}

	return merr.ErrorOrNil()
}

// Updates the status of the cluster as seen from sentinel
func (cluster *Cluster) SentinelDiscover(ctx context.Context, opts ...DiscoveryOption) error {
	merr := operatorutils.MultiError{}

	// Get a healthy sentinel server
	sentinel := cluster.GetSentinel(ctx)
	if sentinel == nil {
		return append(merr, errors.New("unable to find a healthy sentinel server"))
	}

	masters, err := sentinel.SentinelMasters(ctx)
	if err != nil {
		return append(merr, err)
	}

	for _, master := range masters {
		// Get the corresponding shard
		shard := cluster.LookupShardByName(master.Name)

		// Add the shard if not already present
		if shard == nil {
			shard = &Shard{
				Name:    master.Name,
				Servers: []*RedisServer{},
				pool:    cluster.pool,
			}
			cluster.Shards = append(cluster.Shards, shard)
		}

		if err := shard.Discover(ctx, sentinel, opts...); err != nil {
			merr = append(merr, ShardDiscoveryError{ShardName: master.Name, Errors: err.(operatorutils.MultiError)})
			// keep going with the other shards
			continue
		}
	}

	return merr.ErrorOrNil()
}

// GetSentinel returns a healthy SentinelServer from the list of sentinels
// Returns nil if no healthy SentinelServer was found
func (cluster *Cluster) GetSentinel(pctx context.Context) *SentinelServer {
	ctx, cancel := context.WithTimeout(pctx, 5*time.Second)
	defer cancel()

	ch := make(chan int)

	for idx := range cluster.Sentinels {
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					return
				}
			}()

			if err := cluster.Sentinels[i].SentinelPing(ctx); err == nil {
				ch <- i
			}
		}(idx)
	}

	select {
	case <-ctx.Done():
	case idx := <-ch:
		close(ch)

		return cluster.Sentinels[idx]
	}

	return nil
}

type ShardDiscoveryError struct {
	ShardName string
	Errors    operatorutils.MultiError
}

func (e ShardDiscoveryError) Error() string {
	return fmt.Sprintf("errors occurred for shard %s: '%s'", e.ShardName, e.Errors)
}

func (e ShardDiscoveryError) Unwrap() []error {
	return []error(e.Errors)
}
