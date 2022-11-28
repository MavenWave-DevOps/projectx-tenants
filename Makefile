PWD := ${CURDIR}
ADDITIONAL_BUILD_ARGUMENTS?=""

PKG		:= metacontroller
API_GROUPS := metacontroller/v1alpha1

CODE_GENERATOR_VERSION="v0.24.3"

export KUBECONFIG=${PWD}/kubeconfig

all: generate_crds

.PHONY: generate_crds local-dev
generate_crds:
	@echo "+ Generating crds"
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	@controller-gen +crd:generateEmbeddedObjectMeta=true +paths="./api/..." +output:crd:stdout > v1/crdv1.yaml

create_cluster:
	@go install sigs.k8s.io/kind@v0.17.0
	@kind create cluster --name projectx-tenants --kubeconfig ${PWD}/kubeconfig || true

local_dev: create_cluster
	@echo "+ Setup dev environment"
	@kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production

install:
	@kubectl apply -f ${PWD}/v1/
	@kubectl apply -f ${PWD}/manifest/controller/

run:
	@docker build --rm -t tenant-controller .
	@kind load docker-image tenant-controller:latest -n projectx-tenants
	@kubectl apply -f ${PWD}/manifest/deployment/

destroy:
	@kind delete cluster --name projectx-tenants
