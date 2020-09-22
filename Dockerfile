# Build the manager binary
FROM golang:1.13-alpine as builder

# Copy in the go src
WORKDIR /go/src/github.com/kyma-incubator/octopus
COPY pkg/    pkg/
COPY cmd/    cmd/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-incubator/octopus/cmd/manager

# Copy the controller-manager into a thin image
FROM scratch
WORKDIR /

COPY --from=builder /go/src/github.com/kyma-incubator/octopus/manager .

ENTRYPOINT ["/manager"]
