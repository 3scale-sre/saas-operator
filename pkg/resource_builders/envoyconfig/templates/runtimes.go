package templates

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

func Runtime_v1(name string, opts interface{}) (envoy.Resource, error) {
	o := opts.(*saasv1alpha1.Runtime)

	layer, _ := structpb.NewStruct(map[string]interface{}{
		"envoy": map[string]interface{}{
			"resource_limits": map[string]interface{}{
				"listener": func() map[string]interface{} {
					m := map[string]interface{}{}
					for _, name := range o.ListenerNames {
						m[name] = map[string]interface{}{
							"connection_limit": 10000,
						}
					}
					return m
				}(),
			},
		},
		"overload": map[string]interface{}{
			"global_downstream_max_connections": 50000,
		},
	})

	return &envoy_service_runtime_v3.Runtime{
		Name:  name,
		Layer: layer,
	}, nil
}
