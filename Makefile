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

POPULAR_APP?=popular
POPULAR_DOCKERHUB?=ilyashikhaleev/arch-course-popular
POPULAR_PORT?=8000
POPULAR_RELEASE?=0.0.1
POPULAR_HELM_RELEASE_NAME?=popular

CART_APP?=cart
CART_DOCKERHUB?=ilyashikhaleev/arch-course-cart
CART_PORT?=8000
CART_RELEASE?=0.0.1
CART_HELM_RELEASE_NAME?=cart

ORDER_APP?=order
ORDER_DOCKERHUB?=ilyashikhaleev/arch-course-order
ORDER_PORT?=8000
ORDER_RELEASE?=0.0.1
ORDER_HELM_RELEASE_NAME?=order

PAYMENT_APP?=payment
PAYMENT_DOCKERHUB?=ilyashikhaleev/arch-course-payment
PAYMENT_PORT?=8000
PAYMENT_RELEASE?=0.0.1
PAYMENT_HELM_RELEASE_NAME?=payment

all: build

.PHONY: clean
clean:
	rm -f ./bin/${USER_APP} ; \
	rm -f ./bin/${CART_APP} ; \
	rm -f ./bin/${ORDER_APP} ; \
	rm -f ./bin/${PAYMENT_APP} ; \
	rm -f ./bin/${POPULAR_APP} ; \
	rm -f ./bin/${PRODUCT_APP}

.PHONY: build
build: clean build-user build-product build-cart build-order build-payment build-popular

.PHONY: build-user
build-user: clean
	docker build -t $(USER_DOCKERHUB):$(USER_RELEASE) -f User.Dockerfile .

.PHONY: build-product
build-product: clean
	docker build -t $(PRODUCT_DOCKERHUB):$(PRODUCT_RELEASE) -f Product.Dockerfile .

.PHONY: build-popular
build-popular: clean
	docker build -t $(POPULAR_DOCKERHUB):$(POPULAR_RELEASE) -f Popular.Dockerfile .

.PHONY: build-cart
build-cart: clean
	docker build -t $(CART_DOCKERHUB):$(CART_RELEASE) -f Cart.Dockerfile .

.PHONY: build-order
build-order: clean
	docker build -t $(ORDER_DOCKERHUB):$(ORDER_RELEASE) -f Order.Dockerfile .

.PHONY: build-payment
build-payment: clean
	docker build -t $(PAYMENT_DOCKERHUB):$(PAYMENT_RELEASE) -f Payment.Dockerfile .

# helm
.PHONY: start
start: update-helm-dependency-user run-user update-helm-dependency-product run-product update-helm-dependency-cart run-cart update-helm-dependency-order run-order update-helm-dependency-payment run-payment update-helm-dependency-popular run-popular

.PHONY: run
run: run-user run-product run-cart run-order run-payment run-popular

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

.PHONY: run-popular
run-popular:
	helm uninstall $(POPULAR_HELM_RELEASE_NAME) ; \
	helm install $(POPULAR_HELM_RELEASE_NAME) ./helm/popular-chart

.PHONY: update-helm-dependency-popular
update-helm-dependency-popular:
	helm dependency update ./helm/popular-chart

.PHONY: run-cart
run-cart:
	helm uninstall $(CART_HELM_RELEASE_NAME) ; \
	helm install $(CART_HELM_RELEASE_NAME) ./helm/cart-chart

.PHONY: update-helm-dependency-cart
update-helm-dependency-cart:
	helm dependency update ./helm/cart-chart

.PHONY: run-order
run-order:
	helm uninstall $(ORDER_HELM_RELEASE_NAME) ; \
	helm install $(ORDER_HELM_RELEASE_NAME) ./helm/order-chart

.PHONY: update-helm-dependency-order
update-helm-dependency-order:
	helm dependency update ./helm/order-chart

.PHONY: run-payment
run-payment:
	helm uninstall $(PAYMENT_HELM_RELEASE_NAME) ; \
	helm install $(PAYMENT_HELM_RELEASE_NAME) ./helm/payment-chart

.PHONY: update-helm-dependency-payment
update-helm-dependency-payment:
	helm dependency update ./helm/payment-chart

# stresstest
.PHONY: run-stresstest
run-stresstest:
	kubectl apply -f ./helm/stresstest.yaml

.PHONY: stop-stresstest
stop-stresstest:
	kubectl delete -f ./helm/stresstest.yaml
