package client

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type FakeResponse struct {
	InjectResponse func() any
	InjectError    func() error
}

// Some predefined responses used in many tests
func NewPredefinedRedisFakeResponse(dictionary string, err error) FakeResponse {
	var rsp []any

	switch dictionary {
	case "save":
		rsp = []any{"save", "900 1 300 10"}
	case "no-save":
		rsp = []any{"save", ""}
	case "slave-read-only-no":
		rsp = []any{"read-only", "no"}
	case "slave-read-only-yes":
		rsp = []any{"read-only", "yes"}
	case "role-slave":
		rsp = []any{"slave", "127.0.0.1:3333"}
	case "role-master":
		rsp = []any{"master", ""}
	default:
		panic("response not defined")
	}

	return FakeResponse{
		// cmd: RedisConfigGet("save")
		InjectResponse: func() any {
			return rsp
		},
		InjectError: func() error { return err },
	}
}

type FakeClient struct {
	Responses []FakeResponse
}

func NewFakeClient(responses ...FakeResponse) TestableInterface {
	return &FakeClient{
		Responses: responses,
	}
}

func (fc *FakeClient) SentinelMaster(ctx context.Context, shard string) (*SentinelMasterCmdResult, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().(*SentinelMasterCmdResult), rsp.InjectError()
}

func (fc *FakeClient) SentinelGetMasterAddrByName(ctx context.Context, shard string) ([]string, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().([]string), rsp.InjectError()
}

func (fc *FakeClient) SentinelMasters(ctx context.Context) ([]any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().([]any), rsp.InjectError()
}

func (fc *FakeClient) SentinelSlaves(ctx context.Context, shard string) ([]any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().([]any), rsp.InjectError()
}

func (fc *FakeClient) SentinelMonitor(ctx context.Context, name, host string, port string, quorum int) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) SentinelSet(ctx context.Context, shard, parameter, value string) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) SentinelPSubscribe(ctx context.Context, events ...string) (<-chan *redis.Message, func() error) {
	rsp := fc.pop()

	return rsp.InjectResponse().(<-chan *redis.Message), nil
}

func (fc *FakeClient) SentinelInfoCache(ctx context.Context) (any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse(), rsp.InjectError()
}

func (fc *FakeClient) SentinelPing(ctx context.Context) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) SentinelDo(ctx context.Context, args ...any) (any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse(), rsp.InjectError()
}

func (fc *FakeClient) RedisRole(ctx context.Context) (any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse(), rsp.InjectError()
}

func (fc *FakeClient) RedisConfigGet(ctx context.Context, parameter string) ([]any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().([]any), rsp.InjectError()
}

func (fc *FakeClient) RedisConfigSet(ctx context.Context, parameter, value string) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) RedisSlaveOf(ctx context.Context, host, port string) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

// WARNING: this command blocks for the duration
func (fc *FakeClient) RedisDebugSleep(ctx context.Context, duration time.Duration) error {
	rsp := fc.pop()
	if rsp.InjectError() != nil {
		return rsp.InjectError()
	}

	time.Sleep(duration)

	return nil
}

func (fc *FakeClient) RedisDo(ctx context.Context, args ...any) (any, error) {
	rsp := fc.pop()

	return rsp.InjectResponse(), rsp.InjectError()
}

func (fc *FakeClient) RedisBGSave(ctx context.Context) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) RedisLastSave(ctx context.Context) (int64, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().(int64), rsp.InjectError()
}

func (fc *FakeClient) RedisSet(ctx context.Context, key string, value any) error {
	rsp := fc.pop()

	return rsp.InjectError()
}

func (fc *FakeClient) RedisInfo(ctx context.Context, section string) (string, error) {
	rsp := fc.pop()

	return rsp.InjectResponse().(string), rsp.InjectError()
}

func (fc *FakeClient) pop() (fakeRsp FakeResponse) {
	fakeRsp, fc.Responses = fc.Responses[0], fc.Responses[1:]

	return fakeRsp
}

func (fc *FakeClient) Close() error {
	return nil
}
