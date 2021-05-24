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
	"context"
	"text/template"

	v1 "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
)

type Exporter interface {
	Name() string
	Init(ctx context.Context) error
	Export(ctx context.Context, events chan *v1.Event)
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

type SourceTemplate struct {
	serviceTemplate         *template.Template
	serviceInstanceTemplate *template.Template
	endpointTemplate        *template.Template
}

type EventTemplate struct {
	Source          event.Source `mapstructure:"source"`
	sourceTemplate  SourceTemplate
	Message         string `mapstructure:"message"`
	messageTemplate *template.Template
}

func (tmplt *EventTemplate) Init() (err error) {
	if tmplt.Message != "" {
		if tmplt.messageTemplate, err = template.New("EventMessageTemplate").Parse(tmplt.Message); err != nil {
			return err
		}
	}

	if t := tmplt.Source.Service; t != "" {
		if tmplt.sourceTemplate.serviceTemplate, err = template.New("EventSourceServiceTemplate").Parse(t); err != nil {
			return err
		}
	}
	if t := tmplt.Source.ServiceInstance; t != "" {
		if tmplt.sourceTemplate.serviceInstanceTemplate, err = template.New("EventServiceInstanceTemplate").Parse(t); err != nil {
			return err
		}
	}
	if t := tmplt.Source.Endpoint; t != "" {
		if tmplt.sourceTemplate.endpointTemplate, err = template.New("EventEndpointTemplate").Parse(t); err != nil {
			return err
		}
	}

	return err
}
