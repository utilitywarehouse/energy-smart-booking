DOCKER_REGISTRY?=registry.uw.systems
DOCKER_REPOSITORY_NAMESPACE?=energy-smart
DOCKER_REPOSITORY_IMAGE=$(SERVICE)
DOCKER_REPOSITORY=$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY_NAMESPACE)/$(DOCKER_REPOSITORY_IMAGE)

BUILDENV :=
BUILDENV += CGO_ENABLED=0
GIT_HASH := $(GITHUB_SHA)
ifeq ($(GIT_HASH),)
  GIT_HASH := $(shell git rev-parse HEAD)
endif
LINKFLAGS :=-s -X main.gitHash=$(GIT_HASH) -extldflags "-static"
TESTFLAGS := -v -cover -tags testing
LINTER := golangci-lint
GOTEST := gotestsum --

BRANCH_NAME := $(shell echo $(GITHUB_REF_NAME) | sed -e 's/[^a-zA-Z0-9]/-/g')

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
join-with = $(subst $(SPACE),$1,$(strip $2))

LEXC :=

.PHONY: install
install:
	GOPROXY=https://proxy.golang.org GO111MODULE=on GOPRIVATE="github.com/utilitywarehouse/*" go mod download
	go install github.com/golang/mock/mockgen@v1.6.0
	go install gotest.tools/gotestsum@latest

$(LINTER):
	@ [ -e ./bin/$(LINTER) ] || wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

.PHONY: lint
lint: $(LINTER)
	./bin/$(LINTER) run ./${SOURCE_FILES}/... $(LINT_FLAGS)

.PHONY: clean
clean:
	rm -f $(SERVICE)

# builds our binary
$(SERVICE): clean
	GO111MODULE=on $(BUILDENV) go build -o $(SERVICE) -a -ldflags '$(LINKFLAGS)' $(SOURCE_FILES)

build: $(SERVICE)

.PHONY: test
test:
	GO111MODULE=on $(BUILDENV) $(GOTEST) $(TESTFLAGS) ./${SOURCE_FILES}/...

.PHONY: all
all: clean $(LINTER) lint test build

docker-image:
	docker build -t $(DOCKER_REPOSITORY):local . --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) --build-arg SOURCE_FILES=$(SOURCE_FILES)

ci-docker-auth:
	@echo "Logging in to $(DOCKER_REGISTRY) as $(DOCKER_ID)"
	@docker login -u $(DOCKER_ID) -p $(DOCKER_PASSWORD) $(DOCKER_REGISTRY)


ci-docker-build: ci-docker-auth
	docker build -t $(DOCKER_REPOSITORY):$(GITHUB_SHA) . --build-arg SERVICE=$(SERVICE) --build-arg SOURCE_FILES=$(SOURCE_FILES) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)
	if [ "$(GITHUB_REF_NAME)" = "$(DEFAULT_BRANCH_NAME)" ]; then docker tag $(DOCKER_REPOSITORY):$(GITHUB_SHA) $(DOCKER_REPOSITORY):latest; fi
	if [ "$(GITHUB_REF_NAME)" != "$(DEFAULT_BRANCH_NAME)" ]; then docker tag $(DOCKER_REPOSITORY):$(GITHUB_SHA) $(DOCKER_REPOSITORY):$(BRANCH_NAME); fi
	docker push -a $(DOCKER_REPOSITORY)
