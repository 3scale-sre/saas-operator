package scheme

import (
	runtimeconfig "github.com/3scale-sre/basereconciler/runtimeconfig"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// +kubebuilder:scaffold:imports
)

func BuildAndRegister() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(saasv1alpha1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(grafanav1beta1.AddToScheme(scheme))
	utilruntime.Must(externalsecretsv1beta1.AddToScheme(scheme))
	utilruntime.Must(marin3rv1alpha1.AddToScheme(scheme))
	utilruntime.Must(pipelinev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme

	runtimeconfig.SetDefaultScheme(scheme)

	return scheme
}
