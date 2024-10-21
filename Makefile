# Nicked from node_exporter repo and modified for current repo needs

# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOOPTS       ?=
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)

PROMU        := $(FIRST_GOPATH)/bin/promu

ifeq (arm, $(GOHOSTARCH))
	GOHOSTARM ?= $(shell GOARM= $(GO) env GOARM)
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)v$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
endif

PROMU_VERSION ?= 0.17.0
PROMU_URL     := https://github.com/prometheus/promu/releases/download/v$(PROMU_VERSION)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM).tar.gz

PROMTOOL_VERSION ?= 2.50.0
PROMTOOL_URL     ?= https://github.com/prometheus/prometheus/releases/download/v$(PROMTOOL_VERSION)/prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM).tar.gz
PROMTOOL         ?= $(FIRST_GOPATH)/bin/promtool

PREFIX           := $(shell pwd)/bin

TEST_DOCKER             ?= false
DOCKER_IMAGE_NAME       ?= ceems
MACH                    ?= $(shell uname -m)

STATICCHECK_IGNORE =

PROMU_CONF ?= .promu.yml
pkgs := ./pkg/helpers ./cmd/fake_energy_counters

PROMU := $(FIRST_GOPATH)/bin/promu --config $(PROMU_CONF)

all:: build

.PHONY: build
build: promu
	@echo ">> building binaries"
	$(PROMU) build --prefix $(PREFIX) $(PROMU_BINARIES)

.PHONY: promu
promu: $(PROMU)
$(PROMU):
	$(eval PROMU_TMP := $(shell mktemp -d))
	curl -s -L $(PROMU_URL) | tar -xvzf - -C $(PROMU_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu
	rm -r $(PROMU_TMP)

.PHONY: promtool
promtool: $(PROMTOOL)
$(PROMTOOL):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -fsS -L $(PROMTOOL_URL) | tar -xvzf - -C $(FIRST_GOPATH)/bin --strip 1 "prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM)/promtool" 
