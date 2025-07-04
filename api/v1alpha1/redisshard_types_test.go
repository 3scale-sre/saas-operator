/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"
)

func TestRedisShardNodes_GetNodeByPodIndex(t *testing.T) {
	type fields struct {
		Master map[string]string
		Slaves map[string]string
	}

	type args struct {
		podIndex int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  string
	}{
		{
			name: "Returns the node that has the given pod index",
			fields: fields{
				Master: map[string]string{
					"redis-shard-rs0-0": "127.0.0.1:1000",
				},
				Slaves: map[string]string{
					"redis-shard-rs0-1": "127.0.0.1:2000",
					"redis-shard-rs0-2": "127.0.0.1:3000",
				},
			},
			args: args{
				podIndex: 2,
			},
			want:  "redis-shard-rs0-2",
			want1: "127.0.0.1:3000",
		},
		{
			name: "Not found",
			fields: fields{
				Master: map[string]string{
					"redis-shard-rs0-0": "127.0.0.1:1000",
				},
				Slaves: map[string]string{
					"redis-shard-rs0-1": "127.0.0.1:2000",
					"redis-shard-rs0-2": "127.0.0.1:3000",
				},
			},
			args: args{
				podIndex: 3,
			},
			want:  "",
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsn := &RedisShardNodes{
				Master: tt.fields.Master,
				Slaves: tt.fields.Slaves,
			}

			got, got1 := rsn.GetNodeByPodIndex(tt.args.podIndex)
			if got != tt.want {
				t.Errorf("RedisShardNodes.GetNodeByPodIndex() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("RedisShardNodes.GetNodeByPodIndex() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRedisShardNodes_MasterHostPort(t *testing.T) {
	type fields struct {
		Master map[string]string
		Slaves map[string]string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Returns master hostport",
			fields: fields{
				Master: map[string]string{"rs0-0": "127.0.0.1:1000"},
			},
			want: "127.0.0.1:1000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsn := &RedisShardNodes{
				Master: tt.fields.Master,
				Slaves: tt.fields.Slaves,
			}
			if got := rsn.MasterHostPort(); got != tt.want {
				t.Errorf("RedisShardNodes.MasterHostPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisShardNodes_GetHostPortByPodIndex(t *testing.T) {
	type fields struct {
		Master map[string]string
		Slaves map[string]string
	}

	type args struct {
		podIndex int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Returns the hosport of the node with the given replicaset index",
			fields: fields{
				Master: map[string]string{"rs0-0": "127.0.0.1:1000"},
				Slaves: map[string]string{"rs0-1": "127.0.0.1:2000", "rs0-2": "127.0.0.1:3000"},
			},
			args: args{podIndex: 2},
			want: "127.0.0.1:3000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsn := &RedisShardNodes{
				Master: tt.fields.Master,
				Slaves: tt.fields.Slaves,
			}
			if got := rsn.GetHostPortByPodIndex(tt.args.podIndex); got != tt.want {
				t.Errorf("RedisShardNodes.GetHostPortByPodIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisShardNodes_GetAliasByPodIndex(t *testing.T) {
	type fields struct {
		Master map[string]string
		Slaves map[string]string
	}

	type args struct {
		podIndex int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Returns the alias of the node with the given replicaset index",
			fields: fields{
				Master: map[string]string{"rs0-0": "127.0.0.1:1000"},
				Slaves: map[string]string{"rs0-1": "127.0.0.1:2000", "rs0-2": "127.0.0.1:3000"},
			},
			args: args{podIndex: 1},
			want: "rs0-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsn := &RedisShardNodes{
				Master: tt.fields.Master,
				Slaves: tt.fields.Slaves,
			}
			if got := rsn.GetAliasByPodIndex(tt.args.podIndex); got != tt.want {
				t.Errorf("RedisShardNodes.GetAliasByPodIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisShardNodes_GetIndexByHostPort(t *testing.T) {
	type fields struct {
		Master map[string]string
		Slaves map[string]string
	}

	type args struct {
		hostport string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Returns the index of the provided hostport",
			fields: fields{
				Master: map[string]string{"rs0-0": "127.0.0.1:1000"},
				Slaves: map[string]string{"rs0-1": "127.0.0.1:2000", "rs0-2": "127.0.0.1:3000"},
			},
			args: args{
				hostport: "127.0.0.1:3000",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsn := &RedisShardNodes{
				Master: tt.fields.Master,
				Slaves: tt.fields.Slaves,
			}
			if got := rsn.GetIndexByHostPort(tt.args.hostport); got != tt.want {
				t.Errorf("RedisShardNodes.GetIndexByHostPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
