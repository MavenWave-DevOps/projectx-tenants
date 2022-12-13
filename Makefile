PWD := ${CURDIR}
ADDITIONAL_BUILD_ARGUMENTS?=""

PKG	:= metacontroller
API_GROUPS := metacontroller/v1alpha1

CLUSTER := projectx-tenants

CODE_GENERATOR_VERSION="v0.24.3"

export KUBECONFIG=${PWD}/kubeconfig

all: generate_crds

.PHONY: generate_crds local-dev
generate_crds:
	@echo "+ Generating crds"
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	@controller-gen +crd:generateEmbeddedObjectMeta=true +paths="./api/..." +output:crd:stdout > manifests/v1/crdv1.yaml

create_cluster:
	@go install sigs.k8s.io/kind@v0.17.0
	@kind create cluster --name ${CLUSTER} --kubeconfig ${PWD}/kubeconfig || true

local_dev: create_cluster
	@echo "+ Setup dev environment"
	@kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production

install: generate_crds
	@kubectl apply -k ${PWD}/manifests/

run:
	@docker build --rm -t tenant-controller .
	@kind load docker-image tenant-controller:latest -n ${CLUSTER}
	@kubectl -n metacontroller delete pod --selector=app=tenant-controller

destroy:
	@kind delete cluster --name ${CLUSTER}
