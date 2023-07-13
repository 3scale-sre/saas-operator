package sharded

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/3scale/saas-operator/pkg/redis_v2/client"
	redis "github.com/3scale/saas-operator/pkg/redis_v2/server"
	"github.com/3scale/saas-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Shard is a list of the redis Server objects that compose a redis shard
type Shard struct {
	Name    string
	Servers []*RedisServer
	pool    *redis.ServerPool
}

// NewShard returns a Shard object given the passed redis server URLs
func NewShard(name string, servers map[string]string, pool *redis.ServerPool) (*Shard, error) {
	var merr util.MultiError
	shard := &Shard{Name: name, pool: pool}
	shard.Servers = make([]*RedisServer, 0, len(servers))

	for key, connectionString := range servers {
		var alias *string = nil
		if key != connectionString {
			alias = &key
		}
		srv, err := NewRedisServerFromPool(connectionString, alias, pool)
		if err != nil {
			merr = append(merr, err)
			continue
		}
		shard.Servers = append(shard.Servers, srv)
	}

	// sort the slice to obtain consistent results
	sort.Slice(shard.Servers, func(i, j int) bool {
		return shard.Servers[i].ID() < shard.Servers[j].ID()
	})

	return shard, merr.ErrorOrNil()
}

// Discover retrieves the options for all the servers in the shard
// If a SentinelServer is provided, it will be used to autodiscover servers and roles in the shard
func (shard *Shard) Discover(ctx context.Context, sentinel *SentinelServer, options ...DiscoveryOption) error {
	var merr util.MultiError
	logger := log.FromContext(ctx, "function", "(*Shard).Discover")

	switch sentinel {

	// no sentinel provided
	case nil:
		for idx := range shard.Servers {
			if err := shard.Servers[idx].Discover(ctx, options...); err != nil {
				logger.Error(err, "unable to discover redis server %s", shard.Servers[idx].ID())
				merr = append(merr, err)
				continue
			}
		}

	// sentinel provided
	default:
		sentinelMasterResult, err := sentinel.SentinelMaster(ctx, shard.Name)
		if err != nil {
			return append(merr, err)
		}

		// Get the corresponding server or add a new one if not found
		srv, err := shard.GetServerByID(fmt.Sprintf("%s:%d", sentinelMasterResult.IP, sentinelMasterResult.Port))
		if err != nil {
			return append(merr, err)
		}

		// do not try to discover a master flagged as "s_down" or "o_down"
		if strings.Contains(sentinelMasterResult.Flags, "s_down") || strings.Contains(sentinelMasterResult.Flags, "o_down") {
			return append(merr, fmt.Errorf("%s master %s is s_down/o_down", sentinelMasterResult.Name, fmt.Sprintf("%s:%d", sentinelMasterResult.IP, sentinelMasterResult.Port)))
		} else {
			// Confirm the server role
			if err = srv.Discover(ctx, options...); err != nil {
				srv.Role = client.Role(client.Unknown)
				return append(merr, err)
			} else if srv.Role != client.Master {
				// the role that the server reports is different from the role that
				// sentinel sees. Probably the sentinel configuration hasn't converged yet
				// this is an error and should be retried
				srv.Role = client.Role(client.Unknown)
				return append(merr, fmt.Errorf("sentinel config has not yet converged for %s", srv.ID()))
			}
		}

		// discover slaves
		sentinelSlavesResult, err := sentinel.SentinelSlaves(ctx, shard.Name)
		if err != nil {
			return append(merr, err)
		}
		for _, slave := range sentinelSlavesResult {

			// Get the corresponding server or add a new one if not found
			srv, err := shard.GetServerByID(fmt.Sprintf("%s:%d", slave.IP, slave.Port))
			if err != nil {
				merr = append(merr, err)
				continue
			}

			// do not try to discover a slave flagged as "s_down" or "o_down"
			if strings.Contains(slave.Flags, "s_down") || strings.Contains(slave.Flags, "o_down") {
				merr = append(merr, fmt.Errorf("%s slave %s is s_down/o_down", slave.Name, fmt.Sprintf("%s:%d", slave.IP, slave.Port)))
				continue
			} else {
				if err := srv.Discover(ctx, options...); err != nil {
					srv.Role = client.Role(client.Unknown)
					logger.Error(err, "unable to discover redis server %s", srv.ID())
					merr = append(merr, err)
					continue
				}
				if srv.Role != client.Slave {
					// the role that the server reports is different from the role that
					// sentinel sees. Probably the sentinel configuration hasn't converged yet
					// this is an error and should be retried
					srv.Role = client.Role(client.Unknown)
					merr = append(merr, fmt.Errorf("sentinel config has not yet converged for %s", srv.ID()))
					continue
				}
			}
		}
	}

	return merr.ErrorOrNil()
}

// GetMasterAddr returns the URL of the master server in a shard or error if zero
// or more than one master is found
func (shard *Shard) GetMasterAddr() (string, string, error) {
	master := []*RedisServer{}

	for _, srv := range shard.Servers {
		if srv.Role == client.Master {
			master = append(master, srv)
		}
	}

	if len(master) != 1 {
		return "", "", util.WrapError("(*Shard).GetMasterAddr", fmt.Errorf("wrong number of masters: %d != 1", len(master)))
	}

	ip, err := master[0].IP()
	if err != nil {
		return "", "", util.WrapError("(*Shard).GetMasterAddr", err)
	}
	return ip, master[0].GetPort(), nil
}

func (shard *Shard) GetServerByID(hostport string) (*RedisServer, error) {
	var rs *RedisServer
	var err error

	for _, srv := range shard.Servers {
		if srv.ID() == hostport {
			rs = srv
			break
		}
	}

	// If the server is not in the list, add a new one
	if rs == nil {
		rs, err = NewRedisServerFromPool("redis://"+hostport, nil, shard.pool)
		if err != nil {
			return nil, err
		}
		shard.Servers = append(shard.Servers, rs)
	}

	return rs, nil
}

// Init initializes the shard if not already initialized
func (shard *Shard) Init(ctx context.Context, masterIndex int32) ([]string, error) {
	logger := log.FromContext(ctx, "function", "(*Shard).Init")
	changed := []string{}

	for idx, srv := range shard.Servers {
		role, slaveof, err := srv.RedisRole(ctx)
		if err != nil {
			return changed, err
		}

		if role == client.Slave {

			if slaveof == "127.0.0.1" {

				if idx == int(masterIndex) {
					if err := srv.RedisSlaveOf(ctx, "NO", "ONE"); err != nil {
						return changed, err
					}
					logger.Info(fmt.Sprintf("configured %s as master", srv.ID()))
					changed = append(changed, srv.ID())
				} else {
					if err := srv.RedisSlaveOf(ctx, shard.Servers[masterIndex].GetHost(), shard.Servers[masterIndex].GetPort()); err != nil {
						return changed, err
					}
					logger.Info(fmt.Sprintf("configured %s as slave", srv.ID()))
					changed = append(changed, srv.ID())
				}

			} else {
				shard.Servers[idx].Role = client.Slave
			}

		} else if role == client.Master {
			shard.Servers[idx].Role = client.Master
		} else {
			return changed, fmt.Errorf("unable to get role for server %s", srv.ID())
		}
	}

	return changed, nil
}
