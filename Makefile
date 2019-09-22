rebuild: #default action
	go build main.go
	docker build -t peishu/webhook:v1 .
	docker push peishu/webhook:v1
	rm main
	kubectl delete -f config/deployment.yaml
	kubectl apply -f config/deployment.yaml

refresh:
	kubectl delete -f config/deployment.yaml
	kubectl apply -f config/deployment.yaml

deploy:
	kubectl apply -f config/deployment.yaml
	kubectl apply -f config/service.yaml

build:
	go build main.go

cl: #cleanup:
	kubectl delete pr --all
	kubectl delete pipeline --all
	kubectl delete task --all
	kubectl delete pipelineresources --all
