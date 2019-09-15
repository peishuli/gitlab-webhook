package tekton

import (
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
	api "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1" 
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BuildInfo struct {
	ProjectName string
	CommitId string
}

type Client struct {
	TektonClient *tektonv1alpha1.TektonV1alpha1Client
	K8sclient *k8s.Clientset
}

func (c Client) CreateTaskRun(buildInfo BuildInfo) {
	taskrundef := createTaskRunDef(buildInfo)
	_, err := c.TektonClient.TaskRuns("default").Create(taskrundef)

	if err != nil {
		fmt.Printf("error creating taskrun: %v", err)
	}

}

func createTaskRunDef(buildInfo BuildInfo) *api.TaskRun {

	taskrun := api.TaskRun{
		ObjectMeta: metav1.ObjectMeta {
			GenerateName: fmt.Sprintf("%s-taskrun-", buildInfo.ProjectName),
			Namespace: "default",
		},
		Spec: api.TaskRunSpec {
			ServiceAccount: "build-bot",
			TaskRef: &api.TaskRef {
				Name: fmt.Sprintf("%s-build-task", buildInfo.ProjectName),	
			},
			Inputs: api.TaskRunInputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "source",
						ResourceRef: api.PipelineResourceRef{
							Name: fmt.Sprintf("%s-git", buildInfo.ProjectName),
						},
					},
				},
				Params: []api.Param {
					api.Param {
						Name: "COMMITID",
						Value: api.ArrayOrString{
							Type: api.ParamTypeString,
							StringVal: buildInfo.CommitId,
						},
					},
				},
			},	
			Outputs: api.TaskRunOutputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "image",
						ResourceRef: api.PipelineResourceRef{
							Name: fmt.Sprintf("%s-image", buildInfo.ProjectName),
						},
					},
				},
			},
		},
	}

	return &taskrun
}


