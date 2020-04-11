APP?=arch-course
DOCKERHUB?=ilyashikhaleev/arch-course
PORT?=8000
RELEASE?=0.0.3

all: build

.PHONY: clean
clean:
	rm -f ./bin/${APP}

.PHONY: build
build: clean
	docker build -t $(DOCKERHUB):$(RELEASE) .

.PHONY: run
run:
	docker stop $(APP) || true && docker rm $(APP) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --rm \
		-e "PORT=${PORT}" \
		$(DOCKERHUB):$(RELEASE)
