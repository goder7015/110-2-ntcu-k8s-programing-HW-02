FROM golang:1.18.3-alpine3.16 AS builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN  go mod download
COPY /cmd /cmd
RUN go install honnef.co/go/tools/cmd/staticcheck@v0.1.2
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main ./cmd/incluster

FROM alpine:latest AS release
WORKDIR /
COPY --from=builder /go/main .
USER 65532:65532
EXPOSE 80

ENTRYPOINT ["/main"]