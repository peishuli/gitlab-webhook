package main

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/go-playground/webhooks.v5/gitlab"
)

const (
	path = "/webhook"
)

func main() {

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
	fmt.Println("Webhook listieng to port 8080...")
	http.ListenAndServe(":8080", nil)
}
