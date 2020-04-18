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
	helm uninstall archapp ./arch-chart ; \
	helm install archapp ./arch-chart
