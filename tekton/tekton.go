package tekton

import (
	"github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/pipeline"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tekton "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
)

type Client struct {
	Tekton *tekton.TektonV1alpha1Client
	K8s *k8s.Clientset
}

type TaskRunOptions struct {
	Prefix string
	Namespace string
	ServiceAccount string
}

// set some defaults
func NewTaskRunOptions() TaskRunOptions {
	options := TaskRunOptions {}
	options.Namespace = "default"
	options.ServiceAccount = "build-bot"

	return options
}

func (c *Client) CreateTaskRun(options TaskRunOptions) {
	taskRunDef := createTaskRunDef(options)
	taskRun, err := c.Tekton.TaskRuns(options.Namespace).Create(taskRunDef)	
	if err != nil {
		fmt.Errorf("Error creating taskRun: %s\n", err)
	} else {
		fmt.Printf("Taskrun created: %v\n", taskRun)
	}
}

func createTaskRunDef(options TaskRunOptions) *v1alpha1.TaskRun {
	taskRun := v1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-git-source-", options.Prefix),
			Namespace: options.Namespace,

		},
		Spec: v1alpha1.TaskRunSpec{
			ServiceAccount: options.ServiceAccount,
			TaskRef: v1alpha1.TaskRef {
				Name: "identity-build-task" //TODO: remove hard-coding here and maybe below too
			},
			Inputs: v1alpha1.Inputs {
				Resources: []v1alpha1.TaskResource {
					v1alpha1.TaskResource{
						Name: "docker-source",
						TaskRef: v1alpha1.PipelineResourceRef{
							Name: "identity-git",
						},
					},
				},
				Params: []v1alpha1.Param{
					v1alpha1.Param{
						Name:  "pathToDockerFile",
						Value: "Dockerfile",
					},
					v1alpha1.Param{
						Name:  "pathToContext",
						Value: "/workspace/docker-source/",
					},
				},
			},	
			Outputs: v1alpha1.Outputs {
				Resources: []v1alpha1.TaskResource {
					v1alpha1.TaskResource {
						Name: "builtImage",
						TaskRef: v1alpha1.PipelineResourceRef{
							Name: "identity-image",
						},
					},
				},
			},	
		},
	}

	return &taskRun

}
