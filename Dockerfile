FROM golang:1.24-alpine3.21 AS builder

WORKDIR /application

COPY go.sum .
COPY go.mod .

RUN ["go", "mod", "download"]

COPY . .

RUN ["go", "build", "-o", "/myapp", "./cmd/app/main.go"]

FROM alpine AS certificates
RUN apk --no-cache add ca-certificates

FROM scratch

COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /myapp /myapp

ENTRYPOINT ["/myapp"]