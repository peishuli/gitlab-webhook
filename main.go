package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	tekton "github.com/peishuli/gitlab-webhook/tekton"
	tektoncd "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	"gopkg.in/go-playground/webhooks.v5/gitlab"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	path = "/webhook"
)

func main() {

	// Get the clientset
	k8s, err := getK8s()
	if err != nil {
		fmt.Errorf("Could not get k8s client: %s", err)
	}

	tekton, err := getTekton()
	if err != nil {
		fmt.Errorf("Could not get tekton client: %s", err)
	}

	client := tekton.Client{
		K8s:    k8s,
		Tekton: tekton,
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
			fmt.Println("Push event detected...")
			push := payload.(gitlab.PushEventPayload)
			fmt.Printf("%+v", push)
			fmt.Printf("CommitId=%s\n", push.CheckoutSHA)
			fmt.Printf("RepositoryUrl=%s\n", push.Repository.URL)
			parts := strings.Split(push.Ref, "/") //Ref:refs/head/dev
			fmt.Printf("Branch=%s\n", parts[2])

			options := tekton.NewTaskRunOptions()
			options.Prefix = push.Project.Name

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
	http.ListenAndServe(port, nil)
}

func getK8s() (*k8s.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func getTekton(*tekton.TektonV1alpha1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	tekton, err := tektoncd.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return tekton, nil

}
