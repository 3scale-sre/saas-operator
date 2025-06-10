package auto

import (
	"github.com/3scale-sre/marin3r/api/envoy"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/samber/lo"
)

func GenerateSecrets(resources []envoy.Resource) ([]marin3rv1alpha1.EnvoySecretResource, error) {
	refs := []string{}

	for _, res := range resources {
		switch o := res.(type) {
		case *envoy_config_listener_v3.Listener:
			secrets, err := secretRefsFromListener(o)
			if err != nil {
				return nil, err
			}

			refs = append(refs, secrets...)
		}
	}

	secrets := []marin3rv1alpha1.EnvoySecretResource{}
	for _, ref := range lo.Uniq(refs) {
		secrets = append(secrets, marin3rv1alpha1.EnvoySecretResource{Name: ref})
	}

	return secrets, nil
}

func secretRefsFromListener(listener *envoy_config_listener_v3.Listener) ([]string, error) {
	if listener.GetFilterChains()[0].GetTransportSocket() == nil {
		return nil, nil
	}

	secrets := []string{}

	proto, err := listener.GetFilterChains()[0].GetTransportSocket().GetTypedConfig().UnmarshalNew()
	if err != nil {
		return nil, err
	}

	tlsContext := proto.(*envoy_extensions_transport_sockets_tls_v3.DownstreamTlsContext)
	for _, sdsConfig := range tlsContext.GetCommonTlsContext().GetTlsCertificateSdsSecretConfigs() {
		secrets = append(secrets, sdsConfig.GetName())
	}

	return lo.Uniq(secrets), nil
}
