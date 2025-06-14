package envoyconfig

import (
	"errors"

	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_serializer "github.com/3scale-sre/marin3r/api/envoy/serializer"
	envoy_serializer_v3 "github.com/3scale-sre/marin3r/api/envoy/serializer/v3"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/auto"
	descriptor "github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/descriptor"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/factory"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func New(key types.NamespacedName, nodeID string, factory factory.EnvoyDynamicConfigFactory, resources ...descriptor.EnvoyDynamicConfigDescriptor) func(client.Object) (*marin3rv1alpha1.EnvoyConfig, error) {
	return func(client.Object) (*marin3rv1alpha1.EnvoyConfig, error) {
		protos := []envoy.Resource{}

		for _, res := range resources {
			proto, err := factory.NewResource(res)
			if err != nil {
				return nil, err
			}

			protos = append(protos, proto)
		}

		ec, err := newFromProtos(key, nodeID, protos)()
		if err != nil {
			return nil, err
		}

		return ec, nil
	}
}

func newFromProtos(key types.NamespacedName, nodeID string, resources []envoy.Resource) func() (*marin3rv1alpha1.EnvoyConfig, error) {
	return func() (*marin3rv1alpha1.EnvoyConfig, error) {
		clusters := []marin3rv1alpha1.EnvoyResource{}
		routes := []marin3rv1alpha1.EnvoyResource{}
		listeners := []marin3rv1alpha1.EnvoyResource{}
		runtimes := []marin3rv1alpha1.EnvoyResource{}

		secrets, err := auto.GenerateSecrets(resources)
		if err != nil {
			return nil, err
		}

		for i := range resources {
			j, err := envoy_serializer_v3.JSON{}.Marshal(resources[i])
			if err != nil {
				return nil, err
			}

			y, err := yaml.JSONToYAML([]byte(j))
			if err != nil {
				return nil, err
			}

			switch resources[i].(type) {
			case *envoy_config_cluster_v3.Cluster:
				clusters = append(clusters, marin3rv1alpha1.EnvoyResource{Value: string(y)})

			case *envoy_config_route_v3.RouteConfiguration:
				routes = append(routes, marin3rv1alpha1.EnvoyResource{Value: string(y)})

			case *envoy_config_listener_v3.Listener:
				listeners = append(listeners, marin3rv1alpha1.EnvoyResource{Value: string(y)})

			case *envoy_service_runtime_v3.Runtime:
				runtimes = append(runtimes, marin3rv1alpha1.EnvoyResource{Value: string(y)})

			default:
				return nil, errors.New("unknown dynamic configuration type")
			}
		}

		return &marin3rv1alpha1.EnvoyConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: marin3rv1alpha1.EnvoyConfigSpec{
				EnvoyAPI:      ptr.To(envoy.APIv3),
				NodeID:        nodeID,
				Serialization: ptr.To(envoy_serializer.YAML),
				EnvoyResources: &marin3rv1alpha1.EnvoyResources{
					Clusters:  clusters,
					Routes:    routes,
					Listeners: listeners,
					Runtimes:  runtimes,
					Secrets:   secrets,
				},
			},
		}, nil
	}
}
