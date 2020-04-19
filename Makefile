APP?=arch-course
DOCKERHUB?=ilyashikhaleev/arch-course
PORT?=8000
RELEASE?=0.0.6

all: build

.PHONY: clean
clean:
	rm -f ./bin/${APP}

.PHONY: build
build: clean
	docker build -t $(DOCKERHUB):$(RELEASE) .

.PHONY: run
run:
	helm dependency build ./arch-chart ; \
	helm uninstall archapp ; \
	helm install archapp ./arch-chart

.PHONY: k8s-clear
k8s-clear:
	kubectl delete -f ./k8s/

.PHONY: k8s
k8s:
	kubectl apply -f ./k8s/secrets.yaml && \
	kubectl apply -f ./k8s/config.yaml && \
	kubectl apply -f ./k8s/postgres.yaml && \
	kubectl apply -f ./k8s/deployment.yaml && \
	kubectl apply -f ./k8s/service.yaml && \
	kubectl apply -f ./k8s/initdb.yaml && \
	kubectl apply -f ./k8s/ingress.yaml

