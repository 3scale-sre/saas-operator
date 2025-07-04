package templates

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	envoy_serializer_v3 "github.com/3scale-sre/marin3r/api/envoy/serializer/v3"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	operatorutils "github.com/3scale-sre/saas-operator/internal/pkg/util"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func RouteConfiguration_v1(name string, opts any) (envoy.Resource, error) {
	o := opts.(*saasv1alpha1.RouteConfiguration)

	rc := &envoy_config_route_v3.RouteConfiguration{
		Name:         name,
		VirtualHosts: []*envoy_config_route_v3.VirtualHost{},
	}

	merr := operatorutils.MultiError{}

	for _, vhost := range o.VirtualHosts {
		vh := &envoy_config_route_v3.VirtualHost{}

		err := envoy_serializer_v3.JSON{}.Unmarshal(string(vhost.Raw), vh)
		if err != nil {
			merr = append(merr, err)
		}

		rc.VirtualHosts = append(rc.VirtualHosts, vh)
	}

	if len(merr) > 0 {
		return nil, merr
	}

	return rc, nil
}
