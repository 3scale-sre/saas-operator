package config

import (
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/generators/seed"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pod"
)

type Secret string

func (s Secret) String() string { return string(s) }

const (
	CorsProxySystemDatabaseSecret Secret = "cors-proxy-system-database"
)

// NewOptions returns cors-proxy options the given saasv1alpha1.CORSProxySpec
func NewOptions(spec saasv1alpha1.CORSProxySpec) pod.Options {
	opts := pod.Options{}
	opts.AddEnvvar("DATABASE_URL").AsSecretRef(CorsProxySystemDatabaseSecret).WithSeedKey(seed.SystemDatabaseDsn).
		Unpack(spec.Config.SystemDatabaseDSN)

	return opts
}
