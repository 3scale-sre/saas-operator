package auto

import (
	"reflect"
	"testing"

	"github.com/3scale-sre/marin3r/api/envoy"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/templates"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"k8s.io/utils/ptr"
)

func Test_secretRefsFromListener(t *testing.T) {
	type args struct {
		listener *envoy_config_listener_v3.Listener
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "returns the list of secrets used by the listener",
			args: args{
				listener: func() *envoy_config_listener_v3.Listener {
					l, _ := templates.ListenerHTTP_v1("test", &saasv1alpha1.ListenerHttp{
						Port:                  8080,
						RouteConfigName:       "my_route",
						CertificateSecretName: ptr.To("my_certificate"),
						EnableHttp2:           ptr.To(false),
						ProxyProtocol:         ptr.To(false),
					})

					return l.(*envoy_config_listener_v3.Listener)
				}(),
			},
			want:    []string{"my_certificate"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := secretRefsFromListener(tt.args.listener)
			if (err != nil) != tt.wantErr {
				t.Errorf("secretRefsFromListener() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("secretRefsFromListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSecrets(t *testing.T) {
	type args struct {
		resources []envoy.Resource
	}

	tests := []struct {
		name    string
		args    args
		want    []marin3rv1alpha1.EnvoySecretResource
		wantErr bool
	}{
		{
			name: "Generates envoy secret resources",
			args: args{
				resources: []envoy.Resource{
					func() envoy.Resource {
						l, _ := templates.ListenerHTTP_v1("test1", &saasv1alpha1.ListenerHttp{
							Port:                  8080,
							RouteConfigName:       "my_route",
							CertificateSecretName: ptr.To("cert1"),
							EnableHttp2:           ptr.To(false),
							ProxyProtocol:         ptr.To(false),
						})

						return l
					}(),
					func() envoy.Resource {
						l, _ := templates.ListenerHTTP_v1("test2", &saasv1alpha1.ListenerHttp{
							Port:                  8081,
							RouteConfigName:       "my_route",
							CertificateSecretName: ptr.To("cert2"),
							EnableHttp2:           ptr.To(false),
							ProxyProtocol:         ptr.To(false),
						})

						return l
					}(),
					func() envoy.Resource {
						l, _ := templates.ListenerHTTP_v1("test3", &saasv1alpha1.ListenerHttp{
							Port:                  8082,
							RouteConfigName:       "my_route",
							CertificateSecretName: ptr.To("cert1"),
							EnableHttp2:           ptr.To(false),
							ProxyProtocol:         ptr.To(false),
						})

						return l
					}(),
				},
			},
			want: []marin3rv1alpha1.EnvoySecretResource{
				{Name: "cert1"},
				{Name: "cert2"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSecrets(tt.args.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecrets() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}
