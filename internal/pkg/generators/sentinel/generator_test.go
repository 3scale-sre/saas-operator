package sentinel

import (
	"context"
	"testing"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

func TestGenerator_ClusterTopology(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		key     types.NamespacedName
		spec    saasv1alpha1.SentinelSpec
		args    args
		want    map[string]map[string]string
		wantErr bool
	}{
		{
			name: "Generates a correct cluster topology from 'spec.config.monitoredShards'",
			key:  types.NamespacedName{Name: "test", Namespace: "test"},
			spec: saasv1alpha1.SentinelSpec{
				Replicas: ptr.To[int32](3),
				Config: &saasv1alpha1.SentinelConfig{
					MonitoredShards: map[string][]string{
						"shard01": {
							"redis://localhost:1000",
							"redis://localhost:2000",
							"redis://localhost:3000",
						},
						"shard02": {
							"redis://localhost:4000",
							"redis://localhost:5000",
							"redis://localhost:6000",
						}}},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: map[string]map[string]string{
				"shard01": {
					"localhost:1000": "redis://127.0.0.1:1000",
					"localhost:2000": "redis://127.0.0.1:2000",
					"localhost:3000": "redis://127.0.0.1:3000",
				},
				"shard02": {
					"localhost:4000": "redis://127.0.0.1:4000",
					"localhost:5000": "redis://127.0.0.1:5000",
					"localhost:6000": "redis://127.0.0.1:6000",
				},
				"sentinel": {
					"redis-sentinel-0": "redis://redis-sentinel-0.test.svc.cluster.local:26379",
					"redis-sentinel-1": "redis://redis-sentinel-1.test.svc.cluster.local:26379",
					"redis-sentinel-2": "redis://redis-sentinel-2.test.svc.cluster.local:26379",
				},
			},
			wantErr: false,
		},
		{
			name: "Generates a correct cluster topology from 'spec.config.clusterTopology'",
			key:  types.NamespacedName{Name: "test", Namespace: "test"},
			spec: saasv1alpha1.SentinelSpec{
				Replicas: ptr.To(int32(3)),
				Config: &saasv1alpha1.SentinelConfig{
					ClusterTopology: map[string]map[string]string{
						"shard01": {
							"srv1": "redis://localhost:1000",
							"srv2": "redis://localhost:2000",
							"srv3": "redis://localhost:3000",
						},
						"shard02": {
							"srv4": "redis://localhost:4000",
							"srv5": "redis://localhost:5000",
							"srv6": "redis://localhost:6000",
						}}},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: map[string]map[string]string{
				"shard01": {
					"srv1": "redis://127.0.0.1:1000",
					"srv2": "redis://127.0.0.1:2000",
					"srv3": "redis://127.0.0.1:3000",
				},
				"shard02": {
					"srv4": "redis://127.0.0.1:4000",
					"srv5": "redis://127.0.0.1:5000",
					"srv6": "redis://127.0.0.1:6000",
				},
				"sentinel": {
					"redis-sentinel-0": "redis://redis-sentinel-0.test.svc.cluster.local:26379",
					"redis-sentinel-1": "redis://redis-sentinel-1.test.svc.cluster.local:26379",
					"redis-sentinel-2": "redis://redis-sentinel-2.test.svc.cluster.local:26379",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator("test", "test", tt.spec)

			got, err := gen.ClusterTopology(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.ClusterTopology() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if diff := deep.Equal(got, tt.want); len(diff) > 0 {
				t.Errorf("Generator.ClusterTopology() = got diff %v", diff)
			}
		})
	}
}
