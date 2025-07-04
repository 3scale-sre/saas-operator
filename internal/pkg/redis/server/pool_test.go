package server

import (
	"reflect"
	"testing"

	"k8s.io/utils/ptr"
)

func TestServerPool_GetServer(t *testing.T) {
	type fields struct {
		servers []*Server
	}

	type args struct {
		connectionString string
		alias            *string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Server
		wantErr bool
	}{
		{
			name: "Gets the server by hostport",
			fields: fields{
				servers: []*Server{
					{alias: "host1", client: nil, host: "127.0.0.1", port: "1000"},
					{alias: "host2", client: nil, host: "127.0.0.2", port: "2000"},
				}},
			args: args{
				connectionString: "redis://127.0.0.2:2000",
			},
			want:    &Server{alias: "host2", client: nil, host: "127.0.0.2", port: "2000"},
			wantErr: false,
		},
		{
			name: "Gets the server by hostport and sets the alias",
			fields: fields{
				servers: []*Server{
					{alias: "host1", client: nil, host: "127.0.0.1", port: "1000"},
					{alias: "", client: nil, host: "127.0.0.2", port: "2000"},
				}},
			args: args{
				connectionString: "redis://127.0.0.2:2000",
				alias:            ptr.To("host2"),
			},
			want:    &Server{alias: "host2", client: nil, host: "127.0.0.2", port: "2000"},
			wantErr: false,
		},
		{
			name: "Returns error",
			fields: fields{
				servers: []*Server{}},
			args: args{
				connectionString: "host",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &ServerPool{
				servers: tt.fields.servers,
			}

			got, err := pool.GetServer(tt.args.connectionString, tt.args.alias)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerPool.GetServer() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServerPool.GetServer() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Adds a new server to the pool", func(t *testing.T) {
		pool := &ServerPool{
			servers: []*Server{{alias: "host1", client: nil, port: "1000"}},
		}
		newSrv, _ := pool.GetServer("redis://127.0.0.2:2000", ptr.To("host2"))
		exists, _ := pool.GetServer("redis://127.0.0.2:2000", ptr.To("host2"))

		if newSrv != exists {
			t.Errorf("ServerPool.GetServer() = %v, want %v", newSrv, exists)
		}
	})
}

func TestServerPool_indexByHost(t *testing.T) {
	type fields struct {
		servers []*Server
	}

	tests := []struct {
		name   string
		fields fields
		want   map[string]*Server
	}{
		{
			name: "Returns a map indexed by host",
			fields: fields{
				servers: []*Server{
					{alias: "host1", client: nil, host: "127.0.0.1", port: "1000"},
					{alias: "host2", client: nil, host: "127.0.0.2", port: "2000"},
				},
			},
			want: map[string]*Server{
				"127.0.0.1:1000": {alias: "host1", client: nil, host: "127.0.0.1", port: "1000"},
				"127.0.0.2:2000": {alias: "host2", client: nil, host: "127.0.0.2", port: "2000"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &ServerPool{
				servers: tt.fields.servers,
			}
			if got := pool.indexByHostPort(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServerPool.indexByHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
