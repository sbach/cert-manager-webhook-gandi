TOP = $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)

IMAGE_NAME = webhook
IMAGE_TAG = latest

KUBERNETES_VERSION = 1.29.0
KUBEBUILDER_URL = https://go.kubebuilder.io/test-tools/$(KUBERNETES_VERSION)/$(OS)/$(ARCH)
KUBEBUILDER_DIR = $(TOP)/kubebuilder

export TEST_ASSET_ETCD=$(KUBEBUILDER_DIR)/etcd
export TEST_ASSET_KUBECTL=$(KUBEBUILDER_DIR)/kubectl
export TEST_ASSET_KUBE_APISERVER=$(KUBEBUILDER_DIR)/kube-apiserver

TEST_DATA_DIR = ./testdata/

TEST_YAML_SAMPLE_FILES := $(shell find $(TEST_DATA_DIR) -name '*.yaml.sample')
TEST_YAML_GENERATED_FILES := $(patsubst %.yaml.sample,%.generated.yaml,$(TEST_YAML_SAMPLE_FILES))

build:
	docker build --tag "$(IMAGE_NAME):$(IMAGE_TAG)" .

$(KUBEBUILDER_DIR):
	mkdir -p $(KUBEBUILDER_DIR)
	curl -fsSL $(KUBEBUILDER_URL) | \
	    tar \
	        --extract \
	        --gunzip \
	        --strip-components 2 \
	        --directory $(KUBEBUILDER_DIR)

%.generated.yaml: %.yaml.sample
	envsubst < $< > $@

test: $(KUBEBUILDER_DIR) $(TEST_YAML_GENERATED_FILES)
	go test -v -run $(TESTS) .

clean-test-data:
	shred -zu $(TEST_YAML_GENERATED_FILES) | true

clean: clean-test-data
	rm -rf $(KUBEBUILDER_DIR)
