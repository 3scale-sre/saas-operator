package templates

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

func Runtime_v1(name string, opts any) (envoy.Resource, error) {
	o := opts.(*saasv1alpha1.Runtime)

	layer, _ := structpb.NewStruct(map[string]any{
		"envoy": map[string]any{
			"resource_limits": map[string]any{
				"listener": func() map[string]any {
					m := map[string]any{}
					for _, name := range o.ListenerNames {
						m[name] = map[string]any{
							"connection_limit": 10000,
						}
					}

					return m
				}(),
			},
		},
		"overload": map[string]any{
			"global_downstream_max_connections": 50000,
		},
	})

	return &envoy_service_runtime_v3.Runtime{
		Name:  name,
		Layer: layer,
	}, nil
}
