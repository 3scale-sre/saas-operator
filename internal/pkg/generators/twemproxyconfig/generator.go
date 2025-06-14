package twemproxyconfig

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/3scale-sre/basereconciler/resource"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/generators"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/server"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/sharded"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/grafanadashboard"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/twemproxy"
	"github.com/go-logr/logr"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	component string = "twemproxy"
)

var (
	slaveRwConfigured = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "slave_rw_configured",
			Namespace: "saas_twemproxyconfig",
			Help:      "1 if the TwemproxyConfig points to a RW slave, 0 otherwise",
		},
		[]string{"twemproxy_config", "shard"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(slaveRwConfigured)
}

// Generator configures the generators for Sentinel
type Generator struct {
	generators.BaseOptionsV2
	Spec           saasv1alpha1.TwemproxyConfigSpec
	masterTargets  map[string]twemproxy.Server
	slaverwTargets map[string]twemproxy.Server
}

// NewGenerator returns a new Options struct
func NewGenerator(ctx context.Context, instance *saasv1alpha1.TwemproxyConfig, cl client.Client,
	pool *server.ServerPool, log logr.Logger) (Generator, error) {
	gen := Generator{
		BaseOptionsV2: generators.BaseOptionsV2{
			Component:    component,
			InstanceName: instance.GetName(),
			Namespace:    instance.GetNamespace(),
			Labels: map[string]string{
				"app":     component,
				"part-of": "3scale-saas",
			},
		},
		Spec: instance.Spec,
	}

	var err error
	if gen.Spec.SentinelURIs == nil {
		gen.Spec.SentinelURIs, err = discoverSentinels(ctx, cl, instance.GetNamespace())
		if err != nil {
			return Generator{}, err
		}
	}

	clustermap := map[string]map[string]string{}
	clustermap["sentinel"] = make(map[string]string, len(gen.Spec.SentinelURIs))

	for _, uri := range gen.Spec.SentinelURIs {
		u, err := url.Parse(uri)
		if err != nil {
			return Generator{}, err
		}

		alias := strings.Split(u.Hostname(), ".")[0]
		clustermap["sentinel"][alias] = u.String()
	}

	shardedCluster, err := sharded.NewShardedClusterFromTopology(ctx, clustermap, pool)
	if err != nil {
		return Generator{}, err
	}

	// Check if there are pools in the config that require slave discovery
	discoverSlavesRW := false

	for _, pool := range gen.Spec.ServerPools {
		if *pool.Target == saasv1alpha1.SlavesRW {
			discoverSlavesRW = true
		}
	}

	switch discoverSlavesRW {
	case false:
		// any error discovering masters should return
		if merr := shardedCluster.SentinelDiscover(ctx, sharded.OnlyMasterDiscoveryOpt); merr != nil {
			return Generator{}, merr
		}

		gen.masterTargets, err = gen.getMonitoredMasters(shardedCluster)
		if err != nil {
			return Generator{}, err
		}

	case true:
		merr := shardedCluster.SentinelDiscover(ctx, sharded.SlaveReadOnlyDiscoveryOpt)
		if merr != nil {
			log.Error(merr, "DiscoveryError")
			// Only sentinel/master discovery errors should return.
			// Slave failures will just failover to the master without returning error (although it will be logged)
			sentinelError := &sharded.DiscoveryError_Sentinel_Failure{}
			masterError := &sharded.DiscoveryError_Master_SingleServerFailure{}

			if errors.As(merr, sentinelError) || errors.As(merr, masterError) {
				return Generator{}, merr
			}
		}

		gen.masterTargets, err = gen.getMonitoredMasters(shardedCluster)
		if err != nil {
			return Generator{}, err
		}

		gen.slaverwTargets, err = gen.getMonitoredReadWriteSlavesWithFallbackToMasters(shardedCluster)
		if err != nil {
			return Generator{}, err
		}
	}

	return gen, nil
}

func (gen *Generator) GetTargets(poolName string) map[string]twemproxy.Server {
	for _, pool := range gen.Spec.ServerPools {
		if pool.Name == poolName {
			if *pool.Target == saasv1alpha1.Masters {
				return gen.masterTargets
			} else {
				return gen.slaverwTargets
			}
		}
	}

	return nil
}

func discoverSentinels(ctx context.Context, cl client.Client, namespace string) ([]string, error) {
	sl := &saasv1alpha1.SentinelList{}
	if err := cl.List(ctx, sl, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(sl.Items) != 1 {
		return nil, fmt.Errorf("unexpected number (%d) of Sentinel resources in namespace", len(sl.Items))
	}

	uris := make([]string, 0, len(sl.Items[0].Status.Sentinels))
	for _, address := range sl.Items[0].Status.Sentinels {
		uris = append(uris, "redis://"+address)
	}

	return uris, nil
}

func (gen *Generator) getMonitoredMasters(
	cluster *sharded.Cluster) (map[string]twemproxy.Server, error) {
	m := make(map[string]twemproxy.Server, len(cluster.Shards))

	for _, shard := range cluster.Shards {
		master, err := shard.GetMaster()
		if err != nil {
			return nil, err
		}

		m[shard.Name] = twemproxy.Server{
			Address:  master.ID(),
			Priority: 1,
		}
		m[shard.Name] = twemproxy.NewServer(master.ID(), master.GetAlias())
	}

	return m, nil
}

func (gen *Generator) getMonitoredReadWriteSlavesWithFallbackToMasters(
	cluster *sharded.Cluster) (map[string]twemproxy.Server, error) {
	m := make(map[string]twemproxy.Server, len(cluster.Shards))

	for _, shard := range cluster.Shards {
		if slavesRW := shard.GetSlavesRW(); len(slavesRW) > 0 {
			m[shard.Name] = twemproxy.NewServer(slavesRW[0].ID(), slavesRW[0].GetAlias())

			slaveRwConfigured.With(prometheus.Labels{"twemproxy_config": gen.InstanceName, "shard": shard.Name}).Set(1)
		} else {
			// Fall back to the master if there are no
			// available RW slaves for this shard
			master, err := shard.GetMaster()
			if err != nil {
				return nil, err
			}

			m[shard.Name] = twemproxy.NewServer(master.ID(), master.GetAlias())

			slaveRwConfigured.With(prometheus.Labels{"twemproxy_config": gen.InstanceName, "shard": shard.Name}).Set(0)
		}
	}

	return m, nil
}

// Returns the twemproxy config ConfigMap
func (gen *Generator) ConfigMap() *resource.Template[*corev1.ConfigMap] {
	return resource.NewTemplateFromObjectFunction(func() *corev1.ConfigMap { return gen.configMap(true) })
}

func (gen *Generator) GrafanaDashboard() *resource.Template[*grafanav1beta1.GrafanaDashboard] {
	return resource.NewTemplate(
		grafanadashboard.New(types.NamespacedName{Name: fmt.Sprintf("%s-%s", gen.InstanceName, gen.Component), Namespace: gen.Namespace},
			gen.GetLabels(), *gen.Spec.GrafanaDashboard, "dashboards/twemproxy.json.gtpl")).
		WithEnabled(!gen.Spec.GrafanaDashboard.IsDeactivated())
}
