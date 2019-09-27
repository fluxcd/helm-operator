.DEFAULT: all
.PHONY: all clean realclean test integration-test check-generated

SUDO := $(shell docker info > /dev/null 2> /dev/null || echo "sudo")

TEST_FLAGS?=

include docker/kubectl.version
include docker/helm.version

# NB default target architecture is amd64. If you would like to try the
# other one -- pass an ARCH variable, e.g.,
#  `make ARCH=arm64`
ifeq ($(ARCH),)
	ARCH=amd64
endif
CURRENT_OS_ARCH=$(shell echo `go env GOOS`-`go env GOARCH`)
GOBIN?=$(shell echo `go env GOPATH`/bin)

godeps=$(shell go list -deps -f '{{if not .Standard}}{{ $$dep := . }}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $(1) | sed "s%${PWD}/%%g")

HELM_OPERATOR_DEPS:=$(call godeps,./cmd/helm-operator/...)

IMAGE_TAG:=$(shell ./docker/image-tag)
VCS_REF:=$(shell git rev-parse HEAD)
BUILD_DATE:=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

DOCS_PORT:=8000

all: $(GOBIN)/bin/helm-operator build/.helm-operator.done

clean:
	go clean ./cmd/helm-operator
	rm -rf ./build
	rm -f test/bin/kubectl test/bin/helm test/bin/kind

realclean: clean
	rm -rf ./cache

test: test/bin/helm
	PATH="${PWD}/bin:${PWD}/test/bin:${PATH}" GO111MODULES=on go test ${TEST_FLAGS} $(shell go list ./... | grep -v "^github.com/weaveworks/flux/vendor" | sort -u)

e2e: test/bin/helm test/bin/kubectl build/.helm-operator.done
	PATH="${PWD}/test/bin:${PATH}" CURRENT_OS_ARCH=$(CURRENT_OS_ARCH) test/e2e/run.sh

build/.%.done: docker/Dockerfile.%
	mkdir -p ./build/docker/$*
	cp $^ ./build/docker/$*/
	$(SUDO) docker build -t docker.io/fluxcd/$* -t docker.io/fluxcd/$*:$(IMAGE_TAG) \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-f build/docker/$*/Dockerfile.$* ./build/docker/$*
	touch $@

build/.helm-operator.done: build/helm-operator build/kubectl build/helm docker/ssh_config docker/known_hosts.sh docker/helm-repositories.yaml

build/helm-operator: $(HELM_OPERATOR_DEPS)
build/helm-operator: cmd/helm-operator/*.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o $@ $(LDFLAGS) -ldflags "-X main.version=$(shell ./docker/image-tag)" ./cmd/helm-operator

build/kubectl: cache/linux-$(ARCH)/kubectl-$(KUBECTL_VERSION)
test/bin/kubectl: cache/$(CURRENT_OS_ARCH)/kubectl-$(KUBECTL_VERSION)
build/helm: cache/linux-$(ARCH)/helm-$(HELM_VERSION)
test/bin/helm: cache/$(CURRENT_OS_ARCH)/helm-$(HELM_VERSION)
build/kubectl test/bin/kubectl build/helm test/bin/helm:
	mkdir -p build
	cp $< $@
	if [ `basename $@` = "build" -a $(CURRENT_OS_ARCH) = "linux-$(ARCH)" ]; then strip $@; fi
	chmod a+x $@

cache/%/kubectl-$(KUBECTL_VERSION): docker/kubectl.version
	mkdir -p cache/$*
	curl --fail -L -o cache/$*/kubectl-$(KUBECTL_VERSION).tar.gz "https://dl.k8s.io/$(KUBECTL_VERSION)/kubernetes-client-$*.tar.gz"
	[ $* != "linux-$(ARCH)" ] || echo "$(KUBECTL_CHECKSUM_$(ARCH))  cache/$*/kubectl-$(KUBECTL_VERSION).tar.gz" | shasum -a 512 -c
	tar -m --strip-components 3 -C ./cache/$* -xzf cache/$*/kubectl-$(KUBECTL_VERSION).tar.gz kubernetes/client/bin/kubectl
	mv ./cache/$*/kubectl $@

cache/%/helm-$(HELM_VERSION): docker/helm.version
	mkdir -p cache/$*
	curl --fail -L -o cache/$*/helm-$(HELM_VERSION).tar.gz "https://storage.googleapis.com/kubernetes-helm/helm-v$(HELM_VERSION)-$*.tar.gz"
	[ $* != "linux-$(ARCH)" ] || echo "$(HELM_CHECKSUM_$(ARCH))  cache/$*/helm-$(HELM_VERSION).tar.gz" | shasum -a 256 -c
	tar -m -C ./cache -xzf cache/$*/helm-$(HELM_VERSION).tar.gz $*/helm
	mv cache/$*/helm $@

$(GOBIN)/bin/helm-operator: $(HELM_OPERATOR_DEPS)
	GO111MODULE=on go install ./cmd/helm-operator

pkg/install/generated_templates.gogen.go: pkg/install/templates/*
	cd pkg/install && go run generate.go embedded-templates

generate-deploy: pkg/install/generated_templates.gogen.go
	cd deploy && go run ../pkg/install/generate.go deploy

check-generated: generate-deploy pkg/install/generated_templates.gogen.go
	git diff --exit-code -- pkg/install/generated_templates.gogen.go
	./hack/update/verify.sh

build-docs:
	@cd docs && docker build -t flux-docs .

test-docs: build-docs
	@docker run -it flux-docs /usr/bin/linkchecker _build/html/index.html

serve-docs: build-docs
	@echo Stating docs website on http://localhost:${DOCS_PORT}/_build/html/index.html
	@docker run -i -p ${DOCS_PORT}:8000 -e USER_ID=$$UID flux-docs
