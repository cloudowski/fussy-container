SHELL := /bin/bash
NAME = cloudowski/fussy-container

VERSION?=$(shell git tag -l --points-at HEAD)
#VERSION?=latest

all: clean build

.PHONY: build tinybuild run runfg clean push
default: build

build: 
	docker build -t $(NAME):$(VERSION) .

tinybuild: 
	docker build -f Dockerfile.slim -t $(NAME):$(VERSION) .

push: tinybuild
	docker push $(NAME):$(VERSION)

run: 
	docker run -p 8080:8080 --name=$(NAME) -d $(NAME):$(VERSION)

runfg: 
	docker run --rm -p 8080:8080 -ti $(NAME):$(VERSION)

clean:
	-docker rm -f $(NAME)
	-docker rmi $(NAME):$(VERSION)
