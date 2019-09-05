package tekton

import (
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
	api "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1" 
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"
)

type Client struct {
	TektonClient *tektonv1alpha1.TektonV1alpha1Client
	K8sclient *k8s.Clientset
}

func (c Client) CreateTaskRun() {
	taskrundef := createTaskRunDef()
	_, err := c.TektonClient.TaskRuns("default").Create(taskrundef)

	if err != nil {
		fmt.Printf("error creating taskrun: %v", err)
	}

}

func createTaskRunDef() *api.TaskRun {

	taskrun := api.TaskRun{
		ObjectMeta: metav1.ObjectMeta {
			Name: "identity-taskrun-" + strconv.FormatInt(time.Now().Unix(), 10) ,
			Namespace: "default",
		},
		Spec: api.TaskRunSpec {
			ServiceAccount: "build-bot",
			TaskRef: &api.TaskRef {
				Name: "identity-build-task",		
			},
			Inputs: api.TaskRunInputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "docker-source",
						ResourceRef: api.PipelineResourceRef{
							Name: "identity-git",
						},
					},
				},
				Params: []api.Param {
					api.Param {
						Name: "pathToDockerFile",
						Value: api.ArrayOrString{
							Type: api.ParamTypeString,
							StringVal: "Dockerfile",
						},
					},
					api.Param {
						Name: "pathToContext",
						 Value: api.ArrayOrString{
							 Type: api.ParamTypeString,
							 StringVal: "/workspace/docker-source/",
						},
					},
				},
			},	
			Outputs: api.TaskRunOutputs {
				Resources: []api.TaskResourceBinding {
					api.TaskResourceBinding{ 
						Name: "builtImage",
						ResourceRef: api.PipelineResourceRef{
							Name: "identity-image",
						},
					},
				},
			},
		},
	}

	return &taskrun
}


