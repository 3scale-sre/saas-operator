package system

import (
	"fmt"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (gen *SystemTektonGenerator) pipeline() *pipelinev1.Pipeline {
	pipeline := &pipelinev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gen.GetComponent(),
			Namespace: gen.GetNamespace(),
			Labels:    gen.GetLabels(),
		},
		Spec: pipelinev1.PipelineSpec{
			DisplayName: gen.GetComponent(),
			Description: *gen.Spec.Description,
			Params: []pipelinev1.ParamSpec{
				{
					Name:        "container-image",
					Description: "Container image for the task",
					Default: &pipelinev1.ParamValue{
						StringVal: fmt.Sprint(*gen.Image.Name),
						Type:      pipelinev1.ParamTypeString,
					},
					Type: pipelinev1.ParamTypeString,
				},
				{
					Name:        "container-tag",
					Description: "Container tag for the task",
					Default: &pipelinev1.ParamValue{
						StringVal: fmt.Sprint(*gen.Image.Tag),
						Type:      pipelinev1.ParamTypeString,
					},
					Type: pipelinev1.ParamTypeString,
				},
			},
			Tasks: []pipelinev1.PipelineTask{
				{
					Name: *gen.Spec.Name,
					Params: pipelinev1.Params{
						pipelinev1.Param{
							Name: "container-image",
							Value: pipelinev1.ParamValue{
								StringVal: "$(params.container-image)",
								Type:      pipelinev1.ParamTypeString,
							},
						},
						pipelinev1.Param{
							Name: "container-tag",
							Value: pipelinev1.ParamValue{
								StringVal: "$(params.container-tag)",
								Type:      pipelinev1.ParamTypeString,
							},
						},
					},
					TaskRef: &pipelinev1.TaskRef{
						Name: gen.GetComponent(),
						Kind: pipelinev1.NamespacedTaskKind,
					},
				},
			},
		},
	}
	return pipeline
}
