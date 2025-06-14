package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type GoRedisClient struct {
	redis    *redis.Client
	sentinel *redis.SentinelClient
}

func NewFromConnectionString(connectionString string) (*GoRedisClient, error) {
	var err error

	c := &GoRedisClient{}

	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, err
	}

	// don't keep idle connections open
	opt.MinIdleConns = 0

	c.redis = redis.NewClient(opt)
	c.sentinel = redis.NewSentinelClient(opt)

	return c, nil
}

func MustNewFromConnectionString(connectionString string) *GoRedisClient {
	c, err := NewFromConnectionString(connectionString)
	if err != nil {
		panic(err)
	}

	return c
}

func NewFromOptions(opt *redis.Options) *GoRedisClient {
	return &GoRedisClient{
		redis:    redis.NewClient(opt),
		sentinel: redis.NewSentinelClient(opt),
	}
}

func (c *GoRedisClient) CloseRedis() error {
	return c.redis.Close()
}

func (c *GoRedisClient) CloseSentinel() error {
	return c.sentinel.Close()
}

func (c *GoRedisClient) Close() error {
	var firstErr error
	if err := c.CloseRedis(); err != nil {
		firstErr = err
	}

	if err := c.CloseSentinel(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

func (c *GoRedisClient) SentinelMaster(ctx context.Context, shard string) (*SentinelMasterCmdResult, error) {
	result := &SentinelMasterCmdResult{}
	err := c.sentinel.Master(ctx, shard).Scan(result)

	return result, err
}

func (c *GoRedisClient) SentinelGetMasterAddrByName(ctx context.Context, shard string) ([]string, error) {
	values, err := c.sentinel.GetMasterAddrByName(ctx, shard).Result()

	return values, err
}

func (c *GoRedisClient) SentinelMasters(ctx context.Context) ([]any, error) {
	values, err := c.sentinel.Masters(ctx).Result()

	return values, err
}

func (c *GoRedisClient) SentinelSlaves(ctx context.Context, shard string) ([]any, error) {
	values, err := c.sentinel.Slaves(ctx, shard).Result()

	return values, err
}

func (c *GoRedisClient) SentinelMonitor(ctx context.Context, name, host string, port string, quorum int) error {
	_, err := c.sentinel.Monitor(ctx, name, host, port, strconv.Itoa(quorum)).Result()

	return err
}

func (c *GoRedisClient) SentinelSet(ctx context.Context, shard, parameter, value string) error {
	_, err := c.sentinel.Set(ctx, shard, parameter, value).Result()

	return err
}

func (c *GoRedisClient) SentinelPSubscribe(ctx context.Context, events ...string) (<-chan *redis.Message, func() error) {
	pubsub := c.sentinel.PSubscribe(ctx, events...)

	return pubsub.Channel(), pubsub.Close
}

func (c *GoRedisClient) SentinelInfoCache(ctx context.Context) (any, error) {
	val, err := c.redis.Do(ctx, "sentinel", "info-cache").Result()

	return val, err
}

func (c *GoRedisClient) SentinelPing(ctx context.Context) error {
	_, err := c.sentinel.Ping(ctx).Result()

	return err
}

func (c *GoRedisClient) SentinelDo(ctx context.Context, args ...any) (any, error) {
	val, err := c.redis.Do(ctx, args...).Result()

	return val, err
}

func (c *GoRedisClient) RedisRole(ctx context.Context) (any, error) {
	val, err := c.redis.Do(ctx, "role").Result()

	return val, err
}

func (c *GoRedisClient) RedisConfigGet(ctx context.Context, parameter string) ([]any, error) {
	val, err := c.redis.ConfigGet(ctx, parameter).Result()

	return val, err
}

func (c *GoRedisClient) RedisConfigSet(ctx context.Context, parameter, value string) error {
	_, err := c.redis.ConfigSet(ctx, parameter, value).Result()

	return err
}

func (c *GoRedisClient) RedisSlaveOf(ctx context.Context, host, port string) error {
	_, err := c.redis.SlaveOf(ctx, host, port).Result()

	return err
}

// WARNING: this command blocks for the duration
func (c *GoRedisClient) RedisDebugSleep(ctx context.Context, duration time.Duration) error {
	_, err := c.redis.Do(ctx, "debug", "sleep", fmt.Sprintf("%.1f", duration.Seconds())).Result()

	return err
}

func (c *GoRedisClient) RedisDo(ctx context.Context, args ...any) (any, error) {
	val, err := c.redis.Do(ctx, args...).Result()

	return val, err
}

func (c *GoRedisClient) RedisBGSave(ctx context.Context) error {
	_, err := c.redis.BgSave(ctx).Result()

	return err
}

func (c *GoRedisClient) RedisLastSave(ctx context.Context) (int64, error) {
	return c.redis.LastSave(ctx).Result()
}

func (c *GoRedisClient) RedisSet(ctx context.Context, key string, value any) error {
	_, err := c.redis.Set(ctx, key, value, 0).Result()

	return err
}

func (c *GoRedisClient) RedisInfo(ctx context.Context, section string) (string, error) {
	val, err := c.redis.Info(ctx, section).Result()

	return val, err
}
