FROM golang:1.12 as build
COPY . /go/src/docker-hello
WORKDIR /go/src/docker-hello
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN make build

FROM scratch
COPY --from=build /go/src/docker-hello/docker-hello /docker-hello
