.DEFAULT_GOAL := validate

# Image URL to use all building/pushing image targets
APP_NAME ?= octopus
IMG ?= $(APP_NAME):latest
IMG-CI = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME):$(DOCKER_TAG)

# Run tests
.PHONY: test
test: generate manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs and samples into a cluster
.PHONY: install
install: manifests
	kubectl apply -f config/crds
	kubectl apply -f config/samples

.PHONY: uninstall
uninstall:
	kubectl delete pods -l testing.kyma-project.io/created-by-octopus=true
	kubectl delete -f config/samples


# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests:
go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd rbac:roleName=manager-role webhook paths="./apis/..."

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
.PHONY: docker-build
docker-build: resolve vendor-create generate validate
	docker build . -t ${IMG}
	docker tag ${IMG} ${IMG-CI}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG-CI}"'@' ./config/default/manager_image_patch.yaml
	rm -rf vendor/

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push ${IMG-CI}

### Custom targets
# Resolve dependencies
.PHONY: resolve
resolve:
	go mod tidy

# Executes the whole validation
.PHONY: validate
validate: fmt vet test
	go mod verify

.PHONY: vendor-create
vendor-create:
	go mod vendor

# CI specified targets
.PHONY: ci-pr
ci-pr: docker-build docker-push

.PHONY: ci-master
ci-master: docker-build docker-push

.PHONY: ci-release
ci-release: docker-build docker-push
