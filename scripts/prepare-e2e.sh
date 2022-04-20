#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

OS=$(go env GOOS)
ARCH=$(go env GOHOSTARCH)

# prepare base dir
TMP_DIR=/tmp/sw-event-exporter-e2e
BIN_DIR=$TMP_DIR/bin
mkdir -p $TMP_DIR $BIN_DIR && cd $TMP_DIR

export PATH=$BIN_DIR:$PATH

KUBECTL_VERSION=${KUBECTL_VERSION:-'v1.19.1'}
SWCTL_VERSION=${SWCTL_VERSION:-'0.10.0'}
HELM_VERSION=${HELM_VERSION:-'helm-v3.0.0'}

prepare_ok=true

function error_check() {
    if [ $? -ne 0 ]; then
        echo "[ERROR] Failed to install $1, please check"
        prepare_ok=false
    fi
}

function install_kubectl()
{
    if ! command -v kubectl &> /dev/null; then
      echo "Installing kubectl"
      mkdir -p $TMP_DIR/kubectl && cd $TMP_DIR/kubectl
      curl -LO https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl && chmod +x ./kubectl && mv ./kubectl ${BIN_DIR}
      error_check "kubectl"
    fi
}

function install_swctl()
{
    if ! command -v swctl &> /dev/null; then
      echo "Installing swctl"
      mkdir -p $TMP_DIR/swctl && cd $TMP_DIR/swctl
      curl -kLo skywalking-cli.tar.gz https://github.com/apache/skywalking-cli/archive/${SWCTL_VERSION}.tar.gz
      tar -zxf skywalking-cli.tar.gz --strip=1
      VERSION=${SWCTL_VERSION} make install DESTDIR=${BIN_DIR}
      error_check "swctl"
    fi
}

function install_yq()
{
    if ! command -v yq &> /dev/null; then
      echo "Installing yq"
      mkdir -p $TMP_DIR/yq && cd $TMP_DIR/yq
      wget https://github.com/mikefarah/yq/releases/download/v4.11.1/yq_${OS}_${ARCH}.tar.gz -O - |\
      tar xz && mv yq_${OS}_${ARCH} ${BIN_DIR}/yq
      error_check "yq"
    fi
}

function install_helm() {
  if ! command -v helm &> /dev/null; then
    echo "Installing helm"
    mkdir -p $TMP_DIR/helm && cd $TMP_DIR/helm
    curl -sSL https://get.helm.sh/${HELM_VERSION}-${OS}-${ARCH}.tar.gz | tar xz -C $BIN_DIR --strip-components=1 ${OS}-${ARCH}/helm
    mv ${OS}-${ARCH}/helm ${BIN_DIR}/helm
    error_check "helm"
  fi
}

function install_all()
{
    install_kubectl
    install_swctl
    install_yq
    install_helm

    if [ "$prepare_ok" = false ]; then
        echo "Install e2e dependencies failed"
        exit 1
    else
        echo "Install e2e dependencies successfully"
        exit 0
    fi
}

install_all
