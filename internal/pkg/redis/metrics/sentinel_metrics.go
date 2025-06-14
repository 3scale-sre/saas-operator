package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/3scale-sre/saas-operator/internal/pkg/reconcilers/threads"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/client"
	redis "github.com/3scale-sre/saas-operator/internal/pkg/redis/server"
	"github.com/3scale-sre/saas-operator/internal/pkg/redis/sharded"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	linkPendingCommands = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "link_pending_commands",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel master <name> link-pending-commands"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)
	lastOkPingReply = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "last_ok_ping_reply",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel master <name> last-ok-ping-reply"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)
	roleReportedTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "role_reported_time",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel master <name> role-reported-time"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)
	numOtherSentinels = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_other_sentinels",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel master <name> num-other-sentinels"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)

	masterLinkDownTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "master_link_down_time",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel slaves master-link-down-time"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)

	slaveReplOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "slave_repl_offset",
			Namespace: "saas_redis_sentinel",
			Help:      `"sentinel slaves slave-repl-offset"`,
		},
		[]string{"sentinel", "shard", "redis_server_host", "redis_server_alias", "role"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		linkPendingCommands, lastOkPingReply, roleReportedTime,
		numOtherSentinels, masterLinkDownTime, slaveReplOffset,
	)
}

// SentinelEventWatcher implements RunnableThread
var _ threads.RunnableThread = &SentinelMetricsGatherer{}

// SentinelMetricsGatherer is used to export sentinel metrics, obtained
// thrugh several admin commands, as prometheus metrics
type SentinelMetricsGatherer struct {
	refreshInterval time.Duration
	sentinelURI     string
	sentinel        *sharded.SentinelServer
	serverPool      *redis.ServerPool
	started         bool
	cancel          context.CancelFunc
}

func NewSentinelMetricsGatherer(sentinelURI string, refreshInterval time.Duration, pool *redis.ServerPool) (*SentinelMetricsGatherer, error) {
	sentinel, err := sharded.NewSentinelServerFromPool(sentinelURI, nil, pool)
	if err != nil {
		return nil, err
	}

	return &SentinelMetricsGatherer{
		refreshInterval: refreshInterval,
		sentinelURI:     sentinelURI,
		sentinel:        sentinel,
		serverPool:      pool,
	}, nil
}

func (fw *SentinelMetricsGatherer) GetID() string {
	return fw.sentinelURI
}

// IsStarted returns whether the metrics gatherer is running or not
func (smg *SentinelMetricsGatherer) IsStarted() bool {
	return smg.started
}

func (smg *SentinelMetricsGatherer) CanBeDeleted() bool {
	return true
}

// SetChannel is required for SentinelMetricsGatherer to implement the RunnableThread
// interface, but it actually does nothing with the channel.
func (fw *SentinelMetricsGatherer) SetChannel(ch chan event.GenericEvent) {}

// Start starts metrics gatherer for sentinel
func (smg *SentinelMetricsGatherer) Start(parentCtx context.Context, l logr.Logger) error {
	log := l.WithValues("sentinel", smg.sentinelURI)
	if smg.started {
		log.Info("the metrics gatherer is already running")

		return nil
	}

	go func() {
		var ctx context.Context
		ctx, smg.cancel = context.WithCancel(parentCtx)

		ticker := time.NewTicker(smg.refreshInterval)

		log.Info("sentinel metrics gatherer running")

		for {
			select {
			case <-ticker.C:
				if err := smg.gatherMetrics(ctx); err != nil {
					log.Error(err, "error gathering sentinel metrics")
				}

			case <-ctx.Done():
				log.Info("shutting down sentinel metrics gatherer")

				smg.started = false

				return
			}
		}
	}()

	smg.started = true

	return nil
}

// Stop stops metrics gatherering for sentinel
func (smg *SentinelMetricsGatherer) Stop() {
	// stop gathering metrics
	smg.cancel()
	// Reset all gauge metrics so the values related to
	// this exporter are deleted from the collection
	linkPendingCommands.Reset()
	lastOkPingReply.Reset()
	roleReportedTime.Reset()
	numOtherSentinels.Reset()
	masterLinkDownTime.Reset()
	slaveReplOffset.Reset()
}

func (smg *SentinelMetricsGatherer) gatherMetrics(ctx context.Context) error {
	mresult, err := smg.sentinel.SentinelMasters(ctx)
	if err != nil {
		return err
	}

	for _, master := range mresult {
		masterServerHost := fmt.Sprintf("%s:%d", master.IP, master.Port)
		masterServerAlias := smg.serverPool.GetServerAlias(masterServerHost)

		linkPendingCommands.With(prometheus.Labels{
			"sentinel":           smg.sentinelURI,
			"shard":              master.Name,
			"redis_server_host":  masterServerHost,
			"redis_server_alias": masterServerAlias,
			"role":               master.RoleReported,
		}).Set(float64(master.LinkPendingCommands))

		lastOkPingReply.With(prometheus.Labels{
			"sentinel":           smg.sentinelURI,
			"shard":              master.Name,
			"redis_server_host":  masterServerHost,
			"redis_server_alias": masterServerAlias,
			"role":               master.RoleReported,
		}).Set(float64(master.LastOkPingReply))

		roleReportedTime.With(prometheus.Labels{
			"sentinel":           smg.sentinelURI,
			"shard":              master.Name,
			"redis_server_host":  masterServerHost,
			"redis_server_alias": masterServerAlias,
			"role":               master.RoleReported,
		}).Set(float64(master.RoleReportedTime))

		numOtherSentinels.With(prometheus.Labels{
			"sentinel":           smg.sentinelURI,
			"shard":              master.Name,
			"redis_server_host":  masterServerHost,
			"redis_server_alias": masterServerAlias,
			"role":               master.RoleReported,
		}).Set(float64(master.NumOtherSentinels))

		sresult, err := smg.sentinel.SentinelSlaves(ctx, master.Name)
		if err != nil {
			return err
		}

		// Cleanup any vector that corresponds to the same server but with a
		// different role to avoid stale metrics after a role switch
		cleanupMetrics(prometheus.Labels{
			"sentinel":           smg.sentinelURI,
			"shard":              master.Name,
			"redis_server_host":  masterServerHost,
			"redis_server_alias": masterServerAlias,
			"role":               string(client.Slave),
		})

		for _, slave := range sresult {
			slaveServerHost := fmt.Sprintf("%s:%d", slave.IP, slave.Port)
			slaveServerAlias := smg.serverPool.GetServerAlias(slaveServerHost)

			linkPendingCommands.With(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               slave.RoleReported,
			}).Set(float64(slave.LinkPendingCommands))

			lastOkPingReply.With(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               slave.RoleReported,
			}).Set(float64(slave.LastOkPingReply))

			roleReportedTime.With(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               slave.RoleReported,
			}).Set(float64(slave.RoleReportedTime))

			masterLinkDownTime.With(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               slave.RoleReported,
			}).Set(float64(slave.MasterLinkDownTime))

			slaveReplOffset.With(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               slave.RoleReported,
			}).Set(float64(slave.SlaveReplOffset))

			cleanupMetrics(prometheus.Labels{
				"sentinel":           smg.sentinelURI,
				"shard":              master.Name,
				"redis_server_host":  slaveServerHost,
				"redis_server_alias": slaveServerAlias,
				"role":               string(client.Master),
			})
		}
	}

	return nil
}

func cleanupMetrics(labels prometheus.Labels) {
	linkPendingCommands.Delete(labels)
	lastOkPingReply.Delete(labels)
	roleReportedTime.Delete(labels)
	numOtherSentinels.Delete(labels)
	masterLinkDownTime.Delete(labels)
	slaveReplOffset.Delete(labels)
}
