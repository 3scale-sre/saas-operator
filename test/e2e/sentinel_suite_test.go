package e2e

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	redisclient "github.com/3scale-sre/saas-operator/internal/pkg/redis/client"
	testutil "github.com/3scale-sre/saas-operator/test/util"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("sentinel e2e suite", func() {
	var ns string
	var shards []saasv1alpha1.RedisShard
	var sentinel saasv1alpha1.Sentinel

	BeforeEach(func() {
		// Create a namespace for each block
		ns = "test-ns-" + nameGenerator.Generate()

		// Add any setup steps that needs to be executed before each test
		testNamespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		}

		err := k8sClient.Create(context.Background(), testNamespace)
		Expect(err).ToNot(HaveOccurred())

		n := &corev1.Namespace{}
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: ns}, n)
		}, timeout, poll).ShouldNot(HaveOccurred())

		shards = []saasv1alpha1.RedisShard{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "rs0", Namespace: ns},
				Spec:       saasv1alpha1.RedisShardSpec{MasterIndex: ptr.To[int32](0), SlaveCount: ptr.To[int32](2)},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "rs1", Namespace: ns},
				Spec:       saasv1alpha1.RedisShardSpec{MasterIndex: ptr.To[int32](2), SlaveCount: ptr.To[int32](2)},
			},
		}

		for i, shard := range shards {
			err = k8sClient.Create(context.Background(), &shard)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() error {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: shard.GetName(), Namespace: ns}, &shard)
				if err != nil {
					return err
				}
				if shard.Status.ShardNodes != nil && shard.Status.ShardNodes.Master != nil {
					// store the resource for later use
					shards[i] = shard
					GinkgoWriter.Printf("[debug] Shard %s topology: %+v\n", shard.GetName(), *shard.Status.ShardNodes)

					return nil
				} else {
					return fmt.Errorf("RedisShard %s not ready", shard.ObjectMeta.Name)
				}

			}, timeout, poll).ShouldNot(HaveOccurred())
		}
	})

	When("Sentinel resource is created and ready", func() {

		BeforeEach(func() {
			sentinel = saasv1alpha1.Sentinel{
				ObjectMeta: metav1.ObjectMeta{Name: "sentinel", Namespace: ns},
				Spec: saasv1alpha1.SentinelSpec{
					Config: &saasv1alpha1.SentinelConfig{
						MonitoredShards: map[string][]string{
							shards[0].GetName(): {
								"redis://" + shards[0].Status.ShardNodes.GetHostPortByPodIndex(0),
								"redis://" + shards[0].Status.ShardNodes.GetHostPortByPodIndex(1),
								"redis://" + shards[0].Status.ShardNodes.GetHostPortByPodIndex(2),
							},
							shards[1].GetName(): {
								"redis://" + shards[1].Status.ShardNodes.GetHostPortByPodIndex(0),
								"redis://" + shards[1].Status.ShardNodes.GetHostPortByPodIndex(1),
								"redis://" + shards[1].Status.ShardNodes.GetHostPortByPodIndex(2),
							},
						},
					},
				},
			}

			err := k8sClient.Create(context.Background(), &sentinel)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() error {

				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: sentinel.GetName(), Namespace: ns}, &sentinel)
				Expect(err).ToNot(HaveOccurred())

				if len(sentinel.Status.MonitoredShards) != len(shards) {
					return errors.New("sentinel not ready")
				}

				return nil
			}, timeout, poll).ShouldNot(HaveOccurred())

		})

		It("deploys sentinel Pods that monitor each of the redis shards", func() {

			By("issuing a 'sentinel masters' command to ensure all shards are monitored by sentinel", func() {

				sclient, stopCh, err := testutil.SentinelClient(cfg,
					types.NamespacedName{
						Name:      fmt.Sprintf("redis-sentinel-%d", rand.Intn(int(saasv1alpha1.SentinelDefaultReplicas))),
						Namespace: ns,
					})
				Expect(err).ToNot(HaveOccurred())
				defer close(stopCh)

				ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
				defer cancel()

				masters, err := sclient.SentinelMasters(ctx)
				Expect(err).ToNot(HaveOccurred())

				for _, shard := range shards {
					found := false
					for _, master := range masters {
						if strings.Contains(shard.Status.ShardNodes.MasterHostPort(), master.IP) {
							found = true

							break
						}
					}
					if found == false {
						Fail(fmt.Sprintf("master for shard %s not found in 'sentinel master' response", shard.GetName()))
					}
				}
			})
		})

		It("updates the resource status with the current status of each redis shard", func() {

			Eventually(func() error {

				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: sentinel.GetName(), Namespace: ns}, &sentinel)
				if err != nil {
					return err
				}

				if diff := cmp.Diff(sentinel.Status.MonitoredShards, saasv1alpha1.MonitoredShards{
					saasv1alpha1.MonitoredShard{
						Name: "rs0",
						Servers: map[string]saasv1alpha1.RedisServerDetails{
							shards[0].Status.ShardNodes.GetHostPortByPodIndex(0): {
								Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(0),
								Role:    redisclient.Master,
								Config:  map[string]string{"save": "", "slave-priority": "100"},
							},
							shards[0].Status.ShardNodes.GetHostPortByPodIndex(1): {
								Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(1),
								Role:    redisclient.Slave,
								Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
								Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
							},
							shards[0].Status.ShardNodes.GetHostPortByPodIndex(2): {
								Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(2),
								Role:    redisclient.Slave,
								Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
								Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
							},
						},
					},
					saasv1alpha1.MonitoredShard{
						Name: "rs1",
						Servers: map[string]saasv1alpha1.RedisServerDetails{
							shards[1].Status.ShardNodes.GetHostPortByPodIndex(0): {
								Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(0),
								Role:    redisclient.Slave,
								Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
								Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
							},
							shards[1].Status.ShardNodes.GetHostPortByPodIndex(1): {
								Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(1),
								Role:    redisclient.Slave,
								Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
								Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
							},
							shards[1].Status.ShardNodes.GetHostPortByPodIndex(2): {
								Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(2),
								Role:    redisclient.Master,
								Config:  map[string]string{"save": "", "slave-priority": "100"},
							},
						},
					},
				}); diff != "" {
					return fmt.Errorf("got unexpected sentinel status %s", diff)
				}

				return nil
			}, timeout, poll).ShouldNot(HaveOccurred())
		})

		When("a redis master is unavailable", func() {

			BeforeEach(func() {

				By("configuring shard0's slave 2 with priority '0' to ensure always the same slave gets promoted to master", func() {

					rclient, stopCh, err := testutil.RedisClient(cfg,
						types.NamespacedName{
							Name:      "redis-shard-rs0-2",
							Namespace: ns,
						})
					Expect(err).ToNot(HaveOccurred())
					defer close(stopCh)

					ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
					defer cancel()

					err = rclient.RedisConfigSet(ctx, "slave-priority", "0")
					Expect(err).ToNot(HaveOccurred())
				})

				By("making the shard0's master unavailable to force a failover", func() {

					go func() {
						defer GinkgoRecover()

						rclient, stopCh, err := testutil.RedisClient(cfg,
							types.NamespacedName{
								Name:      "redis-shard-rs0-0",
								Namespace: ns,
							})
						Expect(err).ToNot(HaveOccurred())
						defer close(stopCh)

						ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
						defer cancel()

						// a master is considered down after 5s, so we sleep the current master
						// for 10 seconds to simulate a failure and trigger a master failover
						rclient.RedisDebugSleep(ctx, 10*time.Second)
					}()
				})

			})

			It("triggers a failover", func() {

				rclient, stopCh, err := testutil.RedisClient(cfg,
					types.NamespacedName{
						Name:      "redis-shard-rs0-1",
						Namespace: ns,
					})
				Expect(err).ToNot(HaveOccurred())
				defer close(stopCh)

				Eventually(func() error {
					role, _, err := rclient.RedisRole(context.TODO())
					if err != nil {
						return err
					}
					if role != redisclient.Master {
						return fmt.Errorf("expected 'master' but got %s", role)
					}

					return nil
				}, timeout, poll).ShouldNot(HaveOccurred())

			})

			It("updates the status appropriately", func() {
				Eventually(func() error {

					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: sentinel.GetName(), Namespace: ns}, &sentinel)
					if err != nil {
						return err
					}

					if diff := cmp.Diff(sentinel.Status.MonitoredShards, saasv1alpha1.MonitoredShards{
						saasv1alpha1.MonitoredShard{
							Name: "rs0",
							Servers: map[string]saasv1alpha1.RedisServerDetails{
								shards[0].Status.ShardNodes.GetHostPortByPodIndex(0): {
									Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(0),
									Role:    redisclient.Slave,
									Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
									Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
								},
								shards[0].Status.ShardNodes.GetHostPortByPodIndex(1): {
									Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(1),
									Role:    redisclient.Master,
									Config:  map[string]string{"save": "", "slave-priority": "100"},
								},
								shards[0].Status.ShardNodes.GetHostPortByPodIndex(2): {
									Address: shards[0].Status.ShardNodes.GetHostPortByPodIndex(2),
									Role:    redisclient.Slave,
									Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "0"},
									Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
								},
							},
						},
						saasv1alpha1.MonitoredShard{
							Name: "rs1",
							Servers: map[string]saasv1alpha1.RedisServerDetails{
								shards[1].Status.ShardNodes.GetHostPortByPodIndex(0): {
									Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(0),
									Role:    redisclient.Slave,
									Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
									Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
								},
								shards[1].Status.ShardNodes.GetHostPortByPodIndex(1): {
									Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(1),
									Role:    redisclient.Slave,
									Config:  map[string]string{"save": "", "slave-read-only": "yes", "slave-priority": "100"},
									Info:    map[string]string{"replication": "master-link: up, sync-in-progress: no"},
								},
								shards[1].Status.ShardNodes.GetHostPortByPodIndex(2): {
									Address: shards[1].Status.ShardNodes.GetHostPortByPodIndex(2),
									Role:    redisclient.Master,
									Config:  map[string]string{"save": "", "slave-priority": "100"},
								},
							},
						},
					}); diff != "" {
						return fmt.Errorf("got unexpected sentinel status %s", diff)
					}

					return nil
				}, timeout, poll).ShouldNot(HaveOccurred())
			})
		})

	})

	AfterEach(func() {

		// Delete sentinel
		err := k8sClient.Delete(context.Background(), &sentinel, client.PropagationPolicy(metav1.DeletePropagationForeground))
		Expect(err).ToNot(HaveOccurred())

		// Delete redis shards
		for _, shard := range shards {
			err := k8sClient.Delete(context.Background(), &shard, client.PropagationPolicy(metav1.DeletePropagationForeground))
			Expect(err).ToNot(HaveOccurred())
		}

		// Delete the namespace
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		err = k8sClient.Delete(context.Background(), ns, client.PropagationPolicy(metav1.DeletePropagationForeground))
		Expect(err).ToNot(HaveOccurred())
	})

})
