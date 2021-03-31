NAME=repo_watcher
TAG=yassermog/$(NAME)
VER=latest

all: clean build push deploy

build:
	docker build -t $(TAG) -t $(TAG):$(VER) .

run:
	docker run -d -p 7070:7070 -e PORT=7070 --name=$(NAME) $(TAG)

clean:
	-docker stop $(NAME)
	-docker rm $(NAME)
	-kubectl delete deployment repo-watcher
	-kubectl delete configmap app-config

push:
	-docker push $(TAG):$(VER)
	
deploy:
	- kubectl apply -f watcher_deployment.yaml
	- kubectl apply -f configmap.yaml


proxy:
	minikube service --url yasser-chaos

deploytest:
	- kubectl apply -f nginx_deployment.yaml