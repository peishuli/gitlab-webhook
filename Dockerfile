FROM golang:latest
# third party depenencies has to be manually added to the image here
RUN go get -u gopkg.in/go-playground/webhooks.v5/gitlab
#RUN go get -u github.com/peishuli/gitlab-webhook/sample

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]