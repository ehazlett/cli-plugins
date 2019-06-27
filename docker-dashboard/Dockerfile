FROM golang:1.12 as build
COPY . /go/src/docker-dashboard
WORKDIR /go/src/docker-dashboard
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN make build

FROM scratch
COPY --from=build /go/src/docker-dashboard/docker-dashboard /docker-dashboard
