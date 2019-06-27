FROM golang:1.12 as build
COPY . /go/src/docker-scan
WORKDIR /go/src/docker-scan
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN make build

FROM scratch
COPY --from=build /go/src/docker-scan/docker-scan /docker-scan
