package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"flag"
	"gopkg.in/go-playground/webhooks.v5/gitlab"
	tektonutil "github.com/peishuli/gitlab-webhook/tekton"
	"k8s.io/client-go/rest"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	path = "/webhook"
)


func main() {

	flag.Parse()

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Could not retrieve config: %s\n", err.Error())
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
		fmt.Printf("%s\n", client)

	//client := utils.New()
	
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

