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

HUB ?= docker.io/apache
APP = skywalking-kubernetes-event-exporter
VERSION ?= $(shell git rev-parse --short HEAD)
OUT_DIR = bin
ARCH := $(shell uname)
OSNAME := $(if $(findstring Darwin,$(ARCH)),darwin,linux)

GO := GO111MODULE=on go
GO_PATH = $(shell $(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_BUILD_LDFLAGS = -X main.version=$(VERSION)

PLATFORMS := windows linux darwin
os = $(word 1, $@)
ARCH = amd64

RELEASE_BIN = $(APP)-$(VERSION)-bin
RELEASE_SRC = $(APP)-$(VERSION)-src
