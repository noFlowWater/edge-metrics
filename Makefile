.PHONY: test-server test-server-cover test-exporter test-helm lint-helm kubeconform test-all \
       build-exporter build-server build-all deploy

# === Build & Deploy ===
REGISTRY    ?= daclab
BUILDER     ?= edge-builder
PLATFORMS   ?= linux/amd64,linux/arm64
CACHE_REPO  ?= $(REGISTRY)/buildcache

EXPORTER_TAG ?= $(shell python3 -c "import yaml; d=yaml.safe_load(open('charts/edge-metrics/values-cluster.yaml')); print(d['exporter']['image']['tag'])" 2>/dev/null || echo latest)
SERVER_TAG   ?= $(shell python3 -c "import yaml; d=yaml.safe_load(open('charts/edge-metrics/values-cluster.yaml')); print(d['server']['image']['tag'])" 2>/dev/null || echo latest)

build-exporter:
	docker buildx build --builder $(BUILDER) --platform $(PLATFORMS) \
		--cache-from type=registry,ref=$(CACHE_REPO):exporter \
		--cache-to type=registry,ref=$(CACHE_REPO):exporter,mode=max \
		-t $(REGISTRY)/edge-metrics-exporter:$(EXPORTER_TAG) --push ./exporter

build-server:
	docker buildx build --builder $(BUILDER) --platform $(PLATFORMS) \
		--cache-from type=registry,ref=$(CACHE_REPO):server \
		--cache-to type=registry,ref=$(CACHE_REPO):server,mode=max \
		-t $(REGISTRY)/edge-metrics-server:$(SERVER_TAG) --push ./server

build-all: build-exporter build-server

deploy:
	helm upgrade --install edge-metrics charts/edge-metrics \
		-n monitoring $(if $(wildcard charts/edge-metrics/values-cluster.yaml),-f charts/edge-metrics/values-cluster.yaml)

# Go server tests
test-server:
	cd server && go test ./...

test-server-cover:
	cd server && go test ./... -coverprofile=cover.out && go tool cover -func=cover.out

# Python exporter tests
test-exporter:
	cd exporter && python3 -m pytest tests/ -v

# Helm chart tests
lint-helm:
	helm lint charts/edge-metrics
	helm template test charts/edge-metrics > /dev/null

test-helm:
	helm unittest charts/edge-metrics/charts/server
	helm unittest charts/edge-metrics/charts/exporter

kubeconform:
	helm template test charts/edge-metrics | kubeconform -strict -ignore-missing-schemas -summary
	helm template test charts/edge-metrics -f charts/edge-metrics/values-dev.yaml | kubeconform -strict -ignore-missing-schemas -summary
	helm template test charts/edge-metrics -f charts/edge-metrics/values-prod.yaml | kubeconform -strict -ignore-missing-schemas -summary

# Run all tests
test-all: test-server test-exporter lint-helm test-helm kubeconform
