APP?=arch-course
DOCKERHUB?=ilyashikhaleev/arch-course
PORT?=8000
RELEASE?=0.0.8
USER_INFO_APP?=arch-course
USER_INFO_DOCKERHUB?=ilyashikhaleev/arch-course-user-info
USER_INFO_PORT?=8000
USER_INFO_RELEASE?=0.0.1

all: build build-user-info

.PHONY: clean
clean:
	rm -f ./bin/${APP} ; \
	rm -f ./bin/${USER_INFO_APP}

.PHONY: build
build: clean
	docker build -t $(DOCKERHUB):$(RELEASE) .

.PHONY: build-user-info
build-user-info: clean
	docker build -t $(USER_INFO_DOCKERHUB):$(USER_INFO_RELEASE) -f Dockerfile.user-info .

# helm
.PHONY: start
start: update-helm-dependency run-auth update-helm-dependency-user-info run-user-info update-helm-dependency run

.PHONY: run
run: run-auth run-user-info

.PHONY: run-auth
run-auth:
	helm uninstall archapp ; \
	helm install archapp ./helm/arch-chart

.PHONY: update-helm-dependency
update-helm-dependency:
	helm dependency update ./helm/arch-chart

.PHONY: run-user-info
run-user-info:
	helm uninstall archapp-user-info ; \
	helm install archapp-user-info ./helm/user-info-chart

.PHONY: update-helm-dependency-user-info
update-helm-dependency-user-info:
	helm dependency update ./helm/user-info-chart

# stresstest
.PHONY: run-stresstest
run-stresstest:
	kubectl apply -f ./helm/stresstest.yaml

.PHONY: stop-stresstest
stop-stresstest:
	kubectl delete -f ./helm/stresstest.yaml
