package deployment

import (
	"errors"

	"github.com/3scale-sre/basereconciler/mutators"
	"github.com/3scale-sre/basereconciler/resource"
	"github.com/3scale-sre/basereconciler/util"
	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/envoyconfig/factory"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/hpa"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/pdb"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/podmonitor"
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/service"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(main DeploymentWorkload, canary DeploymentWorkload) ([]resource.TemplateInterface, error) {
	resources := workloadResources(main)

	if !lo.IsNil(canary) {
		resources = append(resources, workloadResources(canary)...)
	}

	services := []*resource.Template[*corev1.Service]{}

	// Generate resources to implement the desired publishing strategies
	if _, ok := main.(WithPublishingStrategies); ok {
		strategies, err := main.(WithPublishingStrategies).PublishingStrategies()
		if err != nil {
			return nil, err
		}

		for _, item := range strategies {
			descriptor := item
			switch descriptor.Strategy {
			case saasv1alpha1.SimpleStrategy:
				services = append(services,
					resource.NewTemplateFromObjectFunction(func() *corev1.Service { return descriptor.Service(main.GetKey().Name, "svc") }).
						WithMutation(mutators.SetServiceLiveValues()),
				)

			case saasv1alpha1.Marin3rSidecarStrategy:
				if descriptor.Marin3rSidecar == nil {
					return nil, errors.New("Marin3rSidecarSpec is missing, can't implement strategy without it")
				}

				services = append(services,
					resource.NewTemplateFromObjectFunction(func() *corev1.Service { return descriptor.Service(main.GetKey().Name, "marin3r") }).
						WithMutation(mutators.SetServiceLiveValues()),
				)

				// Add Marin3r sidecar to Deployment
				// NOTE: Deployment is always the first resource
				// NOTE2: a proper mechanism to identify resource templates should exists. Basereconciler
				// should provide one.
				if deployment, ok := resources[0].(*resource.Template[*appsv1.Deployment]); ok {
					deployment.Apply(marin3rSidecarToDeployment(descriptor))
				} else {
					return nil, errors.New("expected a Deployment but found something else")
				}
				// Add EnvoyConfig resource
				dynamicConfigurations := descriptor.Marin3rSidecar.EnvoyDynamicConfig.AsList()
				resources = append(resources,
					resource.NewTemplate(
						envoyconfig.New(EmptyKey, EmptyKey.Name, factory.Default(), dynamicConfigurations...)).
						WithEnabled(len(dynamicConfigurations) > 0).
						Apply(meta[*marin3rv1alpha1.EnvoyConfig](main)).
						Apply(nodeIdToEnvoyConfig(descriptor)),
				)
			}
		}
	}

	// Apply traffic routing logic (canary yes/no)
	for _, svct := range services {
		resources = append(resources,
			svct.Apply(meta[*corev1.Service](main)).
				Apply(trafficSelectorToService(main.(WithCanary), toWithCanaryOrNil(canary))),
		)
	}

	return resources, nil
}

var (
	EmptyKey      types.NamespacedName = types.NamespacedName{}
	EmptyLabel    map[string]string    = map[string]string{}
	EmptySelector map[string]string    = map[string]string{}
)

func workloadResources(workload DeploymentWorkload) []resource.TemplateInterface {
	resources := []resource.TemplateInterface{

		workload.Deployment().
			Apply(meta[*appsv1.Deployment](workload)).
			Apply(selector[*appsv1.Deployment](workload)).
			Apply(trafficSelectorToDeployment(workload)),

		resource.NewTemplate(
			pdb.New(EmptyKey, EmptyLabel, EmptySelector, *workload.PDBSpec())).
			WithEnabled(!workload.PDBSpec().IsDeactivated()).
			Apply(meta[*policyv1.PodDisruptionBudget](workload)).
			Apply(selector[*policyv1.PodDisruptionBudget](workload)),

		resource.NewTemplate(
			hpa.New(EmptyKey, EmptyLabel, *workload.HPASpec())).
			WithEnabled(!workload.HPASpec().IsDeactivated()).
			Apply(meta[*autoscalingv2.HorizontalPodAutoscaler](workload)).
			Apply(scaleTargetRefToHPA(workload)),

		resource.NewTemplate(
			podmonitor.New(EmptyKey, EmptyLabel, EmptySelector, workload.MonitoredEndpoints()...)).
			WithEnabled(len(workload.MonitoredEndpoints()) > 0).
			Apply(meta[*monitoringv1.PodMonitor](workload)).
			Apply(selector[*monitoringv1.PodMonitor](workload)),
	}

	return resources
}

func meta[T client.Object](w WithWorkloadMeta) resource.TemplateBuilderFunction[T] {
	return func(o client.Object) (T, error) {
		switch o.(type) {
		case *corev1.Service:
			// Do not enforce metadata.name:
			//   Services are special because there can be more than one of them, so the Name
			//   is relevant and must be provided by the service template
		default:
			o.SetName(w.GetKey().Name)
		}

		o.SetNamespace(w.GetKey().Namespace)
		o.SetLabels(util.MergeMaps(map[string]string{}, o.GetLabels(), w.GetLabels()))

		return o.(T), nil
	}
}

func trafficSelectorToService(main WithCanary, canary WithCanary) resource.TemplateBuilderFunction[*corev1.Service] {
	return func(o client.Object) (*corev1.Service, error) {
		svc := o.(*corev1.Service)
		svc.Spec.Selector = trafficSwitcher(main, canary)

		return svc, nil
	}
}

func trafficSwitcher(main WithCanary, canary WithCanary) map[string]string {
	// NOTE: due to the fact that services do not yet support set-based selectors, only MatchLabels selectors
	// can be used. This limits a lot what can be done in terms of deciding where to send traffic, as all
	// Deployments that should receive traffic need to have the same labels. The only way of doing this
	// without modifying the Deployment labels (which would trigger a rollout) and acting on the Service
	// selector alone is to choose only between three options:
	//                   traffic to noone / traffic to a single Deployment / traffic to all
	//
	// There seems to be great demand for set-based selectors for Services but it is not yet implamented:
	// https://github.com/kubernetes/kubernetes/issues/48528
	enabledSelectors := []map[string]string{}

	for _, workload := range []WithCanary{main, canary} {
		if workload != nil && workload.SendTraffic() {
			enabledSelectors = append(enabledSelectors, workload.GetSelector())
		}
	}

	switch c := len(enabledSelectors); c {
	case 0:
		return map[string]string{}
	case 1:
		// If there is only one Deployment with SendTraffic() active
		// return its selector together with the shared traffic selector
		return util.MergeMaps(map[string]string{}, enabledSelectors[0], main.TrafficSelector())
	default:
		// If there is more than one Deployment with SendTraffic() active
		// send traffic to all Deployments by using the shared traffic selector
		return util.MergeMaps(map[string]string{}, main.TrafficSelector())
	}
}

func scaleTargetRefToHPA(w WithWorkloadMeta) resource.TemplateBuilderFunction[*autoscalingv2.HorizontalPodAutoscaler] {
	return func(o client.Object) (*autoscalingv2.HorizontalPodAutoscaler, error) {
		hpa := o.(*autoscalingv2.HorizontalPodAutoscaler)
		hpa.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
			Kind:       "Deployment",
			Name:       w.GetKey().Name,
			APIVersion: appsv1.SchemeGroupVersion.String(),
		}

		return hpa, nil
	}
}

func selector[T client.Object](w DeploymentWorkload) resource.TemplateBuilderFunction[T] {
	return func(o client.Object) (T, error) {
		switch v := o.(type) {
		case *appsv1.Deployment:
			v.Spec.Selector = &metav1.LabelSelector{MatchLabels: w.GetSelector()}
			v.Spec.Template.ObjectMeta.Labels = util.MergeMaps(map[string]string{}, v.Spec.Template.ObjectMeta.Labels, w.GetLabels(), w.GetSelector())
		case *policyv1.PodDisruptionBudget:
			v.Spec.Selector = &metav1.LabelSelector{MatchLabels: w.GetSelector()}
		case *monitoringv1.PodMonitor:
			v.Spec.Selector = metav1.LabelSelector{MatchLabels: w.GetSelector()}
		}

		return o.(T), nil
	}
}

func trafficSelectorToDeployment(w DeploymentWorkload) resource.TemplateBuilderFunction[*appsv1.Deployment] {
	return func(o client.Object) (*appsv1.Deployment, error) {
		dep := o.(*appsv1.Deployment)
		if w, ok := w.(WithCanary); ok {
			dep.Spec.Template.ObjectMeta.Labels = util.MergeMaps(map[string]string{}, dep.Spec.Template.ObjectMeta.Labels, w.TrafficSelector())
		}

		return dep, nil
	}
}

func marin3rSidecarToDeployment(sd service.ServiceDescriptor) resource.TemplateBuilderFunction[*appsv1.Deployment] {
	return func(o client.Object) (*appsv1.Deployment, error) {
		return service.AddMarin3rSidecar(o.(*appsv1.Deployment), *sd.Marin3rSidecar), nil
	}
}

func nodeIdToEnvoyConfig(sd service.ServiceDescriptor) resource.TemplateBuilderFunction[*marin3rv1alpha1.EnvoyConfig] {
	return func(o client.Object) (*marin3rv1alpha1.EnvoyConfig, error) {
		ec := o.(*marin3rv1alpha1.EnvoyConfig)
		if sd.Marin3rSidecar.NodeID != nil {
			ec.Spec.NodeID = *sd.Marin3rSidecar.NodeID
		} else {
			ec.Spec.NodeID = ec.GetName()
		}

		return ec, nil
	}
}

func toWithCanaryOrNil(w DeploymentWorkload) WithCanary {
	if lo.IsNil(w) {
		return nil
	}

	return w.(WithCanary)
}
