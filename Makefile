MAKE_REL_PATH:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

GOLANG ?= go
GOARCH ?= arm64
GO_BUILD_FLAGS ?= -ldflags="-s -w"
GO_LAMBDA_TAGS ?= -tags "lambda.norpc"
GO_BUILD ?= CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} ${GOLANG} build ${GO_BUILD_FLAGS}
GO_BUILD_FILES := $(shell find . -type f -not -path "./.git/*" -not -path "./bin/*" -not -name "*_test.go")

bin/%/bootstrap: ${GO_BUILD_FILES}
	cd cmd/$(patsubst bin/%/bootstrap,%,$@) && ${GO_BUILD} ${GO_LAMBDA_TAGS} -o ${MAKE_REL_PATH}/$@

.PHONY: test
test:
	${GOLANG} test -count=1 ./...

.PHONY: mocks
mocks:
	mockery --name=KMSAPI --srcpkg=./internal/issuer --output=internal/mocks
	mockery --name=SSMAPI --srcpkg=./internal/issuer --output=internal/mocks

.PHONY: build
build: bin/key_generator/bootstrap
build: bin/key_info_loader/bootstrap
build: bin/jwt_issuer_kms/bootstrap
build: bin/jwt_issuer_parameter_store/bootstrap
