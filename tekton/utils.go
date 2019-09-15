package tekton

import (
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
	api "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1" 
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	TektonClient *tektonv1alpha1.TektonV1alpha1Client
	K8sclient *k8s.Clientset
}

func (c Client) CreateTaskRun(projectName string) {
	taskrundef := createTaskRunDef(projectName)
	_, err := c.TektonClient.TaskRuns("default").Create(taskrundef)

	if err != nil {
		fmt.Printf("error creating taskrun: %v", err)
	}

}

func createTaskRunDef(projectName string) *api.TaskRun {

	taskrun := api.TaskRun{
		ObjectMeta: metav1.ObjectMeta {
			GenerateName: fmt.Sprintf("%s-taskrun-", projectName),
			Namespace: "default",
		},
		Spec: api.TaskRunSpec {
			ServiceAccount: "build-bot",
			TaskRef: &api.TaskRef {
				Name: fmt.Sprintf("%s-build-task", projectName),	
			},
			Inputs: api.TaskRunInputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "source",
						ResourceRef: api.PipelineResourceRef{
							Name: fmt.Sprintf("%s-git", projectName),
						},
					},
				},
			},	
			Outputs: api.TaskRunOutputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "image",
						ResourceRef: api.PipelineResourceRef{
							Name: fmt.Sprintf("%s-image", projectName),
						},
					},
				},
			},
		},
	}

	return &taskrun
}


