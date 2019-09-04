FROM golang:1.12.9
#third party depenencies has to be manually added to the image here
# RUN go get -u gopkg.in/go-playground/webhooks.v5/gitlab
RUN go get -u github.com/peishuli/gitlab-webhook/tekton
# RUN go get -u github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1
# RUN go get -u k8s.io/client-go/kubernetes
# RUN go get -u k8s.io/client-go/rest
RUN go get -u github.com/imdario/mergo
RUN go get -u github.com/spf13/pflag
RUN mkdir -p /usr/local/go/src
COPY ./vendor/ /usr/local/go/src
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]