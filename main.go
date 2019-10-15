package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"flag"
	"log"
	"gopkg.in/go-playground/webhooks.v5/gitlab"
	tektonutil "github.com/peishuli/gitlab-webhook/tekton"
	"k8s.io/client-go/rest"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
	
	"k8s.io/client-go/tools/clientcmd"
)

const (
	path = "/webhook"
)


func main() {

	flag.Parse()

	namespace := os.Getenv("namespace")
	valuesFile := os.Getenv("valuesFile")
	gitlabEmail := os.Getenv("gitlabEmail")
	gitlabUsername := os.Getenv("gitlabUsername")
	gitlabPassword := os.Getenv("gitlabPassword")
	gitlabSecretToken := os.Getenv("gitlabSecretToken")
	gitlabGroup := os.Getenv("gitlabGroup")
	env := os.Getenv("env")

	var config *rest.Config
	var err error
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		fmt.Printf("Error building kubeconfig from %s: %s\n", kubeconfig, err.Error())
		os.Exit(1)
	}
	k8sClient, err := k8s.NewForConfig(config)
	if err != nil {
		fmt.Printf("Could not create k8sClient: %s\n", err.Error())
		os.Exit(1)
	}
	
	tektonClient, err := tektonv1alpha1.NewForConfig(config)
	if err != nil {
		fmt.Printf("Could not create tektonClient: %s\n", err.Error())
		os.Exit(1)
	}


	client := tektonutil.Client{
		K8sclient: k8sClient,
		TektonClient: tektonClient,
	}	

	
	hook, _ := gitlab.New(gitlab.Options.Secret(gitlabSecretToken))

	hookPath := path

	if env != "" {
		hookPath = fmt.Sprintf("%s-%s", path, env)
	}

	http.HandleFunc(hookPath, func(w http.ResponseWriter, r *http.Request) {

		payload, err := hook.Parse(r, gitlab.PushEvents, gitlab.MergeRequestEvents)
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				fmt.Printf("Event not found: %s\n", err.Error())
				os.Exit(1)
			}
		}

		switch payload.(type) {

		case gitlab.PushEventPayload:
			fmt.Println("Push event detected...")
			push := payload.(gitlab.PushEventPayload)
			
			buildInfo := tektonutil.BuildInfo {
				ProjectName: strings.ToLower(push.Project.Name),
				CommitId: push.CheckoutSHA,
				Namespace: namespace,
				ValuesFile: valuesFile,
				GitlabEmail: gitlabEmail,
				GitlabUsername: gitlabUsername, 
				GitlabPassword: gitlabPassword,
				GitlabGroup: gitlabGroup,
				GitlabConfigRepository: fmt.Sprintf("%s-config", push.Project.Name),
				Revision: push.Project.DefaultBranch, 

			}

			client.CreatePipelineRun(buildInfo)			
			
		case gitlab.MergeRequestEventPayload:
			fmt.Println("Merge request event detected...")
			mergeRequest := payload.(gitlab.MergeRequestEventPayload)	

			buildInfo := tektonutil.BuildInfo {
				ProjectName: strings.ToLower(mergeRequest.Project.Name),
				CommitId: mergeRequest.ObjectAttributes.SHA,
				Namespace: namespace,
				ValuesFile: valuesFile,
				GitlabEmail: gitlabEmail,
				GitlabUsername: gitlabUsername, 
				GitlabPassword: gitlabPassword,
				GitlabGroup: gitlabGroup,
				GitlabConfigRepository: fmt.Sprintf("%s-config", mergeRequest.Project.Name ),
				Revision: mergeRequest.Project.DefaultBranch,
			}

			client.CreatePipelineRun(buildInfo)

		default:
			fmt.Printf("Unknown event detected...\n%s\n", payload)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("Webhook is listieng to port %s...\n", port)
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

