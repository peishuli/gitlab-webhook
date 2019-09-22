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
	Namespace string
	ValuesFile string
	GitlabEmail string
	GitlabUsername string
	GitlabPassword string
	GitlabGroup string
	GitlabConfigRepository string
	Revision string
}

type Client struct {
	TektonClient *tektonv1alpha1.TektonV1alpha1Client
	K8sclient *k8s.Clientset
}


func (c Client) CreatePipelineRun(buildInfo BuildInfo) {
	// Create git pipeline resource if not exists
	c.createGitResource(buildInfo)

	// Create image pipeline resource if not exists
	c.createImageResource(buildInfo)

	// Create the build task if not exists
	c.createBuildTask(buildInfo)

	// Create the push task if not exists
	c.createPushTask(buildInfo)

	// Create the pipeline if not exists
	c.createPipeline(buildInfo)

	// Now create pipelinerun
	pipelinerunDef := createPipelineRunDef(buildInfo)

	_, err := c.TektonClient.PipelineRuns(buildInfo.Namespace).Create(pipelinerunDef)

	if err != nil {
		fmt.Printf("error creating pipelinerun: %v\n", err)
	}
}

func createPipelineRunDef(buildInfo BuildInfo) *api.PipelineRun {
	pipelineRun := api.PipelineRun {
		ObjectMeta: metav1.ObjectMeta {
			GenerateName: fmt.Sprintf("%s-pipelinerun-", buildInfo.ProjectName),
			Namespace: buildInfo.Namespace,
		},
		Spec: api.PipelineRunSpec {
			ServiceAccount: "build-bot",
			PipelineRef: api.PipelineRef {
				Name: fmt.Sprintf("%s-pipeline", buildInfo.ProjectName),
			},
			Resources: []api.PipelineResourceBinding {
				api.PipelineResourceBinding {
					Name: "source",
					ResourceRef: api.PipelineResourceRef {
						Name: fmt.Sprintf("%s-git", buildInfo.ProjectName),
					},
				
				},
				api.PipelineResourceBinding {
					Name: "image",
					ResourceRef: api.PipelineResourceRef {
						Name: fmt.Sprintf("%s-image", buildInfo.ProjectName),
					},
				},
			},
			Params: []api.Param {
				api.Param {
					Name: "COMMITID",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.CommitId,
					},
				},
				api.Param {
					Name: "VALUES_FILE",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.ValuesFile, //"values.yaml"
					},
				},
				api.Param {
					Name: "GIT_EMAIL",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.GitlabEmail, //"peishuli62@gitlab.com"
					},
				},
				api.Param {
					Name: "GIT_USERNAME",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.GitlabUsername, //"peishu"
					},
				},
				api.Param {
					Name: "GIT_PASSWORD",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.GitlabPassword, //"Pass%40word1"
					},
				},
				api.Param {
					Name: "GIT_GROUP",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.GitlabGroup, //"peishu"
					},
				},
				api.Param {
					Name: "GIT_REPO",
					Value: api.ArrayOrString {
						Type: api.ParamTypeString,
						StringVal: buildInfo.GitlabConfigRepository, //"identity-config" 
					},
				},
			},
		},
	}

	return &pipelineRun
}

func (c Client) createBuildTask(buildInfo BuildInfo) {
	taskName := fmt.Sprintf("%s-build-task", buildInfo.ProjectName)
	_, err := c.TektonClient.Tasks(buildInfo.Namespace).Get(taskName, metav1.GetOptions{})

	if err == nil  {
		// named task already exists
		return
	} 
	
	taskDef := createBuildTaskDef(buildInfo)

	_, err = c.TektonClient.Tasks(buildInfo.Namespace).Create(taskDef)

	if err != nil {
		fmt.Printf("error creating task: %v", err)
	}
}

func createBuildTaskDef(buildInfo BuildInfo) *api.Task {

	task := api.Task {
		ObjectMeta: metav1.ObjectMeta {
			Name: fmt.Sprintf("%s-build-task", buildInfo.ProjectName),
			Namespace: buildInfo.Namespace,
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
							"--output", "type=image,name=${outputs.resources.image.url}:${inputs.params.COMMITID},push=true",
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

func (c Client) createPushTask(buildInfo BuildInfo) {
	taskName := fmt.Sprintf("%s-push-to-git-task", buildInfo.ProjectName)
	_, err := c.TektonClient.Tasks(buildInfo.Namespace).Get(taskName, metav1.GetOptions{})
	
	if err == nil  {
		// named task already exists
		return
	} 
	
	taskDef := createPushTaskDef(buildInfo)
	_, err = c.TektonClient.Tasks(buildInfo.Namespace).Create(taskDef)
	
	if err != nil {
		fmt.Printf("Task creation error: %s.\n", err.Error())
	} 
}

func createPushTaskDef(buildInfo BuildInfo) *api.Task {
	// define the commands
	argsString := "git config --global user.email \"${inputs.params.GIT_EMAIL}@gitlab.com\"\n"
	argsString += "git config --global user.name \"${inputs.params.GIT_USERNAME}\"\n"
	argsString += "git remote set-url origin git@gitlab.com/${inputs.params.GIT_GROUP}/${inputs.params.GIT_REPO}.git\n"
	argsString += "git clone https://${inputs.params.GIT_USERNAME}:${inputs.params.GIT_PASSWORD}@gitlab.com/${inputs.params.GIT_GROUP}/${inputs.params.GIT_REPO}.git\n"
	argsString += "cd ${inputs.params.GIT_REPO}\n"
	argsString += "cat ${inputs.params.VALUES_FILE} | yq w - image.tag ${inputs.params.COMMITID}> values2.yaml && mv values2.yaml ${inputs.params.VALUES_FILE}\n"
	argsString += "git add .\n"
	argsString += "git commit -m \"Image tag updated by the webhook.\"\n"
	argsString += "git push"	
	
	task := api.Task {
		ObjectMeta: metav1.ObjectMeta {
			Name: fmt.Sprintf("%s-push-to-git-task", buildInfo.ProjectName),
			Namespace: buildInfo.Namespace,
		},
		Spec: api.TaskSpec {
			Inputs: &api.Inputs {
				Params: []api.ParamSpec {
					api.ParamSpec {
						Name: "COMMITID",
						Description: "Gitlab repo commit Id",						
					},
					api.ParamSpec {
						Name: "VALUES_FILE",
						Description: "The name of the values file of the Helm chart",
						Default: &api.ArrayOrString {
							Type: api.ParamTypeString,
							StringVal: "values.yaml",
						},
					},
					api.ParamSpec {
						Name: "GIT_EMAIL",
						Description: "The email address of the gitlab account",						
					},
					api.ParamSpec {
						Name: "GIT_USERNAME",
						Description: "The Gitlab username",						
					},
					api.ParamSpec {
						Name: "GIT_PASSWORD",
						Description: "The Gitlab password (urlencode)",						
					},
					api.ParamSpec {
						Name: "GIT_GROUP",
						Description: "The Gitlab group name",						
					},
					api.ParamSpec {
						Name: "GIT_REPO",
						Description: "The Gitlab config repository name",						
					},
										
				},
			},
			Steps: []api.Step {
				api.Step {
					corev1.Container {
						Name: "update-config-repo",
						Image: "peishu/yq-git",
						Command: []string {"/bin/sh", "-c"},
						Args: []string { 							
							argsString,
						},						
					},
				},
			},
		},
	}	    

	return &task
}


func (c Client) createPipeline(buildInfo BuildInfo) {
	pipelineName := fmt.Sprintf("%s-pipeline", buildInfo.ProjectName)
	_, err := c.TektonClient.Pipelines(buildInfo.Namespace).Get(pipelineName, metav1.GetOptions{})

	if err == nil  {
		// named pipeline already exists
		return
	} 
	
	pipelineDef := c.createPipelineDef(buildInfo)

	_, err = c.TektonClient.Pipelines(buildInfo.Namespace).Create(pipelineDef)

	if err != nil {
		fmt.Printf("error creating pipeline: %v", err)
	}
}

func (c Client) createPipelineDef(buildInfo BuildInfo) *api.Pipeline {
	pipeline := api.Pipeline {
		ObjectMeta: metav1.ObjectMeta {
			Name: fmt.Sprintf("%s-pipeline", buildInfo.ProjectName),
			Namespace: buildInfo.Namespace,
		},
		Spec: api.PipelineSpec {
			Resources: []api.PipelineDeclaredResource {
				api.PipelineDeclaredResource {
					Name: "source",
					Type: api.PipelineResourceTypeGit,
				},
				api.PipelineDeclaredResource {
					Name: "image",
					Type: api.PipelineResourceTypeImage,
				},
			},
			Params: []api.ParamSpec {
				api.ParamSpec {
					Name: "COMMITID",
					Description: "Gitlab repo commit Id",	
				},

				api.ParamSpec {
					Name: "VALUES_FILE",
					Description: "The name of the Helm chart values file",	
				},
				api.ParamSpec {
					Name: "GIT_EMAIL",
					Description: "The Gitlab email address",	
				},
				api.ParamSpec {
					Name: "GIT_USERNAME",
					Description: "The Gitlab username",	
				},
				api.ParamSpec {
					Name: "GIT_PASSWORD",
					Description: "The Gitlab password (urlencoded)",	
				},
				api.ParamSpec {
					Name: "GIT_GROUP",
					Description: "The Gitlab group name",	
				},
				api.ParamSpec {
					Name: "GIT_REPO",
					Description: "The Gitlab config repository name",	
				},
			},
			Tasks: []api.PipelineTask {
				api.PipelineTask {
					Name: "build-and-push",
					TaskRef: api.TaskRef {
						Name: fmt.Sprintf("%s-build-task", buildInfo.ProjectName),
					},
					Params: []api.Param {
						api.Param {
							Name: "COMMITID",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.COMMITID}",
							},
						},
					},
					Resources: &api.PipelineTaskResources {
						Inputs: []api.PipelineTaskInputResource {
							api.PipelineTaskInputResource {
								Name: "source",
								Resource: "source",
							},
						},
						Outputs: []api.PipelineTaskOutputResource {
							api.PipelineTaskOutputResource {
								Name: "image",
								Resource: "image",
							},
						},
					},
				},
				api.PipelineTask {
					Name: "update-config-repo",
					RunAfter: []string {"build-and-push"},
					TaskRef: api.TaskRef {
						Name: fmt.Sprintf("%s-push-to-git-task", buildInfo.ProjectName),
					},
					Params: []api.Param {
						api.Param {
							Name: "COMMITID",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.COMMITID}",
							},
						},
						api.Param {
							Name: "COMMITID",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.COMMITID}",
							},
						},
						api.Param {
							Name: "VALUES_FILE",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.VALUES_FILE}",
							},
						},
						api.Param {
							Name: "GIT_EMAIL",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.GIT_EMAIL}",
							},
						},
						api.Param {
							Name: "GIT_USERNAME",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.GIT_USERNAME}",
							},
						},
						api.Param {
							Name: "GIT_PASSWORD",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.GIT_PASSWORD}",
							},
						},
						api.Param {
							Name: "GIT_GROUP",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.GIT_GROUP}",
							},
						},
						api.Param {
							Name: "GIT_REPO",
							Value: api.ArrayOrString{
								Type: api.ParamTypeString,
								StringVal: "${params.GIT_REPO}",
							},
						},
					},					
				},
			},
		},
	}
	

	return &pipeline
}

func (c Client) createGitResource(buildInfo BuildInfo) {
	name := fmt.Sprintf("%s-git", buildInfo.ProjectName)
	_, err := c.TektonClient.PipelineResources(buildInfo.Namespace).Get(name, metav1.GetOptions{})

	if err == nil  {
		// named resource already exists
		return
	} 

	url := fmt.Sprintf("https://gitlab.com/%s/%s", buildInfo.GitlabGroup, buildInfo.ProjectName)
	params := []api.ResourceParam{{Name: "revision", Value: buildInfo.Revision},{Name: "url", Value: url}}

	resourceDef := createPipelineResourceDef(name, params, api.PipelineResourceTypeGit, buildInfo)

	_, err = c.TektonClient.PipelineResources(buildInfo.Namespace).Create(resourceDef)

	if err != nil {
		fmt.Printf("error creating taskrun: %v", err)
	}
	
}

func (c Client) createImageResource(buildInfo BuildInfo){
	name := fmt.Sprintf("%s-image", buildInfo.ProjectName)
	_, err := c.TektonClient.PipelineResources(buildInfo.Namespace).Get(name, metav1.GetOptions{})

	if err == nil  {
		// named resourcealready exists
		return
	} 

	url := fmt.Sprintf("registry.gitlab.com/%s/%s", buildInfo.GitlabGroup, buildInfo.ProjectName)
	params := []api.ResourceParam{{Name: "url", Value: url}}

	resourceDef := createPipelineResourceDef(name, params, api.PipelineResourceTypeImage, buildInfo)

	_, err = c.TektonClient.PipelineResources(buildInfo.Namespace).Create(resourceDef)

	if err != nil {
		fmt.Printf("error creating taskrun: %v", err)
	}
}

func createPipelineResourceDef(name string, params []api.ResourceParam, resourceType api.PipelineResourceType, buildInfo BuildInfo) *api.PipelineResource {
	reource := api.PipelineResource{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: buildInfo.Namespace},
		Spec: api.PipelineResourceSpec{
			Type:   resourceType,
			Params: params,
		},
	}
	return &reource
}