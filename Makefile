USER_APP?=user
USER_DOCKERHUB?=ilyashikhaleev/arch-course-user
USER_PORT?=8000
USER_RELEASE?=0.0.1
USER_HELM_RELEASE_NAME?=user

PRODUCT_APP?=product
PRODUCT_DOCKERHUB?=ilyashikhaleev/arch-course-product
PRODUCT_PORT?=8000
PRODUCT_RELEASE?=0.0.1
PRODUCT_HELM_RELEASE_NAME?=product

all: build

.PHONY: clean
clean:
	rm -f ./bin/${USER_APP} ; \
	rm -f ./bin/${PRODUCT_APP}

.PHONY: build
build: clean build-user build-product

.PHONY: build-user
build-user: clean
	docker build -t $(USER_DOCKERHUB):$(USER_RELEASE) -f Dockerfile.user .

.PHONY: build-product
build-product: clean
	docker build -t $(PRODUCT_DOCKERHUB):$(PRODUCT_RELEASE) -f Dockerfile.product .

# helm
.PHONY: start
start: update-helm-dependency-user run-user update-helm-dependency-product run-product

.PHONY: run
run: run-user run-product

.PHONY: run-user
run-user:
	helm uninstall $(USER_HELM_RELEASE_NAME) ; \
	helm install $(USER_HELM_RELEASE_NAME) ./helm/user-chart

.PHONY: update-helm-dependency-user
update-helm-dependency-user:
	helm dependency update ./helm/user-chart

.PHONY: run-product
run-product:
	helm uninstall $(PRODUCT_HELM_RELEASE_NAME) ; \
	helm install $(PRODUCT_HELM_RELEASE_NAME) ./helm/product-chart

.PHONY: update-helm-dependency-product
update-helm-dependency-product:
	helm dependency update ./helm/product-chart

# stresstest
.PHONY: run-stresstest
run-stresstest:
	kubectl apply -f ./helm/stresstest.yaml

.PHONY: stop-stresstest
stop-stresstest:
	kubectl delete -f ./helm/stresstest.yaml
