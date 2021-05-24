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

package pipe

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	exp "github.com/apache/skywalking-kubernetes-event-exporter/pkg/exporter"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/k8s"
)

type workflow struct {
	filter   *configs.FilterConfig
	exporter exp.Exporter
	events   chan *v1.Event
}

type Pipe struct {
	Watcher   *k8s.EventWatcher
	workflows []workflow
}

func (p *Pipe) Init(ctx context.Context) error {
	logger.Log.Debugf("initializing pipe")

	p.workflows = []workflow{}

	initialized := map[string]bool{}
	for _, filter := range configs.GlobalConfig.Filters {
		filter.Init()

		for _, name := range filter.Exporters {
			if _, ok := configs.GlobalConfig.Exporters[name]; !ok {
				return fmt.Errorf("exporter %v is not defined", filter.Exporters)
			}
			exporter := exp.GetExporter(name)
			if exporter == nil {
				return fmt.Errorf("exporter %v is not defined", filter.Exporters)
			}
			if initialized[name] {
				logger.Log.Debugf("exporter %+v has been initialized, skip", name)
				continue
			}
			if err := exporter.Init(ctx); err != nil {
				return err
			}
			initialized[name] = true

			events := make(chan *v1.Event)

			p.workflows = append(p.workflows, workflow{
				filter:   filter,
				exporter: exporter,
				events:   events,
			})
		}
	}

	if err := k8s.Registry.Init(); err != nil {
		return err
	}

	logger.Log.Debugf("pipe has been initialized")

	return nil
}

func (p *Pipe) Start(ctx context.Context) error {
	p.Watcher.Start(ctx)

	k8s.Registry.Start(ctx)

	for _, wkfl := range p.workflows {
		go wkfl.exporter.Export(ctx, wkfl.events)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("stopping pipe")
			return nil
		case e := <-p.Watcher.Events:
			for _, wkfl := range p.workflows {
				go func(w workflow) {
					fCtx, cancel := context.WithTimeout(ctx, time.Minute)
					defer cancel()

					if !w.filter.Filter(fCtx, e) {
						w.events <- e
					}
				}(wkfl)
			}
		}
	}
}
