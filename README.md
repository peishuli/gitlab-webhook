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

## Hint: you can generate a random secretToken with:
head -c 8 /dev/urandom | base64

