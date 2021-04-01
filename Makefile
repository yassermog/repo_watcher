NAME=repo_watcher
TAG=yassermog/$(NAME)
VER=latest

all: clean build push deploy

build:
	docker build -t $(TAG) -t $(TAG):$(VER) .

run:
	docker run -d -p 6060:6060 -e PORT=6060 --name=$(NAME) $(TAG)

clean:
	-docker stop $(NAME)
	-docker rm $(NAME)
	-kubectl delete deployment repo-watcher
	-kubectl delete configmap app-config

push:
	-docker push $(TAG):$(VER)
	
deploy:
	- kubectl apply -f watcher_deployment.yaml
	- kubectl apply -f watcher_service.yaml
	- kubectl apply -f configmap.yaml

proxy:
	minikube service --url repo-watcher-service

install: 
	helm install repo-watcher ./repo-watcher/ --set service.type=NodePort

auth:
	- kubectl apply -f service-admin-role.yaml
	- kubectl create clusterrolebinding service-admin-pod --clusterrole=cluster-admin --serviceaccount=default:repo-watcher
