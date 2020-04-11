APP?=arch-course
PORT?=8000
RELEASE?=0.0.1

all: build

.PHONY: clean
clean:
	rm -f ./bin/${APP}

.PHONY: build
build: clean
	docker build -t $(APP):$(RELEASE) .

.PHONY: run
run:
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --rm \
		-e "PORT=${PORT}" \
		$(APP):$(RELEASE)
