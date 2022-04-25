# Licensed to Apache Software Foundation (ASF) under one or more contributor
# license agreements. See the NOTICE file distributed with
# this work for additional information regarding copyright
# ownership. Apache Software Foundation (ASF) licenses this file to you under
# the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

include ../../../scripts/base.mk

all: build

.PHONY: build
build: PUSH_OR_LOAD = --load
build: PLATFORMS =

.PHONY: release
release: PUSH_OR_LOAD = --push
release: PLATFORMS = --platform linux/386,linux/amd64,linux/arm64

build release:
	docker buildx create --use --driver docker-container --name skywalking_k8s_event_exporter > /dev/null 2>&1 || true
	cd ../../.. ; \
	docker buildx build $(PUSH_OR_LOAD) $(PLATFORMS) --build-arg VERSION=$(VERSION) . -f build/package/docker/Dockerfile -t $(HUB)/$(APP):$(VERSION) -t $(HUB)/$(APP):latest
	docker buildx rm skywalking_k8s_event_exporter

.PHONY: push
push: build
	docker push $(HUB)/$(APP):$(VERSION)
