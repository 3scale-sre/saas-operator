package system

import (
	"github.com/3scale-sre/saas-operator/internal/pkg/resource_builders/twemproxy"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func (gen *SystemTektonGenerator) task() *pipelinev1.Task {
	task := &pipelinev1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.GetComponent(),
			Namespace: gen.GetNamespace(),
			Labels:    gen.GetLabels(),
		},
		Spec: pipelinev1.TaskSpec{
			DisplayName: gen.GetComponent(),
			Description: *gen.Spec.Description,
			Params: []pipelinev1.ParamSpec{
				{
					Name:        "container-image",
					Description: "Container image for the task",
					Default: &pipelinev1.ParamValue{
						StringVal: *gen.Image.Name,
						Type:      pipelinev1.ParamTypeString,
					},
					Type: pipelinev1.ParamTypeString,
				},
				{
					Name:        "container-tag",
					Description: "Container tag for the task",
					Default: &pipelinev1.ParamValue{
						StringVal: *gen.Image.Tag,
						Type:      pipelinev1.ParamTypeString,
					},
					Type: pipelinev1.ParamTypeString,
				},
			},
			StepTemplate: &pipelinev1.StepTemplate{
				Image: "$(params.container-image):$(params.container-tag)",
				Env:   gen.Options.WithExtraEnv(gen.Spec.Config.ExtraEnv).BuildEnvironment(),
			},
			Steps: []pipelinev1.Step{
				{
					Name:             "task-command",
					Command:          gen.Spec.Config.Command,
					Args:             gen.Spec.Config.Args,
					ComputeResources: corev1.ResourceRequirements(*gen.Spec.Resources),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "system-config",
							ReadOnly:  true,
							MountPath: "/opt/system-extra-configs",
						},
					},
					Timeout: gen.Spec.Config.Timeout,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "system-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							DefaultMode: ptr.To[int32](420),
							SecretName:  gen.ConfigFilesSecret,
						},
					},
				},
			},
		},
	}

	if gen.TwemproxySpec != nil {
		task.Spec = twemproxy.AddTwemproxyTaskSidecar(task.Spec, gen.TwemproxySpec)
	}

	return task
}
