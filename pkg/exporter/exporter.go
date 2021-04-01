/*
 * Licensed to Apache Software Foundation (ASF) under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Apache Software Foundation (ASF) licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package exporter

import (
	v1 "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
)

type Exporter interface {
	Name() string
	Init() error
	Export(events chan *v1.Event)
	Stop()
}

type MessageTemplate struct {
	Source  *event.Source `mapstructure:"source"`
	Message string        `mapstructure:"message"`
}

var exporters = map[string]Exporter{}

func RegisterExporter(name string, exporter Exporter) {
	if _, ok := exporters[name]; ok {
		logger.Log.Panicf("exporter with name %v has already existed", name)
	}

	exporters[name] = exporter
}

func GetExporter(name string) Exporter {
	return exporters[name]
}
