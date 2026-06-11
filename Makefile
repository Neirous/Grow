APP_NAME    := grow
IMAGE_NAME  := grow
REGISTRY    := crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com
NAMESPACE   := aliyun3175536781
TAG         ?= latest
FULL_IMAGE  := $(REGISTRY)/$(NAMESPACE)/$(IMAGE_NAME):$(TAG)
PORT        ?= 8080

.PHONY: run
run:
	go run main.go --port=$(PORT)

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(APP_NAME) .

.PHONY: docker
docker:
	docker build -t $(FULL_IMAGE) .

.PHONY: docker-push
docker-push: docker
	docker push $(FULL_IMAGE)

.PHONY: docker-run
docker-run:
	docker run -d --name $(APP_NAME) -p $(PORT):8080 -v $$PWD/data:/data -e TZ=Asia/Shanghai $(FULL_IMAGE)

.PHONY: docker-stop
docker-stop:
	docker rm -f $(APP_NAME) 2>/dev/null || true

.PHONY: docker-login
docker-login:
	sudo docker login --username=aliyun3175536781 $(REGISTRY)

.PHONY: clean
clean:
	rm -f $(APP_NAME)
	docker rm -f $(APP_NAME) 2>/dev/null || true
