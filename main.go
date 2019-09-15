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
	}
	k8sClient, err := k8s.NewForConfig(config)
	if err != nil {
		fmt.Printf("Could not create k8sClient: %s\n", err.Error())
	}
	
	tektonClient, err := tektonv1alpha1.NewForConfig(config)
	if err != nil {
		fmt.Printf("Could not create tektonClient: %s\n", err.Error())
	}


	client := tektonutil.Client{
		K8sclient: k8sClient,
		TektonClient: tektonClient,
	}	

	
	hook, _ := gitlab.New(gitlab.Options.Secret("MyGitLabSuperSecretSecrect"))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

		payload, err := hook.Parse(r, gitlab.PushEvents, gitlab.MergeRequestEvents)
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
				fmt.Println("Got an error here...")
			}
		}

		switch payload.(type) {

		case gitlab.PushEventPayload:
			//TODO: parameterize create task runs
			fmt.Println("Push event detected...")
			push := payload.(gitlab.PushEventPayload)
			fmt.Printf("%+v", push)
			fmt.Printf("CommitId=%s\n", push.CheckoutSHA)
			fmt.Printf("RepositoryUrl=%s\n", push.Repository.URL)
			parts := strings.Split(push.Ref, "/") //Ref:refs/head/dev
			fmt.Printf("Branch=%s\n", parts[2])
			projectName := strings.ToLower(push.Project.Name)
			client.CreateTaskRun(projectName)
			
		case gitlab.MergeRequestEventPayload:
			fmt.Println("Merge request event detected...")
			mergeRequest := payload.(gitlab.MergeRequestEventPayload)
			// Do whatever you want from here...
			//fmt.Printf("%+v", mergeRequest)
			fmt.Printf("CommitId=%s\n", mergeRequest.ObjectAttributes.SHA)
			fmt.Printf("RepositoryUrl=%s\n", mergeRequest.Repository.URL)
			fmt.Printf("Branch=%s\n", mergeRequest.ObjectAttributes.TargetBranch)

		case gitlab.TagEventPayload:
			fmt.Println("Tag event detected...")

		default:
			fmt.Println("Unknown event detected...")
		}
	})

	port := os.Getenv("PORT")
	if port != "" {
		port = ":" + port
	}
	fmt.Printf("Webhook listieng to port %s...\n", port)
	//http.ListenAndServe(port, nil)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

