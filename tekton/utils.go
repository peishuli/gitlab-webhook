package tekton

import (
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
	api "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1" 
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
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
	// Create the build task if not exists (an idempontent opration)
	c.createBuildTask(buildInfo)
	
	// Now create taskrun
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

func (c Client) createBuildTask(buildInfo BuildInfo) {
	taskName := fmt.Sprintf("%s-build-task", buildInfo.ProjectName)
	_, err := c.TektonClient.Tasks("default").Get(taskName, metav1.GetOptions{})

	if err == nil  {
		// named task already exists
		return
	} 
	
	taskDef := createBuildTaskDef(buildInfo)

	_, err = c.TektonClient.Tasks("default").Create(taskDef)

	if err != nil {
		fmt.Printf("error creating task: %v", err)
	}

	
}

func createBuildTaskDef(buildInfo BuildInfo) *api.Task {

	task := api.Task {
		ObjectMeta: metav1.ObjectMeta {
			Name: fmt.Sprintf("%s-build-task", buildInfo.ProjectName),
			Namespace: "default",
		},
		Spec: api.TaskSpec {
			Inputs: &api.Inputs {
				Resources: []api.TaskResource {
					api.TaskResource {
						Name: "source",
						Type: api.PipelineResourceTypeGit,
					},
				},
				Params: []api.ParamSpec {
					api.ParamSpec {
						Name: "DOCKERFILE",
						Description: "The name of the Dockerfile",
						Default: &api.ArrayOrString {
							Type: api.ParamTypeString,
							StringVal: "Dockerfile",
						},
					},
					api.ParamSpec {
						Name: "BUILDKIT_CLIENT_IMAGE",
						Description: "The name of the BuildKit client (buildctl) image",
						Default: &api.ArrayOrString {
							Type: api.ParamTypeString,
							StringVal: "moby/buildkit:v0.5.1@sha256:d45d15f3b22fcfc1b112ffafc40fd2f2d530245e63cfe346a65bd75acdc4d346",
						},
					},
					api.ParamSpec {
						Name: "BUILDKIT_DAEMON_ADDRESS",
						Description: "The address of the BuildKit daemon (buildkitd) service",
						Default: &api.ArrayOrString {
							Type: api.ParamTypeString,
							StringVal: "tcp://buildkitd:1234",
						},
					},
					api.ParamSpec {
						Name: "COMMITID",
						Description: "Gitlab repo commit Id",
					},
				},
			},
			Outputs: &api.Outputs {
				Resources: []api.TaskResource {
					api.TaskResource {
						Name: "image",
						Type: api.PipelineResourceTypeImage,
					},
				},
			},
			Steps: []api.Step {
				api.Step {
					//Name: "something",
					corev1.Container {
						Name: "build-and-push",
						Image: "${inputs.params.BUILDKIT_CLIENT_IMAGE}",
						WorkingDir: "/workspace/source",
						Command: []string {
							"buildctl", "--debug", "--addr=${inputs.params.BUILDKIT_DAEMON_ADDRESS}", "build",
							"--progress=plain",
							"--frontend=dockerfile.v0",
							"--opt", "filename=${inputs.params.DOCKERFILE}",
							"--local", "context=.", "--local", "dockerfile=.",
							"--output", "type=image,name=${outputs.resources.image.url}:dev-$(inputs.params.COMMITID),push=true",
							"--export-cache", "type=inline",
							"--import-cache", "type=registry,ref=${outputs.resources.image.url}",
						},
					},
				},
			},
		},
	}

	return &task
}

