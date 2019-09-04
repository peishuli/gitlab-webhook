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

## Hint: you can generate a random secretToken with:
head -c 8 /dev/urandom | base64