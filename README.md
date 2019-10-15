# Instructions

## Depenencies 
"gopkg.in/go-playground/webhooks.v5/gitlab"
Refs: 
- https://github.com/go-playground/webhooks
- https://godoc.org/gopkg.in/go-playground/webhooks.v5/gitlab
- https://docs.gitlab.com/ee/user/project/integrations/webhooks.html

## Dummy GitLab Repo:
https://gitlab.com/peishu/webhoo-demo
local repo folder:
/mnt/c/temp/webhooktest/webhoo-demo

## Running into issues
- https://github.com/kubernetes/client-go/issues/656
- Antoher issue (was able to fix it with dep locally but have to figure out how to fix it in Docker build):
```
Step 5/11 : RUN go get -u k8s.io/client-go/kubernetes
 ---> Running in 481ae1252cf6
# k8s.io/client-go/transport
src/k8s.io/client-go/transport/round_trippers.go:70:11: cannot convert klog.V(9) (type klog.Verbose) to type bool
src/k8s.io/client-go/transport/round_trippers.go:72:11: cannot convert klog.V(8) (type klog.Verbose) to type bool
src/k8s.io/client-go/transport/round_trippers.go:74:11: cannot convert klog.V(7) (type klog.Verbose) to type bool
src/k8s.io/client-go/transport/round_trippers.go:76:11: cannot convert klog.V(6) (type klog.Verbose) to type bool
The command '/bin/sh -c go get -u k8s.io/client-go/kubernetes' returned a non-zero code: 2
```
Possible soluton for the docker dep issue above:
https://stackoverflow.com/questions/52578581/trying-to-install-dependencies-using-dep-in-docker
## Hint: you can generate a random secretToken with:
head -c 8 /dev/urandom | base64

## Demo setup

- Deploy webhook
```
make deploy
...
deploy secret and sa (from the config repo)
```
- Setup ArgoCD
- Setup Tekton (including dashboard and BuildKit daemon)
```
kubectl apply -f https://raw.githubusercontent.com/tektoncd/catalog/master/buildkit/0-buildkitd.yaml
```
- Create ns demo
- Setup Tekton and its dashboard (notice the versions)
```
k apply -f https://storage.googleapis.com/tekton-releases/previous/v0.6.0/release.yaml
k apply -f https://github.com/tektoncd/dashboard/releases/download/v0.1.1/release.yaml
```
Note: In ASK, if rbac is enabled, add clusterrolebinding to the default user (yaml in the config repo)

- Optional: patch argoCD and tkn dashboards
```
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'
kubectl patch svc tekton-dashboard -n tekton-pipelines -p '{"spec": {"type": "LoadBalancer"}}'
```
## URLs
- [ArgoCD](https://104.214.109.6/)
- [Tekton](http://104.214.104.154:9097)
- [Webhook](http://65.52.39.93:8080/webhook)

## Notes:
ArgoCD's (resync)  interval is hardcoded as 3 minutes (defaultAppResyncPeriod = 180) in https://github.com/argoproj/argo-cd/blob/master/cmd/argocd-application-controller/main.go. This behavior can be overwritten via argocd webhook (i.e., push vs pull)

## Install Tekton CD Pipelines on OpenShift
```
oc new-project tekton-pipelines
oc adm policy add-scc-to-user anyuid -z tekton-pipelines-controller
oc apply --filename https://storage.googleapis.com/tekton-releases/latest/release.yaml
```

## Install Teckon Dashboard on OpenShift
```
oc apply --filename https://github.com/tektoncd/dashboard/releases/download/v0.2.0/openshift-tekton-dashboard.yaml --validate=false
```

