FROM golang:1.18.3-alpine3.16 AS builder
RUN mkdir /go/src
WORKDIR /go/src
ENV GO111MODULE=on CGO_ENABLED=0
# download dependency
COPY go.mod go.sum deployment.go /go/src
RUN  go mod download
RUN go install honnef.co/go/tools/cmd/staticcheck@v0.1.2
COPY . .
RUN go vet ./... && staticcheck ./... && go test ./... && go build -o deployment deployment.go

FROM alpine:latest AS release
EXPOSE 80
CMD [./deployment]