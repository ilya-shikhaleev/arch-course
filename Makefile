APP?=arch-course
DOCKERHUB?=ilyashikhaleev/arch-course
PORT?=8000
RELEASE?=0.0.8
AUTHAPP?=arch-course
AUTHDOCKERHUB?=ilyashikhaleev/arch-course-auth
PORT?=8000
AUTHRELEASE?=0.0.1

all: build build-auth

.PHONY: clean
clean:
	rm -f ./bin/${APP} ; \
	rm -f ./bin/${AUTHAPP}

.PHONY: build
build: clean
	docker build -t $(DOCKERHUB):$(RELEASE) .

.PHONY: build-auth
build-auth: clean
	docker build -t $(AUTHDOCKERHUB):$(AUTHRELEASE) -f Dockerfile.auth .

# helm
.PHONY: start
start: build build-auth update-helm-dependency run

.PHONY: run
run:
	helm uninstall archapp ; \
	helm install archapp ./helm/arch-chart

.PHONY: update-helm-dependency
update-helm-dependency:
	helm dependency update ./helm/arch-chart

.PHONY: run-auth
run-auth:
	helm uninstall archappauth ; \
	helm install archappauth ./helm/auth-chart

.PHONY: update-helm-dependency-auth
update-helm-dependency-auth:
	helm dependency update ./helm/auth-chart

# stresstest
.PHONY: run-stresstest
run-stresstest:
	kubectl apply -f ./helm/stresstest.yaml

.PHONY: stop-stresstest
stop-stresstest:
	kubectl delete -f ./helm/stresstest.yaml

# k8s commands
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