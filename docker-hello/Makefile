all: build

build:
	@go build -v .

build-static:
	@go build -v -a -tags netgo -ldflags '-w -extldflags "-static"' .

image:
	@buildctl build --frontend=dockerfile.v0 --local context=. --local dockerfile=. --progress plain --exporter=local --exporter-opt output=.
