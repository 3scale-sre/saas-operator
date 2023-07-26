package twemproxyconfig

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	basereconciler "github.com/3scale-ops/basereconciler/reconciler"
	basereconciler_resources "github.com/3scale-ops/basereconciler/resources"
	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/generators"
	"github.com/3scale/saas-operator/pkg/redis_v2/server"
	"github.com/3scale/saas-operator/pkg/redis_v2/sharded"
	"github.com/3scale/saas-operator/pkg/resource_builders/grafanadashboard"
	"github.com/3scale/saas-operator/pkg/resource_builders/twemproxy"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
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

	shardedCluster, err := sharded.NewShardedCluster(ctx, clustermap, pool)
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
		shardedCluster.Discover(ctx, sharded.OnlyMasterDiscoveryOpt)
		gen.masterTargets, err = gen.getMonitoredMasters(ctx, shardedCluster, log.WithName("masterTargets"))
		if err != nil {
			return Generator{}, err
		}

	case true:
		shardedCluster.Discover(ctx, sharded.SlaveReadOnlyDiscoveryOpt)
		gen.masterTargets, err = gen.getMonitoredMasters(ctx, shardedCluster, log.WithName("masterTargets"))
		if err != nil {
			return Generator{}, err
		}
		gen.slaverwTargets, err = gen.getMonitoredReadWriteSlavesWithFallbackToMasters(
			ctx, shardedCluster, log.WithName("slaverwTargets"),
		)
		if err != nil {
			return Generator{}, err
		}
	}

	return gen, nil
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
		uris = append(uris, fmt.Sprintf("redis://%s", address))
	}

	return uris, nil
}

func (gen *Generator) getMonitoredMasters(ctx context.Context,
	cluster *sharded.Cluster, log logr.Logger) (map[string]twemproxy.Server, error) {

	m := make(map[string]twemproxy.Server, len(cluster.Shards))
	for _, shard := range cluster.Shards {
		hostport, err := shard.GetMaster()
		if err != nil {
			return nil, err
		}
		m[shard.Name] = twemproxy.Server{
			Name:     shard.Name,
			Address:  hostport,
			Priority: 1,
		}
	}

	return m, nil
}

func (gen *Generator) getMonitoredReadWriteSlavesWithFallbackToMasters(ctx context.Context,
	cluster *sharded.Cluster, log logr.Logger) (map[string]twemproxy.Server, error) {

	m := make(map[string]twemproxy.Server, len(cluster.Shards))
	for _, shard := range cluster.Shards {

		if slavesRW := shard.GetSlavesRW(); len(slavesRW) > 0 {
			m[shard.Name] = twemproxy.Server{
				Name:     shard.Name,
				Address:  slavesRW[0],
				Priority: 1,
			}
			slaveRwConfigured.With(prometheus.Labels{"twemproxy_config": gen.InstanceName, "shard": shard.Name}).Set(1)

		} else {
			// Fall back to the master if there are no
			// available RW slaves for this shard
			hostport, err := shard.GetMaster()
			if err != nil {
				return nil, err
			}
			m[shard.Name] = twemproxy.Server{
				Name:     shard.Name,
				Address:  hostport,
				Priority: 1,
			}
			slaveRwConfigured.With(prometheus.Labels{"twemproxy_config": gen.InstanceName, "shard": shard.Name}).Set(0)
		}
	}

	return m, nil
}

// Returns the twemproxy config ConfigMap
func (gen *Generator) ConfigMap() basereconciler.Resource {
	return basereconciler_resources.ConfigMapTemplate{
		Template:  gen.configMap(true),
		IsEnabled: true,
	}
}

func (gen *Generator) GrafanaDashboard() basereconciler_resources.GrafanaDashboardTemplate {
	return basereconciler_resources.GrafanaDashboardTemplate{
		Template: grafanadashboard.New(types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", gen.InstanceName, gen.Component),
			Namespace: gen.Namespace,
		}, gen.GetLabels(), *gen.Spec.GrafanaDashboard, "dashboards/twemproxy.json.gtpl"),
		IsEnabled: !gen.Spec.GrafanaDashboard.IsDeactivated(),
	}
}
