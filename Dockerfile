# Build the manager binary
FROM golang:1.11.4-alpine3.8 as builder

# Copy in the go src
WORKDIR /go/src/github.com/kyma-incubator/octopus
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-incubator/octopus/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.8
WORKDIR /
COPY --from=builder /go/src/github.com/kyma-incubator/octopus/manager .
ENTRYPOINT ["/manager"]
